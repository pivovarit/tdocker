package ui

type scrollState struct {
	offset     int
	autoScroll bool
}

func (s scrollState) up() scrollState {
	if s.offset > 0 {
		s.offset--
		s.autoScroll = false
	}
	return s
}

func (s scrollState) down(contentLen, viewHeight int) scrollState {
	maxOff := max(0, contentLen-viewHeight)
	if s.offset < maxOff {
		s.offset++
	}
	if s.offset >= maxOff {
		s.autoScroll = true
	}
	return s
}

func (s scrollState) top() scrollState {
	s.offset = 0
	s.autoScroll = false
	return s
}

func (s scrollState) bottom(contentLen, viewHeight int) scrollState {
	s.offset = max(0, contentLen-viewHeight)
	s.autoScroll = true
	return s
}
