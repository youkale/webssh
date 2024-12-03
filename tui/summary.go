package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type summary struct {
	addresses map[string]string
	qrcode    string
}

func (s summary) Init() tea.Cmd {
	return nil
}

func (s summary) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if addr, found := s.addresses["alias"]; found {
		s.qrcode = addr
		return s, nil
	} else {
		if addr, found := s.addresses["access"]; found {
			s.qrcode = addr
			return s, nil
		}
	}
	return s, nil
}

func (s summary) View() string {
	return s.addresses["access"]
}
