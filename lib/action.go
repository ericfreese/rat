package rat

var actions = make(map[string]func())

func init() {
	actions["reload"] = func() {
		pagers.Last().Reload()
	}

	actions["cursor-up"] = func() {
		pagers.Last().CursorUp()
	}

	actions["cursor-down"] = func() {
		pagers.Last().CursorDown()
	}

	actions["parent-cursor-up"] = func() {
		pagers.ParentCursorUp()
	}

	actions["parent-cursor-down"] = func() {
		pagers.ParentCursorDown()
	}

	actions["cursor-first-line"] = func() {
		pagers.Last().CursorFirstLine()
	}

	actions["cursor-last-line"] = func() {
		pagers.Last().CursorLastLine()
	}

	actions["scroll-up"] = func() {
		pagers.Last().ScrollUp()
	}

	actions["scroll-down"] = func() {
		pagers.Last().ScrollDown()
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
