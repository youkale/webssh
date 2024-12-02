package tui

import (
	"fmt"
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
	id     string
	height int
	width  int

	active   string
	items    []string
	requests []string // Store request history
}

func (m tabs) Init() tea.Cmd {
	return nil
}

func (m tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return m, nil
		}

		for _, item := range m.items {
			if zone.Get(m.id + item).InBounds(msg) {
				m.active = item
				break
			}
		}

		return m, nil
	}
	return m, nil
}

// AddRequest adds a new request to the history
func (m *tabs) AddRequest(method, path, status string) {
	request := fmt.Sprintf("%s %s %s", method, path, status)
	m.requests = append(m.requests, request)
}

func (m *tabs) requestsView() string {
	var requestList []string
	requestStyle := lipgloss.NewStyle().
		Foreground(secondary).
		PaddingLeft(2)

	for _, req := range m.requests {
		requestList = append(requestList, requestStyle.Render(req))
	}

	return strings.Join(requestList, "\n")
}

func (m tabs) View() string {
	var out []string

	// Render tabs
	for _, item := range m.items {
		if item == m.active {
			out = append(out, zone.Mark(m.id+item, activeTab.Render(item)))
		} else {
			out = append(out, zone.Mark(m.id+item, tab.Render(item)))
		}
	}
	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)
	gap := tabGap.Render(strings.Repeat(" ", max(0, m.width-lipgloss.Width(row)-2)))
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	// If Connection tab is active, show requests
	if m.active == "Request" && len(m.requests) > 0 {
		return lipgloss.JoinVertical(
			lipgloss.Left,
			row,
			m.requestsView(),
		)
	}

	return row
}
