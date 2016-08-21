package rat

type mode struct {
	annotatorCtors      []func(Context) Annotator
	eventListenerAdders []func(Context) func(Pager)
}

func NewMode() Mode {
	m := &mode{}

	m.annotatorCtors = make([]func(Context) Annotator, 0, 8)
	m.eventListenerAdders = make([]func(Context) func(Pager), 0, 8)

	return m
}

func (m *mode) InitParsers(ctx Context) func() []Annotator {
	return func() []Annotator {
		annotators := make([]Annotator, 0, len(m.annotatorCtors))

		for _, ctor := range m.annotatorCtors {
			annotators = append(annotators, ctor(ctx))
		}

		return annotators
	}
}

func (m *mode) AddEventListeners(ctx Context) func(Pager) {
	return func(p Pager) {
		for _, adder := range m.eventListenerAdders {
			adder(ctx)(p)
		}
	}
}

func (m *mode) RegisterAnnotator(ctor func(Context) Annotator) {
	m.annotatorCtors = append(m.annotatorCtors, ctor)
}

func (m *mode) RegisterEventListener(adder func(Context) func(Pager)) {
	m.eventListenerAdders = append(m.eventListenerAdders, adder)
}
