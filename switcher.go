package my3status

// Switcher switches between it's constituent widgets when clicked on
type Switcher []Widget

func (s Switcher) Status() (StatusBlock, error) {
	return s[0].Status()
}

func (s Switcher) Click(c ClickEvent) bool {
	t := s[0]
	copy(s, s[1:])
	s[len(s)-1] = t
	return true
}
