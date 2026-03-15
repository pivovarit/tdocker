package ui

import tea "charm.land/bubbletea/v2"

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

func (s scrollState) handleKey(msg tea.KeyPressMsg, contentLen, viewHeight int) (scrollState, bool) {
	switch msg.Code {
	case tea.KeyUp, 'k':
		return s.up(), true
	case tea.KeyDown, 'j':
		return s.down(contentLen, viewHeight), true
	case 'g', tea.KeyHome:
		if msg.Text == keyScrollBottom {
			return s.bottom(contentLen, viewHeight), true
		}
		return s.top(), true
	case tea.KeyEnd:
		return s.bottom(contentLen, viewHeight), true
	}
	return s, false
}
