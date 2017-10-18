package rat

type Widget interface {
	SetBox(Box)
	GetBox() Box
	Render()
	HandleEvent([]keyEvent) bool
	Destroy()
}

type WidgetStack interface {
	Widget
	Push(w Widget)
	Pop() Widget
	Size() int
}

type widgetStack struct {
	lastEl *widgetStackElement
	box    Box
	size   int
}

type widgetStackElement struct {
	widget   Widget
	previous *widgetStackElement
}

func NewWidgetStack() WidgetStack {
	return &widgetStack{}
}

func (ws *widgetStack) Push(w Widget) {
	ws.lastEl = &widgetStackElement{w, ws.lastEl}
	ws.size++
	ws.layout()
}

func (ws *widgetStack) Pop() Widget {
	if ws.lastEl == nil {
		return nil
	}

	w := ws.lastEl.widget
	ws.lastEl = ws.lastEl.previous
	ws.size--

	w.Destroy()

	return w
}

func (ws *widgetStack) Size() int {
	return ws.size
}

func (ws *widgetStack) SetBox(b Box) {
	ws.box = b
	ws.layout()
}

func (ws *widgetStack) GetBox() Box {
	return ws.box
}

func (ws *widgetStack) Render() {
	if ws.lastEl != nil {
		ws.lastEl.widget.Render()
	}
}

func (ws *widgetStack) HandleEvent(ks []keyEvent) bool {
	return ws.lastEl.widget.HandleEvent(ks)
}

func (ws *widgetStack) Destroy() {
	for ws.size > 0 {
		ws.Pop()
	}
}

func (ws *widgetStack) layout() {
	for e := ws.lastEl; e != nil; e = e.previous {
		e.widget.SetBox(ws.box)
	}
}
