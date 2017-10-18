package rat

type EventHandler interface {
	Supports(Context) bool
	Call(Context)
	Specificity() int
}

type eventHandler struct {
	handler func()
}

func NewEventHandler(h func()) EventHandler {
	eh := &eventHandler{}
	eh.handler = h
	return eh
}

func (eh *eventHandler) Supports(ctx Context) bool {
	return true
}

func (eh *eventHandler) Call(ctx Context) {
	eh.handler()
}

func (eh *eventHandler) Specificity() int {
	return 1
}

type ctxEventHandler struct {
	requirements []string
	handler      func(Context)
}

func NewCtxEventHandler(req []string, h func(Context)) EventHandler {
	ceh := &ctxEventHandler{}

	ceh.requirements = req
	ceh.handler = h

	return ceh
}

func (ceh *ctxEventHandler) Supports(ctx Context) bool {
	if ctx == nil && len(ceh.requirements) > 0 {
		return false
	}

	for _, t := range ceh.requirements {
		if _, ok := ctx[t]; !ok {
			return false
		}
	}

	return true
}

func (ceh *ctxEventHandler) Call(ctx Context) {
	ceh.handler(ctx)
}

func (ceh *ctxEventHandler) Specificity() int {
	return len(ceh.requirements)
}

type HandlerRegistry interface {
	Add([]keyEvent, EventHandler)
	Find([]keyEvent) EventHandler
	FindCtx([]keyEvent, Context) EventHandler
}

type handlerRegistry struct {
	children map[keyEvent]*handlerRegistry
	handlers []EventHandler
}

func NewHandlerRegistry() HandlerRegistry {
	hn := &handlerRegistry{}
	hn.children = make(map[keyEvent]*handlerRegistry)
	return hn
}

func (hn *handlerRegistry) Add(ks []keyEvent, h EventHandler) {
	if len(ks) == 0 {
		return
	}

	hn.add(ks, h)
}

func (hn *handlerRegistry) Find(ks []keyEvent) EventHandler {
	return hn.FindCtx(ks, nil)
}

func (hn *handlerRegistry) FindCtx(ks []keyEvent, ctx Context) EventHandler {
	if len(ks) == 0 {
		return nil
	}

	return hn.find(ks, ctx, nil)
}

func (hn *handlerRegistry) add(ks []keyEvent, h EventHandler) {
	n := len(ks)

	if n == 0 {
		hn.handlers = append(hn.handlers, h)
	} else {
		child, ok := hn.children[ks[n-1]]

		if !ok {
			child = &handlerRegistry{
				children: make(map[keyEvent]*handlerRegistry),
				handlers: make([]EventHandler, 0),
			}
			hn.children[ks[n-1]] = child
		}

		child.add(ks[:n-1], h)
	}
}

func (hn *handlerRegistry) find(ks []keyEvent, ctx Context, candidate EventHandler) EventHandler {
	n := len(ks)

	if h := hn.handlerFor(ctx); h != nil {
		candidate = h
	}

	if n == 0 {
		return candidate
	}

	child, ok := hn.children[ks[n-1]]

	if !ok {
		return candidate
	}

	return child.find(ks[:n-1], ctx, candidate)
}

func (hn *handlerRegistry) handlerFor(ctx Context) EventHandler {
	var mostSpecific EventHandler

	for _, h := range hn.handlers {
		if h.Supports(ctx) && (mostSpecific == nil || h.Specificity() > mostSpecific.Specificity()) {
			mostSpecific = h
		}
	}

	return mostSpecific
}
