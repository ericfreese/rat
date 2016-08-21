package rat

type annotation struct {
	start BufferPoint
	end   BufferPoint
	class string
	val   string
}

func NewAnnotation(start, end BufferPoint, class string, val string) Annotation {
	return &annotation{start, end, class, val}
}

func (a *annotation) Start() BufferPoint {
	return a.start
}

func (a *annotation) End() BufferPoint {
	return a.end
}

func (a *annotation) Class() string {
	return a.class
}

func (a *annotation) Val() string {
	return a.val
}
