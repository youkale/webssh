package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	zone "github.com/lrstanley/bubblezone"
)

var (
	activeTabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      " ",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┘",
		BottomRight: "└",
	}

	tabBorder = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "┴",
		BottomRight: "┴",
	}

	tab = lipgloss.NewStyle().
		Border(tabBorder, true).
		BorderForeground(subtle).
		Foreground(secondary).
		Padding(0, 1)

	activeTab = tab.
			Border(activeTabBorder, true).
			BorderForeground(primary).
			Foreground(primary)

	tabGap = tab.
		BorderTop(false).
		BorderLeft(false).
		BorderRight(false)
)

type tabs struct {
	id       string
	height   int
	width    int
	active   string
	tabItems map[string]tea.Model
}

type activeTabMsg struct {
	tabName string
}

func (m tabs) Init() tea.Cmd {
	return nil
}

func (m tabs) updateHistory(history httpExchange) {
	m.tabItems["Requests"].Update(history)
}

func (m tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case activeTabMsg:
		m.active = msg.tabName
	case httpExchange:
		m.updateHistory(msg)
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}
		for key := range m.tabItems {
			if zone.Get(m.id + key).InBounds(msg) {
				m.active = key
				break
			}
		}
		return m, nil
	}
	return m, nil
}

func (m tabs) View() string {
	var out []string

	// Render tabs
	for key := range m.tabItems {
		if key == m.active {
			out = append(out, zone.Mark(m.id+key, activeTab.Render(key)))
		} else {
			out = append(out, zone.Mark(m.id+key, tab.Render(key)))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	if am, found := m.tabItems[m.active]; found {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			row,
			am.View(),
		)
	}
	return row
}
