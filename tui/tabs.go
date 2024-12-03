package tui

import (
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
			Foreground(primary).
			Bold(true)

	tabGap = lipgloss.NewStyle().
		BorderStyle(lipgloss.Border{
			Bottom: "─",
		}).
		BorderForeground(subtle)

	helpStyle = lipgloss.NewStyle().
			Foreground(subtle).
			PaddingLeft(2).
			PaddingTop(1).
			PaddingBottom(1).
			BorderTop(true).
			BorderStyle(lipgloss.Border{
			Top: "─",
		}).
		BorderForeground(subtle)

	shortcutStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary)
)

// TabItem represents a tab item with its model and name
type TabItem struct {
	Name  string
	Model tea.Model
}

type tabs struct {
	id         string
	height     int
	width      int
	active     string
	tabItems   map[string]TabItem
	sortKeys   []string
	isSwitched bool
}

// newTabs creates a new tabs instance with the given items
func newTabs(height int, items ...TabItem) *tabs {
	t := &tabs{
		id:       zone.NewPrefix(),
		height:   height,
		tabItems: make(map[string]TabItem, len(items)),
		sortKeys: make([]string, 0, len(items)),
	}

	for _, item := range items {
		t.tabItems[item.Name] = item
		t.sortKeys = append(t.sortKeys, item.Name)
	}

	if len(t.sortKeys) > 0 {
		t.active = t.sortKeys[0]
	}

	return t
}

func (t *tabs) Init() tea.Cmd {
	var cmds []tea.Cmd
	// Initialize all tab items
	for _, item := range t.tabItems {
		if cmd := item.Model.Init(); cmd != nil {
			cmds = append(cmds, cmd)
		}
	}
	return tea.Batch(cmds...)
}

func (t *tabs) updateHistory(history *httpExchange) {
	if !t.isSwitched {
		t.active = "Requests"
		t.isSwitched = true
	}
	if item, ok := t.tabItems["Requests"]; ok {
		if n, ok := item.Model.(notification); ok {
			n.notify(history)
		}
	}
}

func (t *tabs) notify(message interface{}) {
	switch msg := message.(type) {
	case *httpExchange:
		t.updateHistory(msg)
	}
}

func (t *tabs) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "tab", "right", "l":
			// Move to next tab
			for i, key := range t.sortKeys {
				if key == t.active && i < len(t.sortKeys)-1 {
					t.active = t.sortKeys[i+1]
					break
				}
			}
		case "shift+tab", "left", "h":
			// Move to previous tab
			for i, key := range t.sortKeys {
				if key == t.active && i > 0 {
					t.active = t.sortKeys[i-1]
					break
				}
			}
		}
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		// Propagate window size to active tab
		if item, ok := t.tabItems[t.active]; ok {
			if model, cmd := item.Model.Update(msg); cmd != nil {
				item.Model = model
				cmds = append(cmds, cmd)
			}
		}
	case tea.MouseMsg:
		if msg.Action != tea.MouseActionRelease || msg.Button != tea.MouseButtonLeft {
			return t, nil
		}
		for key := range t.tabItems {
			if zone.Get(t.id + key).InBounds(msg) {
				t.active = key
				break
			}
		}
	}
	return t, tea.Batch(cmds...)
}

func (t *tabs) helpHeight() int {
	return 1 // 1 for content + 1 for top border + 1 for padding
}

func (t *tabs) View() string {
	var out []string

	// Render tabs
	for _, key := range t.sortKeys {
		style := tab
		if key == t.active {
			style = activeTab
		}
		out = append(out, zone.Mark(t.id+key, style.Render(key)))
	}

	// Join tabs horizontally
	row := lipgloss.JoinHorizontal(lipgloss.Top, out...)

	// Add gap with bottom border
	gapWidth := max(0, t.width-lipgloss.Width(row)-2)
	gap := tabGap.Width(gapWidth).Render("")
	row = lipgloss.JoinHorizontal(lipgloss.Bottom, row, gap)

	// Calculate available height for content
	contentHeight := t.height - t.helpHeight() - lipgloss.Height(row) - 4

	// Render active tab content
	var content string
	if item, ok := t.tabItems[t.active]; ok {
		content = item.Model.View()
	}

	// Create keyboard shortcuts help with full width
	help := helpStyle.Width(t.width).Render(
		"Shortcuts: " +
			shortcutStyle.Render("←/h") + " Previous tab • " +
			shortcutStyle.Render("→/l") + " Next tab • " +
			shortcutStyle.Render("ctrl+c") + " Quit",
	)

	// Create a container for the content that takes all available space
	contentStyle := lipgloss.NewStyle().
		Height(contentHeight).
		MaxHeight(contentHeight)

	// Join everything vertically with the help section at the bottom
	return lipgloss.JoinVertical(
		lipgloss.Left,
		row,
		contentStyle.Render(content),
		help,
	)
}
