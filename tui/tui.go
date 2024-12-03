package tui

import (
	"errors"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gliderlabs/ssh"
	zone "github.com/lrstanley/bubblezone"
	"github.com/youkale/webssh/logger"
	"net/http"
	"time"
)

var (
	subtle    = lipgloss.AdaptiveColor{Light: "#D9DCCF", Dark: "#383838"}
	highlight = lipgloss.AdaptiveColor{Light: "#874BFD", Dark: "#7D56F4"}
	special   = lipgloss.AdaptiveColor{Light: "#43BF6D", Dark: "#73F59F"}

	// Additional theme colors
	primary    = lipgloss.AdaptiveColor{Light: "#2E8B57", Dark: "#00FF7F"} // Sea Green / Spring Green
	secondary  = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#A0A0A0"}
	accent     = lipgloss.AdaptiveColor{Light: "#3CB371", Dark: "#98FB98"} // Medium Sea Green / Pale Green
	background = lipgloss.AdaptiveColor{Light: "#FFFFFF", Dark: "#1A1A1A"}
	border     = lipgloss.AdaptiveColor{Light: "#E0E0E0", Dark: "#4A4A4A"}
)

type model struct {
	height   int
	width    int
	quitFunc func()
	tabs     tea.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) isInitialized() bool {
	return m.height != 0 && m.width != 0
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !m.isInitialized() {
		if _, ok := msg.(tea.WindowSizeMsg); !ok {
			return m, nil
		}
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Example of toggling mouse event tracking on/off.
		if msg.String() == "ctrl+e" {
			zone.SetEnabled(!zone.Enabled())
			return m, nil
		}

		if msg.String() == "ctrl+c" {
			quit := tea.Quit
			after := time.After(1 * time.Second)
			go func() {
				<-after
				m.quitFunc()
			}()
			return m, quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		msg.Height -= 2
		msg.Width -= 4
		return m.propagate(msg), nil

	case httpExchange:
		m.tabs.Update(msg)
	}

	return m.propagate(msg), nil
}

func (m *model) propagate(msg tea.Msg) tea.Model {
	// Propagate to all children.
	m.tabs, _ = m.tabs.Update(msg)

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		msg.Height -= m.tabs.(tabs).height
		return m
	}

	return m
}

func (m model) View() string {
	if !m.isInitialized() {
		return ""
	}

	s := lipgloss.NewStyle().MaxHeight(m.height).MaxWidth(m.width).Padding(1, 2, 1, 2)

	return zone.Scan(s.Render(lipgloss.JoinVertical(lipgloss.Top,
		m.tabs.View(), "",
		lipgloss.PlaceHorizontal(
			m.width, lipgloss.Center,
			lipgloss.JoinHorizontal(
				lipgloss.Top,
			),
			lipgloss.WithWhitespaceChars(" "),
		),
	)))
}

type Tui struct {
	*tea.Program
	exchangeChan chan *httpExchange
}

type httpExchange struct {
	*http.Response
	*http.Request
}

func (t *Tui) Notify(w *http.Response, r *http.Request) {
	t.exchangeChan <- &httpExchange{Response: w, Request: r}
}

func (t *Tui) Start() error {
	_, err := t.Run()
	if err != nil {
		logger.Error("run pty", err, map[string]interface{}{
			"module": "tui",
		})
		return err
	}
	return nil
}

func NewPty(sess ssh.Session) (*Tui, error) {
	pty, windowCh, hasPty := sess.Pty()

	if !hasPty {
		return nil, errors.New("no pty")
	}

	ctx := sess.Context()

	zone.NewGlobal()

	t := &tabs{
		id:     zone.NewPrefix(),
		height: 3,
		active: "Requests",
		tabItems: map[string]tea.Model{
			"Requests": newHistory(),
		},
	}

	exChan := make(chan *httpExchange, 2)

	m := &model{
		quitFunc: func() {
			sess.Close()
		},
		width:  pty.Window.Width,
		height: pty.Window.Height,
		tabs:   t,
	}

	// Configure bubbletea program with SSH-specific options
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithOutput(sess),
		tea.WithInput(sess),
		tea.WithMouseCellMotion(),
	)

	go func() {
		for {
			select {
			case <-ctx.Done():
				p.Quit()
				return
			case exch := <-exChan:
				p.Send(exch)
			case win := <-windowCh:
				if m.width != win.Width || m.height != win.Height {
					p.Send(tea.WindowSizeMsg{
						Width:  win.Width,
						Height: win.Height,
					})
				}
			}
		}
	}()
	return &Tui{Program: p, exchangeChan: exChan}, nil
}
