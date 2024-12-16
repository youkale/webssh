package tui

import (
	"fmt"
	"strconv"
	"time"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/skip2/go-qrcode"
	q "github.com/youkale/echogy/pkg/queue"
)

// Constants for layout and styling
const (
	defaultTableHeight = 10
	maxRequestHistory  = 32
)

// Style definitions
var (
	dashStyle = lipgloss.NewStyle().Padding(1)

	headerStyle = lipgloss.NewStyle().Bold(true).Border(lipgloss.NormalBorder(), false, false, true, false).PaddingBottom(1)

	urlStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.AdaptiveColor{Light: "#3182CE", Dark: "#90CDF4"})

	statsStyle = lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Foreground(lipgloss.AdaptiveColor{Light: "#4A5568", Dark: "#A0AEC0"})

	// Column styles
	colMethodHeaderStyle = lipgloss.NewStyle().Align(lipgloss.Center).Bold(true)

	colNoStyle = lipgloss.NewStyle().Align(lipgloss.Left)

	colStatusStyle = lipgloss.NewStyle().Bold(true)

	colPathHeaderStyle = lipgloss.NewStyle().Bold(false)

	colPathStyle = lipgloss.NewStyle().Inherit(colPathHeaderStyle)

	colUseTimeHeaderStyle = lipgloss.NewStyle().Align(lipgloss.Right)

	colUseTimeStyle = lipgloss.NewStyle().Inherit(colUseTimeHeaderStyle).Foreground(lipgloss.AdaptiveColor{Light: "#4A5568", Dark: "#A0AEC0"})

	// QR code styles
	qrStyle = lipgloss.NewStyle().Align(lipgloss.Center).Foreground(lipgloss.AdaptiveColor{Light: "#1A1A1A", Dark: "#F1F1F1"})

	githubStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#3182CE", Dark: "#90CDF4"}).Bold(true)

	descStyle = lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "#4A5568", Dark: "#A0AEC0"}).PaddingTop(1)

	// Table styles
	tableStyle = table.Styles{
		Header: lipgloss.NewStyle().
			Bold(true),
		Selected: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{Light: "#EDF2F7", Dark: "#2D3748"}).
			Foreground(lipgloss.AdaptiveColor{Light: "#3182CE", Dark: "#90CDF4"}).
			Bold(true),
		Cell: lipgloss.NewStyle(),
	}
)

// TableColumn defines the structure for table column configuration
type TableColumn struct {
	Title  string
	Weight float64
}

// RequestTable wraps the table model with additional configuration
type RequestTable struct {
	*table.Model
	columns []TableColumn
}

func (r *RequestTable) Init() tea.Cmd {
	return nil
}

func (r *RequestTable) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var m table.Model
	m, cmd = r.Model.Update(msg)
	r.Model = &m
	return r, cmd
}

// newRequestTable creates a new request table with predefined columns
func newRequestTable(width int) *RequestTable {
	columns := []TableColumn{
		{Title: colNoStyle.Render("#"), Weight: 0.05},                 // 5% of available width
		{Title: colMethodHeaderStyle.Render("Method"), Weight: 0.1},   // 10% of available width
		{Title: colStatusStyle.Render("Status"), Weight: 0.1},         // 10% of available width
		{Title: colPathHeaderStyle.Render("Path"), Weight: 0.65},      // 65% of available width
		{Title: colUseTimeHeaderStyle.Render("UseTime"), Weight: 0.1}, // 10% of available width
	}

	cols := make([]table.Column, len(columns))
	for i, col := range columns {
		cols[i] = table.Column{
			Title: col.Title,
			Width: int(col.Weight * float64(width)),
		}
	}

	t := table.New(
		table.WithColumns(cols),
		table.WithFocused(true),
		table.WithHeight(defaultTableHeight),
	)

	t.SetStyles(tableStyle)

	return &RequestTable{
		Model:   &t,
		columns: columns,
	}
}

// Dashboard represents the main TUI dashboard
type Dashboard struct {
	width      int
	height     int
	quitFunc   func()
	tunnelInfo TunnelInfo
	table      *RequestTable
	requests   *q.FixedQueue
}

// TunnelInfo holds information about the tunnel connection
type TunnelInfo struct {
	URL       string
	ExpiresIn time.Duration
	BytesRecv int64
	BytesSent int64
	ReqCount  int
	ResCount  int
}

// newDashboard creates a new dashboard instance
func newDashboard(tunnelAddr string, width, height int, quitFunc func()) *Dashboard {
	return &Dashboard{
		quitFunc: quitFunc,
		tunnelInfo: TunnelInfo{
			URL:       tunnelAddr,
			ExpiresIn: 10 * time.Minute,
		},
		width:    width,
		height:   height,
		table:    newRequestTable(width),
		requests: q.NewFixedQueue(maxRequestHistory),
	}
}

func (d *Dashboard) availableWidth() int {
	return d.width - 4
}

// updateTableWidth adjusts table column widths based on terminal width
func (d *Dashboard) updateTableWidth() {
	if d.width <= 0 {
		return
	}

	// Calculate available width (accounting for margins)
	aw := d.availableWidth()

	// Apply new column widths
	tableColumns := make([]table.Column, len(d.table.columns))
	for i, col := range d.table.columns {
		tableColumns[i] = table.Column{
			Title: col.Title, // Keep original styled title
			Width: int(float64(aw) * col.Weight),
		}
	}

	d.table.SetColumns(tableColumns)
}

// Init implements tea.Model
func (d *Dashboard) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model
func (d *Dashboard) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			tea.Quit()
			d.quitFunc()
			return d, nil
		}
	case tea.WindowSizeMsg:
		d.width = msg.Width
		d.height = msg.Height
		d.table.SetHeight(msg.Height - 8)
		d.updateTableWidth()
	}

	// Always update the table model
	var m table.Model
	m, cmd = d.table.Model.Update(msg)
	d.table.Model = &m
	return d, cmd
}

// renderHeader renders the header section with URLs and stats
func (d *Dashboard) renderHeader() string {
	// URLs section
	leftURLS := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Inherit(urlStyle).Width(d.width/2).Render(fmt.Sprintf("HTTP:  http://%s", d.tunnelInfo.URL)),
		lipgloss.NewStyle().Inherit(urlStyle).Width(d.width/2).Render(fmt.Sprintf("HTTPS: https://%s", d.tunnelInfo.URL)),
	)

	// Stats section
	leftStats := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Inherit(statsStyle).Width(d.width/4).Render(fmt.Sprintf("↓ %s", humanBytes(d.tunnelInfo.BytesRecv))),
		lipgloss.NewStyle().Inherit(statsStyle).Width(d.width/4).Render(fmt.Sprintf("↑ %s", humanBytes(d.tunnelInfo.BytesSent))),
	)

	srCount := lipgloss.JoinVertical(
		lipgloss.Left,
		lipgloss.NewStyle().Inherit(statsStyle).Width(d.width/4).Align(lipgloss.Center).Render(fmt.Sprintf("ReqCount: %d", d.tunnelInfo.ReqCount)),
		lipgloss.NewStyle().Inherit(statsStyle).Width(d.width/4).Align(lipgloss.Center).Render(fmt.Sprintf("ResCount: %d", d.tunnelInfo.ResCount)),
	)

	rightStats := lipgloss.JoinHorizontal(lipgloss.Top, leftStats, srCount)
	return lipgloss.JoinHorizontal(lipgloss.Top, leftURLS, rightStats)
}

// renderProjectInfo renders the project information section
func (d *Dashboard) renderProjectInfo() string {
	return lipgloss.JoinVertical(
		lipgloss.Left,
		githubStyle.Render("GitHub: https://github.com/youkale/echogy"),
		descStyle.Render("Echogy - A lightweight and efficient SSH reverse proxy tool\n\n"+
			"Features:\n"+
			"• Easy to use HTTP/HTTPS tunnel\n"+
			"• Real-time traffic monitoring\n"+
			"• Beautiful TUI interface\n"+
			"• Cross-platform support\n\n"+
			"Scan QR code to access your tunnel →"),
	)
}

// View implements tea.Model
func (d *Dashboard) View() string {
	head := d.renderHeader()

	var content string
	if d.requests.Len() == 0 {
		// Show QR code and project info when table is empty
		qrCode := generateQRCode(d.tunnelInfo.URL)
		content = lipgloss.JoinHorizontal(
			lipgloss.Center,
			lipgloss.NewStyle().PaddingRight(4).Render(d.renderProjectInfo()),
			qrStyle.Render(qrCode),
		)
	} else {
		content = d.table.View()
	}

	return dashStyle.Render(
		lipgloss.JoinVertical(lipgloss.Left,
			headerStyle.Render(head),
			content),
	)
}

func getBytes(s string) int64 {
	atoi, err := strconv.ParseInt(s, 10, 64)
	if nil != err {
		return 0
	}
	return atoi
}

func renderStatusCode(status int) string {
	s := strconv.Itoa(status)
	style := lipgloss.NewStyle().Inherit(colStatusStyle)
	switch {
	case status >= 500:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#E53E3E", Dark: "#FC8181"}).Render(s) // Bright Red
	case status >= 400:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#DD6B20", Dark: "#FBD38D"}).Render(s) // Bright Orange
	case status >= 300:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#3182CE", Dark: "#90CDF4"}).Render(s) // Bright Blue
	case status >= 200:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#38A169", Dark: "#9AE6B4"}).Render(s) // Bright Green
	default:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#718096", Dark: "#A0AEC0"}).Render(s) // Cool Gray
	}
}

func renderMethod(method string) string {
	style := lipgloss.NewStyle().Inherit(colMethodHeaderStyle)
	switch method {
	case "GET":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#3182CE", Dark: "#90CDF4"}).Render(method) // Bright Blue
	case "POST":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#38A169", Dark: "#9AE6B4"}).Render(method) // Bright Green
	case "PUT":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#805AD5", Dark: "#B794F4"}).Render(method) // Bright Purple
	case "DELETE":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#E53E3E", Dark: "#FC8181"}).Render(method) // Bright Red
	case "PATCH":
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#DD6B20", Dark: "#FBD38D"}).Render(method) // Bright Orange
	default:
		return style.Foreground(lipgloss.AdaptiveColor{Light: "#718096", Dark: "#A0AEC0"}).Render(method) // Cool Gray
	}
}

// AddRequest adds a new request to the dashboard
func (d *Dashboard) AddRequest(req *httpExchange) {

	reqByte := getBytes(req.Request.Header.Get("Content-Length"))
	respByte := getBytes(req.Response.Header.Get("Content-Length"))

	d.tunnelInfo.BytesRecv += reqByte
	d.tunnelInfo.BytesSent += respByte

	d.tunnelInfo.ReqCount += 1
	d.tunnelInfo.ResCount += 1

	d.requests.Push(req)

	rows := make([]table.Row, d.requests.Len())
	// Update table rows
	for i, item := range d.requests.Items() {
		r := item.(*httpExchange)
		path := colPathStyle.Render(r.RequestURI)
		t := colUseTimeStyle.Render(humanMillis(r.useTime))

		rows[i] = table.Row{
			colNoStyle.Render(strconv.Itoa(i + 1)),
			renderMethod(r.Method),
			renderStatusCode(r.StatusCode),
			path,
			t,
		}
	}
	d.table.SetRows(rows)
}

// UpdateStats updates the tunnel information statistics
func (d *Dashboard) UpdateStats(info TunnelInfo) {
	d.tunnelInfo = info
}

// generateQRCode creates an ASCII QR code for the given URL
func generateQRCode(url string) string {
	qr, err := qrcode.New("https://"+url, qrcode.Low)
	if err != nil {
		return ""
	}
	return qr.ToString(true)
}
