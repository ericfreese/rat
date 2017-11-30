package rat

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/nsf/termbox-go"
)

var (
	events        chan termbox.Event
	done          chan bool
	eventHandlers HandlerRegistry
	modes         map[string]Mode
	cfg           Configurer
	annotatorsDir string
	keyStack      []keyEvent

	widgets WidgetStack
	pagers  PagerStack
	prompt  ConfirmPrompt
)

func Init() error {
	if err := initTermbox(); err != nil {
		return err
	}

	events = make(chan termbox.Event)
	widgets = NewWidgetStack()
	pagers = NewPagerStack()
	done = make(chan bool)
	eventHandlers = NewHandlerRegistry()
	modes = make(map[string]Mode)
	cfg = NewConfigurer()

	widgets.Push(pagers)
	prompt = NewConfirmPrompt()

	AddEventHandler("C-c", Quit)

	w, h := termbox.Size()
	layout(w, h)

	return nil
}

func closeTermbox() {
	termbox.Close()
}

func initTermbox() error {
	var err error

	if err = termbox.Init(); err != nil {
		return err
	}

	termbox.SetInputMode(termbox.InputAlt)
	termbox.SetOutputMode(termbox.Output256)

	return nil
}

func SetAnnotatorsDir(dir string) {
	annotatorsDir = dir
}

func LoadConfig(rd io.Reader) {
	cfg.Process(rd)
}

func Close() {
	closeTermbox()
}

func Quit() {
	close(done)
}

func Run() {
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()

loop:
	for {
		render()

		select {
		case <-done:
			break loop
		case e := <-events:
			switch e.Type {
			case termbox.EventKey:
				keyStack = append(keyStack, KeyEventFromTBEvent(&e))

				if handleEvent(keyStack) {
					keyStack = nil
				}
			case termbox.EventResize:
				layout(e.Width, e.Height)
			}
		case <-time.After(time.Second / 10):
		}
	}

	widgets.Destroy()
}

func AddChildPager(parent, child Pager, creatingKeys string) {
	pagers.AddChild(parent, child, creatingKeys)
}

func PushPager(p Pager) {
	pagers.Push(p)
}

func PopPager() {
	pagers.Pop()

	if pagers.Size() == 0 {
		Quit()
	}
}

func Confirm(message string, callback func()) {
	prompt.Confirm(message, callback)
}

func ConfirmExec(cmd string, ctx Context, callback func()) {
	Confirm(fmt.Sprintf("Run `%s`", cmd), func() {
		Exec(cmd, ctx)
		callback()
	})
}

func Exec(cmd string, ctx Context) {
	c := exec.Command(os.Getenv("SHELL"), "-c", cmd)

	c.Env = ContextEnvironment(ctx)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr

	closeTermbox()
	defer initTermbox()

	c.Run()
}

func RegisterMode(name string, mode Mode) {
	modes[name] = mode
}

func layout(width, height int) {
	widgets.SetBox(NewBox(0, 0, width, height-1))
	prompt.SetBox(NewBox(0, height-1, width, 1))
}

func render() {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	widgets.Render()
	prompt.Render()
	termbox.Flush()
}

func AddEventHandler(keyStr string, handler func()) {
	eventHandlers.Add(KeySequenceFromString(keyStr), NewEventHandler(handler))
}

func handleEvent(ks []keyEvent) bool {
	if prompt.HandleEvent(ks) {
		return true
	}

	if widgets.HandleEvent(ks) {
		return true
	}

	if handler := eventHandlers.Find(ks); handler != nil {
		handler.Call(nil)
		return true
	}

	return false
}
