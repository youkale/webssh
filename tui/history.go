package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type history struct {
	httpExchanges *fixedLengthVec
}

func newHistory() *history {
	return &history{
		httpExchanges: newFixedLengthVec(32),
	}
}

func (h *history) Init() tea.Cmd {
	return nil
}

func (h *history) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case httpExchange:
		h.httpExchanges.Push(msg.(httpExchange))
	}
	return h, nil
}

func (h *history) View() string {

	return ""
}
