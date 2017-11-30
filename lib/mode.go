package rat

type Mode interface {
	RegisterAnnotator(func(Context) Annotator)
	RegisterEventHandler(func(Context) func(Pager))
	InitAnnotators(Context) func() []Annotator
	AddEventHandlers(Context) func(Pager)
}

type mode struct {
	annotatorCtors     []func(Context) Annotator
	eventHandlerAdders []func(Context) func(Pager)
}

func NewMode() Mode {
	m := &mode{}

	m.annotatorCtors = make([]func(Context) Annotator, 0, 8)
	m.eventHandlerAdders = make([]func(Context) func(Pager), 0, 8)

	return m
}

func (m *mode) InitAnnotators(ctx Context) func() []Annotator {
	return func() []Annotator {
		annotators := make([]Annotator, 0, len(m.annotatorCtors))

		for _, ctor := range m.annotatorCtors {
			annotators = append(annotators, ctor(ctx))
		}

		return annotators
	}
}

func (m *mode) AddEventHandlers(ctx Context) func(Pager) {
	return func(p Pager) {
		for _, adder := range m.eventHandlerAdders {
			adder(ctx)(p)
		}
	}
}

func (m *mode) RegisterAnnotator(ctor func(Context) Annotator) {
	m.annotatorCtors = append(m.annotatorCtors, ctor)
}

func (m *mode) RegisterEventHandler(adder func(Context) func(Pager)) {
	m.eventHandlerAdders = append(m.eventHandlerAdders, adder)
}
