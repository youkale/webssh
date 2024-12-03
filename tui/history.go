package tui

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
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

func (h *history) notify(msg interface{}) {
	switch msg.(type) {
	case *httpExchange:
		h.httpExchanges.Push(msg.(*httpExchange))
	}
}

func (h *history) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return h, nil
}

func getStatusColor(status int) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	switch {
	case status >= 500:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#d73a49", Dark: "#f85149"}) // Red
	case status >= 400:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#f66a0a", Dark: "#ffa657"}) // Orange
	case status >= 300:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#0366d6", Dark: "#58a6ff"}) // Blue
	case status >= 200:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#28a745", Dark: "#3fb950"}) // Green
	default:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#6c757d", Dark: "#6c757d"}) // Secondary
	}
}

func getMethodStyle(method string) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	switch method {
	case "GET":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#0366d6", Dark: "#58a6ff"}) // Blue
	case "POST":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#28a745", Dark: "#3fb950"}) // Green
	case "PUT":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#f66a0a", Dark: "#ffa657"}) // Orange
	case "DELETE":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#d73a49", Dark: "#f85149"}) // Red
	default:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#6c757d", Dark: "#6c757d"}) // Secondary
	}
}

func (h *history) View() string {
	var rows [][]string

	// Headers
	headers := []string{"No.", "Code", "Method", "Type", "Rec.", "Send.", "Uri"}
	rows = append(rows, headers)

	// Data rows
	for idx, item := range h.httpExchanges.Items() {
		ex := item.(*httpExchange)
		contentType := parseContentType(ex.Response.Header.Get("Content-Type"))
		status := strconv.Itoa(ex.StatusCode)
		reqLen := orStr(ex.Request.Header.Get("Content-Length"), "0")
		respLen := orStr(ex.Response.Header.Get("Content-Length"), "0")

		// Format URI with query parameters
		uri := ex.RequestURI
		if len(uri) > 35 {
			uri = uri[:35] + "..."
		}

		row := []string{
			strconv.Itoa(idx + 1),
			getStatusColor(ex.StatusCode).Render(status),
			getMethodStyle(ex.Method).Render(ex.Method),
			contentType,
			humanSize(reqLen),
			humanSize(respLen),
			uri,
		}
		rows = append(rows, row)
	}

	// Create table with specific column widths
	t := table.New().
		Border(lipgloss.Border{}).
		Headers(headers...).
		Rows(rows[1:]...).
		StyleFunc(func(row, col int) lipgloss.Style {
			// Apply consistent width and alignment for both headers and rows
			var style lipgloss.Style
			switch col {
			case 0: // No.
				style = lipgloss.NewStyle().Width(4).Align(lipgloss.Left)
			case 1: // Status
				style = lipgloss.NewStyle().Width(5).Align(lipgloss.Center)
			case 2: // Method
				style = lipgloss.NewStyle().Width(7).Align(lipgloss.Center)
			case 3: // Type
				style = lipgloss.NewStyle().Width(7).Align(lipgloss.Center)
			case 4, 5: // Rec., Send.
				style = lipgloss.NewStyle().Width(6).Align(lipgloss.Center)
			default: // Uri
				style = lipgloss.NewStyle().Width(40).PaddingLeft(2).Align(lipgloss.Left)
			}
			if row == -1 {
				return style.Inherit(headerStyle)
			}
			return style.Inherit(rowStyle)
		})

	return "\n" + t.String() + "\n"
}
