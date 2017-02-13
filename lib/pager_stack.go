package rat

type PagerStack interface {
	Widget
	AddEventListener(string, func())
	Show(int)
	Push(p Pager)
	Pop()
	Size() int
	AddChild(parent Pager, child Pager, creatingKey string)
	PushAsChild(Pager, string)
	ParentCursorUp()
	ParentCursorDown()
}

type pagerStack struct {
	lastEl         *pagerStackElement
	size           int
	numToShow      int
	eventListeners map[keyEvent]func()
	box            Box
	validLayout    bool
}

type pagerStackElement struct {
	pager       Pager
	previous    *pagerStackElement
	creatingKey string
}

func NewPagerStack() PagerStack {
	ps := &pagerStack{}

	ps.numToShow = 3
	ps.eventListeners = make(map[keyEvent]func())
	ps.addDefaultListeners()

	return ps
}

func (ps *pagerStack) AddChild(parent, child Pager, creatingKey string) {
	for el := ps.lastEl; el != nil; el = el.previous {
		if el.pager != parent {
			ps.Pop()
		} else {
			ps.PushAsChild(child, creatingKey)
			return
		}
	}
}

func (ps *pagerStack) PushAsChild(p Pager, creatingKey string) {
	ps.lastEl = &pagerStackElement{p, ps.lastEl, creatingKey}
	ps.size++
	ps.validLayout = false
}

func (ps *pagerStack) Push(p Pager) {
	ps.PushAsChild(p, "")
}

func (ps *pagerStack) Pop() {
	if ps.lastEl == nil {
		return
	}

	ps.lastEl.pager.Destroy()
	ps.lastEl = ps.lastEl.previous
	ps.size--

	ps.validLayout = false
}

func (ps *pagerStack) Size() int {
	return ps.size
}

func (ps *pagerStack) Show(numToShow int) {
	if numToShow > 0 {
		ps.numToShow = numToShow
		ps.validLayout = false
	}
}

func (ps *pagerStack) SetBox(b Box) {
	ps.box = b
	ps.validLayout = false
}

func (ps *pagerStack) visiblePagers() []Pager {
	var n int

	if ps.numToShow > ps.size {
		n = ps.size
	} else {
		n = ps.numToShow
	}

	pagers := make([]Pager, n)

	for i, el := 0, ps.lastEl; i < n && el != nil; i, el = i+1, el.previous {
		pagers[n-i-1] = el.pager
	}

	return pagers
}

func (ps *pagerStack) splitHorizontal() bool {
	return ps.box.Width() > 180
}

func (ps *pagerStack) layout() {
	pagers := ps.visiblePagers()
	n := len(pagers)
	offset := 0

	var totalSize, size int

	if ps.splitHorizontal() {
		totalSize = ps.box.Width()
	} else {
		totalSize = ps.box.Height()
	}

	remaining := totalSize

	for i, p := range pagers {
		size = (remaining - (n - i - 1)) / (n - i)

		if ps.splitHorizontal() {
			p.SetBox(NewBox(offset, 0, size, ps.box.Height()))
		} else {
			p.SetBox(NewBox(0, offset, ps.box.Width(), size))
		}

		offset = offset + size + 1
		remaining = totalSize - offset
	}
}

func (ps *pagerStack) GetBox() Box {
	return ps.box
}

func (ps *pagerStack) drawVerticalDivider(offset int) {
	sr := NewStyledRune('│', gTermStyles.Default())
	for y := 0; y < ps.box.Height(); y++ {
		ps.box.DrawStyledRune(offset, y, sr)
	}
}

func (ps *pagerStack) drawHorizontalDivider(offset int) {
	sr := NewStyledRune('─', gTermStyles.Default())
	for x := 0; x < ps.box.Width(); x++ {
		ps.box.DrawStyledRune(x, offset, sr)
	}
}

func (ps *pagerStack) Render() {
	if !ps.validLayout {
		ps.layout()
	}

	pagers := ps.visiblePagers()

	for i, p := range pagers {
		p.Render()

		if i < len(pagers)-1 {
			pBox := p.GetBox()
			if ps.splitHorizontal() {
				ps.drawVerticalDivider(pBox.Left() + pBox.Width())
			} else {
				ps.drawHorizontalDivider(pBox.Top() + pBox.Height())
			}
		}
	}
}

func (ps *pagerStack) AddEventListener(keyStr string, handler func()) {
	ps.eventListeners[KeyEventFromString(keyStr)] = handler
}

func (ps *pagerStack) HandleEvent(ke keyEvent) bool {
	if ps.size == 0 {
		return false
	}

	if handler, ok := ps.eventListeners[ke]; ok && ps.size > 1 && ps.numToShow > 1 && len(ps.lastEl.creatingKey) > 0 {
		handler()
		return true
	}

	return ps.lastEl.pager.HandleEvent(ke)
}

func (ps *pagerStack) Destroy() {
	for ps.size > 0 {
		ps.Pop()
	}
}

func (ps *pagerStack) addDefaultListeners() {
	ps.AddEventListener("S-j", ps.ParentCursorDown)
	ps.AddEventListener("S-k", ps.ParentCursorUp)
}

func (ps *pagerStack) parentPager() Pager {
	if ps.size == 0 {
		return nil
	}

	if ps.size > 1 {
		return ps.lastEl.previous.pager
	} else {
		return ps.lastEl.pager
	}
}

func (ps *pagerStack) ParentCursorUp() {
	if len(ps.lastEl.creatingKey) > 0 {
		ps.parentPager().CursorUp()
		ps.parentPager().HandleEvent(KeyEventFromString(ps.lastEl.creatingKey))
	}
}

func (ps *pagerStack) ParentCursorDown() {
	if len(ps.lastEl.creatingKey) > 0 {
		ps.parentPager().CursorDown()
		ps.parentPager().HandleEvent(KeyEventFromString(ps.lastEl.creatingKey))
	}
}
