package rat

import (
	"fmt"
	"os"
	"strings"

	"github.com/ericfreese/otto"
	_ "github.com/ericfreese/otto/underscore"
)

func InitJs() *otto.Otto {
	vm := otto.New()

	vm.Set("Rat", BuildRat(vm))

	return vm
}

func BuildContext(ctx Context) *otto.Object {
	context, _ := vm.Object("({})")
	for k, v := range ctx {
		context.Set(k, v)
	}

	context.SetPrivate(ctx)

	context.Set("get", func(call otto.FunctionCall) otto.Value {
		key := call.Argument(0).String()
		ctx := call.This.Object().GetPrivate().(Context)

		val, _ := otto.ToValue(ctx[key])

		return val
	})

	context.Set("set", func(call otto.FunctionCall) otto.Value {
		key := call.Argument(0).String()
		val, _ := call.Argument(1).ToString()
		ctx := call.This.Object().GetPrivate().(Context)

		ctx[key] = val

		return otto.NullValue()
	})

	return context
}

func BuildPager(p Pager) *otto.Object {
	pager, _ := vm.Object("({})")

	pager.SetPrivate(p)

	pager.Set("getContext", func(call otto.FunctionCall) otto.Value {
		return BuildContext(call.This.Object().GetPrivate().(Pager).GetContext()).Value()
	})

	pager.Set("getWorkingDir", func(call otto.FunctionCall) otto.Value {
		wd, _ := otto.ToValue(call.This.Object().GetPrivate().(Pager).GetWorkingDir())
		return wd
	})

	pager.Set("addAnnotator", func(call otto.FunctionCall) otto.Value {
		a := call.Argument(0).Object().GetPrivate().(Annotator)
		p := call.This.Object().GetPrivate().(Pager)

		p.AddAnnotator(a)

		return otto.NullValue()
	})

	pager.Set("addEventHandler", func(call otto.FunctionCall) otto.Value {
		keyStr := call.Argument(0).String()
		arg2 := call.Argument(1)

		var (
			requirements string
			callback     otto.Value
			eventHandler EventHandler
		)

		if arg2.IsString() {
			requirements = arg2.String()
			callback = call.Argument(2)
			eventHandler = NewCtxEventHandler(strings.Split(requirements, ","), func(ctx Context) {
				callback.Call(otto.NullValue(), BuildContext(ctx))
			})
		} else {
			callback = arg2
			eventHandler = NewEventHandler(func() {
				callback.Call(otto.NullValue())
			})
		}

		p := call.This.Object().GetPrivate().(Pager)

		p.AddEventHandler(keyStr, eventHandler)

		return otto.NullValue()
	})

	pager.Set("onDestroy", func(call otto.FunctionCall) otto.Value {
		hook := call.Argument(0)
		p := call.This.Object().GetPrivate().(Pager)

		p.AddDestroyHook(func() {
			hook.Call(otto.NullValue())
		})

		return otto.NullValue()
	})

	pager.Set("reload", func(call otto.FunctionCall) otto.Value {
		call.This.Object().GetPrivate().(Pager).Reload()
		return otto.NullValue()
	})

	return pager
}

func BuildRat(vm *otto.Otto) *otto.Object {
	rat, _ := vm.Object("({})")

	rat.Set("Context", func(call otto.FunctionCall) otto.Value {
		call.This.Object().SetPrivate(Context{})
		return otto.NullValue()
	})

	rat.Set("CmdPager", func(call otto.FunctionCall) otto.Value {
		modeNames := call.Argument(0).String()
		command := call.Argument(1).String()
		workingDirVal := call.Argument(2)
		ctxVal := call.Argument(3)

		var (
			workingDir string
			ctx        Context
		)

		if workingDirVal.IsUndefined() {
			workingDir, _ = os.Getwd()
		} else {
			workingDir = workingDirVal.String()
		}

		if ctxVal.IsUndefined() {
			ctx = Context{}
		} else {
			ctx = ctxVal.Object().GetPrivate().(Context)
		}

		call.This.Object().SetPrivate(
			NewCmdPager(
				modeNames,
				command,
				workingDir,
				ctx,
			),
		)

		return otto.NullValue()
	})

	rat.Set("MatchAnnotator", func(call otto.FunctionCall) otto.Value {
		command := call.Argument(0).String()
		workingDir := call.Argument(1).String()
		class := call.Argument(2).String()
		ctx := call.Argument(3).Object().GetPrivate().(Context)

		call.This.Object().SetPrivate(
			NewMatchAnnotator(
				command,
				workingDir,
				class,
				ctx,
			),
		)

		return otto.NullValue()
	})

	rat.Set("ExternalAnnotator", func(call otto.FunctionCall) otto.Value {
		command := call.Argument(0).String()
		workingDir := call.Argument(1).String()
		class := call.Argument(2).String()
		ctx := call.Argument(3).Object().GetPrivate().(Context)

		call.This.Object().SetPrivate(
			NewExternalAnnotator(
				command,
				workingDir,
				class,
				ctx,
			),
		)

		return otto.NullValue()
	})

	rat.Set("pushPager", func(call otto.FunctionCall) otto.Value {
		pager := call.Argument(0).Object().GetPrivate().(Pager)

		PushPager(pager)

		return otto.NullValue()
	})

	rat.Set("addChildPager", func(call otto.FunctionCall) otto.Value {
		parent := call.Argument(0).Object().GetPrivate().(Pager)
		child := call.Argument(1).Object().GetPrivate().(Pager)
		creatingKeys := call.Argument(2).String()

		AddChildPager(parent, child, creatingKeys)

		return otto.NullValue()
	})

	rat.Set("mergeContext", func(call otto.FunctionCall) otto.Value {
		orig := call.Argument(0).Object().GetPrivate().(Context)
		extra := call.Argument(1).Object().GetPrivate().(Context)

		return BuildContext(MergeContext(orig, extra)).Value()
	})

	rat.Set("registerMode", func(call otto.FunctionCall) otto.Value {
		name := call.Argument(0).String()
		decorator := call.Argument(1)

		RegisterMode(name, NewMode(func(p Pager) {
			decorator.Call(otto.NullValue(), BuildPager(p).Value())
		}))

		return otto.NullValue()
	})

	rat.Set("addEventHandler", func(call otto.FunctionCall) otto.Value {
		keyEvent := call.Argument(0).String()
		callback := call.Argument(1)

		AddEventHandler(keyEvent, func() {
			callback.Call(otto.NullValue())
		})

		return otto.NullValue()
	})

	rat.Set("exec", func(call otto.FunctionCall) otto.Value {
		command := call.Argument(0).String()
		ctx := call.Argument(1).Object().GetPrivate().(Context)

		Exec(command, ctx)

		return otto.NullValue()
	})

	rat.Set("confirm", func(call otto.FunctionCall) otto.Value {
		message := call.Argument(0).String()
		callback := call.Argument(1)

		Confirm(message, func() { callback.Call(otto.NullValue()) })

		return otto.NullValue()
	})

	rat.Set("confirmExec", func(call otto.FunctionCall) otto.Value {
		command := call.Argument(0).String()
		ctx := call.Argument(1).Object().GetPrivate().(Context)
		callback := call.Argument(2)

		ConfirmExec(command, ctx, func() { callback.Call(otto.NullValue()) })

		return otto.NullValue()
	})

	rat.Set("prompt", func(call otto.FunctionCall) otto.Value {
		label := call.Argument(0).String()
		callback := call.Argument(1)

		gPrompt.Text(label, func(text string, success bool) {
			t, _ := otto.ToValue(text)
			s, _ := otto.ToValue(success)

			callback.Call(otto.NullValue(), t, s)
		})

		return otto.NullValue()
	})

	rat.Set("log", func(call otto.FunctionCall) otto.Value {
		fmt.Fprintln(os.Stderr, call.Argument(0).String())
		return otto.NullValue()
	})

	rat.Set("actions", BuildActions(vm))

	return rat
}

func BuildActions(vm *otto.Otto) *otto.Object {
	actions, _ := vm.Object("({})")

	actions.Set("moveCursor", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.Last().MoveCursor(int(offset))

		return otto.NullValue()
	})

	actions.Set("moveCursorTo", func(call otto.FunctionCall) otto.Value {
		to, _ := call.Argument(0).ToInteger()

		pagers.Last().MoveCursorTo(int(to))

		return otto.NullValue()
	})

	actions.Set("moveCursorNext", func(call otto.FunctionCall) otto.Value {
		annotationClass := call.Argument(0).String()

		pagers.Last().MoveCursorNext(annotationClass)

		return otto.NullValue()
	})

	actions.Set("moveCursorPrevious", func(call otto.FunctionCall) otto.Value {
		annotationClass := call.Argument(0).String()

		pagers.Last().MoveCursorPrevious(annotationClass)

		return otto.NullValue()
	})

	actions.Set("scroll", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.Last().Scroll(int(offset))

		return otto.NullValue()
	})

	actions.Set("pageUp", func(call otto.FunctionCall) otto.Value {
		pagers.Last().PageUp()

		return otto.NullValue()
	})

	actions.Set("pageDown", func(call otto.FunctionCall) otto.Value {
		pagers.Last().PageDown()

		return otto.NullValue()
	})

	actions.Set("moveParentCursor", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.MoveParentCursor(int(offset))

		return otto.NullValue()
	})

	actions.Set("popPager", func(call otto.FunctionCall) otto.Value {
		PopPager()
		return otto.NullValue()
	})

	actions.Set("quit", func(call otto.FunctionCall) otto.Value {
		Quit()
		return otto.NullValue()
	})

	actions.Set("reload", func(call otto.FunctionCall) otto.Value {
		pagers.Last().Reload()
		return otto.NullValue()
	})

	actions.Set("showPagers", func(call otto.FunctionCall) otto.Value {
		n, _ := call.Argument(0).ToInteger()
		pagers.Show(int(n))
		return otto.NullValue()
	})

	return actions
}
