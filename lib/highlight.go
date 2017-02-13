package rat

type Highlight interface {
	TermStyle
}

type Highlights interface {
	Start(int, TermStyle)
	End(int)
	AtPoint(int) Highlight
	Len() int
}

type highlight struct {
	start int
	end   int
	TermStyle
}

type highlights struct {
	highlights []Highlight
	inProgress *highlight
	index      map[int]Highlight
}

func NewHighlights() Highlights {
	c := &highlights{}

	c.highlights = make([]Highlight, 0, 8)
	c.index = make(map[int]Highlight)

	return c
}

func (c *highlights) add(h Highlight) {
	c.highlights = append(c.highlights, h)
}

func (c *highlights) Start(offset int, ts TermStyle) {
	c.End(offset)

	c.inProgress = &highlight{offset, -1, ts}
	c.add(c.inProgress)
}

func (c *highlights) End(offset int) {
	if c.inProgress == nil {
		return
	}

	c.inProgress.end = offset

	for p := c.inProgress.start; p < c.inProgress.end; p++ {
		c.index[p] = c.inProgress
	}

	c.inProgress = nil
}

func (c *highlights) AtPoint(p int) Highlight {
	if h, ok := c.index[p]; ok {
		return h
	} else if c.inProgress != nil && p >= c.inProgress.start {
		return c.inProgress
	}

	return nil
}

func (c *highlights) Len() int {
	return len(c.highlights)
}
