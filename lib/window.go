package rat

type Window interface {
	MoveCursorTo(int)
	MoveCursor(int)
	ScrollTo(int)
	Scroll(int)
	PageUp()
	PageDown()
	GetCursor() int
	GetScroll() int
}

type window struct {
	scroll         int
	cursor         int
	getHeight      func() int
	getTotalHeight func() int
}

func NewWindow(getHeight, getTotalHeight func() int) Window {
	w := &window{}

	w.getHeight = getHeight
	w.getTotalHeight = getTotalHeight

	return w
}

func (w *window) MoveCursorTo(cursor int) {
	if cursor < 0 {
		w.cursor = w.getTotalHeight() + cursor
	} else if cursor >= w.getTotalHeight() {
		w.cursor = w.getTotalHeight() - 1
	} else {
		w.cursor = cursor
	}

	if w.cursor < w.scroll {
		w.ScrollTo(w.cursor)
	} else if w.cursor > w.scroll+w.getHeight()-1 {
		w.ScrollTo(w.cursor - (w.getHeight() - 1))
	}
}

func (w *window) MoveCursor(delta int) {
	dest := w.cursor + delta

	if dest < 0 {
		dest = 0
	} else if dest > w.getTotalHeight() {
		dest = w.getTotalHeight() - 1
	}

	w.MoveCursorTo(dest)
}

func (w *window) ScrollTo(scrollY int) {
	if scrollY < 0 {
		w.scroll = 0
	} else if scrollY >= w.getTotalHeight()-w.getHeight() {
		if w.getTotalHeight() > w.getHeight() {
			w.scroll = w.getTotalHeight() - w.getHeight()
		} else {
			w.scroll = 0
		}
	} else {
		w.scroll = scrollY
	}

	if w.cursor < w.scroll {
		w.MoveCursorTo(w.scroll)
	} else if w.cursor > w.scroll+w.getHeight()-1 {
		w.MoveCursorTo(w.scroll + w.getHeight() - 1)
	}
}

func (w *window) Scroll(delta int) {
	w.ScrollTo(w.scroll + delta)
}

func (w *window) PageUp() {
	w.Scroll(-w.getHeight())
}

func (w *window) PageDown() {
	w.Scroll(w.getHeight())
}

func (w *window) GetCursor() int {
	return w.cursor
}

func (w *window) GetScroll() int {
	return w.scroll
}
