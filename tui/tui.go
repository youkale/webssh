package tui

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gliderlabs/ssh"
	zone "github.com/lrstanley/bubblezone"
	"github.com/muesli/termenv"
	"github.com/youkale/echogy/logger"
)

var (
	// Modern color palette
	subtle    = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#4A4A4A"}
	highlight = lipgloss.AdaptiveColor{Light: "#2188ff", Dark: "#58a6ff"} // GitHub-like blue
	special   = lipgloss.AdaptiveColor{Light: "#28a745", Dark: "#3fb950"} // GitHub-like green

	// Additional theme colors
	primary    = lipgloss.AdaptiveColor{Light: "#1a73e8", Dark: "#58a6ff"} // Modern blue
	secondary  = lipgloss.AdaptiveColor{Light: "#666666", Dark: "#8b949e"} // Subtle gray
	accent     = lipgloss.AdaptiveColor{Light: "#00c853", Dark: "#3fb950"} // Vibrant green
	background = lipgloss.AdaptiveColor{Light: "#ffffff", Dark: "#0d1117"} // GitHub-like dark theme
	border     = lipgloss.AdaptiveColor{Light: "#e1e4e8", Dark: "#30363d"} // Subtle border
)

// Table styles
var (
	headerStyle = lipgloss.NewStyle().
			Foreground(primary).
			Bold(true).
			PaddingLeft(2)

	rowStyle = lipgloss.NewStyle().
			Foreground(secondary).
			PaddingLeft(2)
)

type notification interface {
	tea.Model
	notify(message interface{})
}

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
			m.quitFunc()
			return m, quit
		}
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
		msg.Height -= 2
		msg.Width -= 4
		return m.propagate(msg), nil
	}

	return m.propagate(msg), nil
}

func (m *model) propagate(msg tea.Msg) tea.Model {
	// Propagate to all children.
	m.tabs, _ = m.tabs.Update(msg)

	if msg, ok := msg.(tea.WindowSizeMsg); ok {
		msg.Height -= m.tabs.(*tabs).height
		return m
	}

	return m
}

func (m model) View() string {
	if !m.isInitialized() {
		return ""
	}

	s := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		Width(m.width-2).   // Account for border
		Height(m.height-2). // Account for border
		Padding(0, 1)

	content := lipgloss.JoinVertical(lipgloss.Left,
		"",
		m.tabs.View(),
		"",
	)

	return zone.Scan(s.Render(content))
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

type sshEnviron struct {
	environ []string
}

func (s *sshEnviron) Getenv(key string) string {
	for _, v := range s.environ {
		if strings.HasPrefix(v, key+"=") {
			return v[len(key)+1:]
		}
	}
	return ""
}

func (s *sshEnviron) Environ() []string {
	return s.environ
}

const (
	maxWidth  = 80
	maxHeight = 220
)

// NewPty creates a new terminal UI instance
func NewPty(sess ssh.Session, addr map[string]string) (*Tui, error) {
	pty, windowCh, hasPty := sess.Pty()

	if !hasPty {
		return nil, errors.New("no pty")
	}

	environ := sess.Environ()
	environ = append(environ, fmt.Sprintf("TERM=%s", pty.Term))

	// Add color support detection
	if !termenv.EnvNoColor() {
		switch {
		case strings.Contains(strings.ToLower(pty.Term), "256color"):
			environ = append(environ, "COLORTERM=256color")
		case strings.Contains(strings.ToLower(pty.Term), "color"):
			environ = append(environ, "COLORTERM=color")
		}
		// Add FORCE_COLOR to ensure color output
		environ = append(environ, "FORCE_COLOR=true")
	}

	ctx := sess.Context()

	// Create renderer with color support
	renderer := lipgloss.NewRenderer(sess,
		termenv.WithUnsafe(),
		termenv.WithEnvironment(&sshEnviron{environ}),
	)

	zone.NewGlobal()

	// Create tabs with initial items
	t := newTabs(pty.Window.Height-3,
		TabItem{
			Name:  "Summary",
			Model: &summary{addresses: addr},
		},
		TabItem{
			Name:  "Requests",
			Model: newHistory(),
		},
	)

	exChan := make(chan *httpExchange, 2)

	m := &model{
		quitFunc: func() {
			time.AfterFunc(1*time.Second, func() {
				sess.Close()
			})
		},
		tabs: t,
	}

	// Configure bubbletea program with SSH-specific options
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithOutput(renderer.Output()),
		tea.WithInput(sess),
		tea.WithMouseCellMotion(),
		tea.WithContext(ctx),
	)

	// Start window size monitoring
	go func() {
		for {
			select {
			case <-ctx.Done():
				p.Quit()
				return
			case exch := <-exChan:
				t.notify(exch)
				p.Send(tea.ShowCursor())
			case win := <-windowCh:
				if m.width != win.Width || m.height != win.Height {
					p.Send(tea.WindowSizeMsg{
						Width:  min(maxWidth, win.Width),
						Height: min(maxHeight, win.Height),
					})
				}
			}
		}
	}()

	return &Tui{Program: p, exchangeChan: exChan}, nil
}
