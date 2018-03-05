package rat

var actions = make(map[string]func())

func init() {
	actions["reload"] = func() {
		pagers.Last().Reload()
	}

	actions["cursor-up"] = func() {
		pagers.Last().MoveCursor(-1)
	}

	actions["cursor-down"] = func() {
		pagers.Last().MoveCursor(1)
	}

	actions["parent-cursor-up"] = func() {
		pagers.MoveParentCursor(-1)
	}

	actions["parent-cursor-down"] = func() {
		pagers.MoveParentCursor(1)
	}

	actions["cursor-first-line"] = func() {
		pagers.Last().MoveCursorTo(0)
	}

	actions["cursor-last-line"] = func() {
		pagers.Last().MoveCursorTo(-1)
	}

	actions["scroll-up"] = func() {
		pagers.Last().Scroll(-1)
	}

	actions["scroll-down"] = func() {
		pagers.Last().Scroll(1)
	}

	actions["page-up"] = func() {
		pagers.Last().PageUp()
	}

	actions["page-down"] = func() {
		pagers.Last().PageDown()
	}

	actions["quit"] = func() {
		Quit()
	}

	actions["pop-pager"] = func() {
		PopPager()
	}

	actions["show-one"] = func() {
		pagers.Show(1)
	}

	actions["show-two"] = func() {
		pagers.Show(2)
	}

	actions["show-three"] = func() {
		pagers.Show(3)
	}
}
