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
	"github.com/muesli/termenv"
	"github.com/youkale/echogy/logger"
)

type Tui struct {
	*tea.Program
	exchangeChan chan *httpExchange
}

type httpExchange struct {
	*http.Response
	*http.Request
	useTime int64
}

func (t *Tui) Notify(w *http.Response, r *http.Request, useTime int64) {
	t.exchangeChan <- &httpExchange{Response: w, Request: r, useTime: useTime}
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
	maxWidth  = 860
	maxHeight = 480
)

// NewPty creates a new terminal UI instance
func NewPty(sess ssh.Session, addr string) (*Tui, error) {
	pty, windowCh, hasPty := sess.Pty()
	if !hasPty {
		return nil, errors.New("no pty")
	}

	// Setup terminal environment
	environ := setupTerminalEnv(sess, pty)
	ctx := sess.Context()

	// Configure terminal renderer with color support
	renderer := lipgloss.NewRenderer(sess,
		termenv.WithEnvironment(&sshEnviron{environ}),
		termenv.WithUnsafe(),
		termenv.WithProfile(termenv.ANSI256))

	// Initialize dashboard
	exChan := make(chan *httpExchange, 2)
	m := newDashboard(addr, pty.Window.Width, pty.Window.Height, func() {
		time.AfterFunc(200*time.Millisecond, func() {
			if sess != nil {
				sess.Close()
			}
		})
	})

	// Configure bubbletea program
	p := tea.NewProgram(
		m,
		tea.WithAltScreen(),
		tea.WithOutput(renderer.Output()),
		tea.WithInput(sess),
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
				m.AddRequest(exch)
				p.Send(tea.ShowCursor())
			case newSize := <-windowCh:
				if newSize.Height == 0 || newSize.Width == 0 {
					continue
				}
				p.Send(tea.WindowSizeMsg{
					Width:  min(newSize.Width, maxWidth),
					Height: min(newSize.Height, maxHeight),
				})
			}
		}
	}()

	return &Tui{
		Program:      p,
		exchangeChan: exChan,
	}, nil
}

// setupTerminalEnv configures the terminal environment variables for proper color support
func setupTerminalEnv(sess ssh.Session, pty ssh.Pty) []string {
	environ := sess.Environ()
	environ = append(environ, fmt.Sprintf("TERM=%s", pty.Term))

	// Add color support environment variables
	environ = append(environ, "CLICOLOR=1")
	environ = append(environ, "CLICOLOR_FORCE=1")

	// Set color term based on terminal type
	termType := strings.ToLower(pty.Term)
	if strings.Contains(termType, "256color") {
		environ = append(environ, "COLORTERM=truecolor")
	} else if strings.Contains(termType, "color") || strings.Contains(termType, "xterm") {
		environ = append(environ, "COLORTERM=color")
	}

	return environ
}
