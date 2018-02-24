package rat

import (
	"strings"

	"github.com/robertkrimen/otto"
	_ "github.com/robertkrimen/otto/underscore"
)

func InitJs() *otto.Otto {
	vm := otto.New()

	vm.Set("Rat", BuildRat(vm))

	return vm
}

func BuildRat(vm *otto.Otto) *otto.Object {
	rat, _ := vm.Object("({})")

	rat.Set("addEventHandler", func(call otto.FunctionCall) otto.Value {
		keyEvent := call.Argument(0).String()
		callback := call.Argument(1)

		AddEventHandler(keyEvent, func() {
			callback.Call(otto.NullValue())
		})

		return otto.NullValue()
	})

	rat.Set("log", func(call otto.FunctionCall) otto.Value {
		return otto.NullValue()
	})

	rat.Set("openPager", func(call otto.FunctionCall) otto.Value {
		command := call.Argument(0).String()
		options := call.Argument(1).Object()

		context, _ := options.Get("context")
		contextObj := context.Object()

		ctx := Context{}

		for _, k := range contextObj.Keys() {
			v, _ := contextObj.Get(k)
			ctx[k] = v.String()
		}

		pager := NewCmdPager("", command, ctx)

		childKeys, errChild := options.Get("child")

		if errChild == nil {
			AddChildPager(pagers.ParentPager(), pager, childKeys.String())
		} else {
			PushPager(pager)
		}

		return otto.NullValue()
	})

	rat.Set("defineMode", func(call otto.FunctionCall) otto.Value {
		name := call.Argument(0).String()
		callback := call.Argument(1)

		mode := NewMode()

		jsmode, _ := vm.Object("({})")

		jsmode.Set("annotate", func(call otto.FunctionCall) otto.Value {
			options := call.Argument(0).Object()
			//strategy, _ := options.Get("strategy").String()
			class, _ := options.Get("class")
			command, _ := options.Get("command")

			mode.RegisterAnnotator(func(ctx Context) Annotator {
				return NewMatchAnnotator(
					command.String(),
					class.String(),
					ctx,
				)
			})

			return otto.NullValue()
		})

		jsmode.Set("addEventHandler", func(call otto.FunctionCall) otto.Value {
			key := call.Argument(0).String()
			requirements := call.Argument(1).String()
			callback := call.Argument(2)

			mode.RegisterEventHandler(func(ctx Context) func(Pager) {
				return func(p Pager) {
					p.AddEventHandler(key, NewCtxEventHandler(strings.Split(requirements, ","), func(ectx Context) {
						callback.Call(otto.NullValue(), ectx)
					}))
				}
			})

			return otto.NullValue()
		})

		callback.Call(otto.NullValue(), jsmode)
		RegisterMode(name, mode)

		return otto.NullValue()
	})

	rat.Set("Actions", BuildActions(vm))

	return rat
}

func BuildActions(vm *otto.Otto) *otto.Object {
	actions, _ := vm.Object("({})")

	actions.Set("moveCursor", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.Last().MoveCursor(int(offset))

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

	actions.Set("scroll", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.Last().Scroll(int(offset))

		return otto.NullValue()
	})

	actions.Set("moveParentCursor", func(call otto.FunctionCall) otto.Value {
		offset, _ := call.Argument(0).ToInteger()

		pagers.MoveParentCursor(int(offset))

		return otto.NullValue()
	})

	return actions
}
