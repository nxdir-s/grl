package components

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

type ResponseHeaderPaneStyles struct {
	keys   lipgloss.Style
	values lipgloss.Style
}

func NewResponseHeaderPaneStyles() *ResponseHeaderPaneStyles {
	return &ResponseHeaderPaneStyles{
		keys:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
		values: lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")),
	}
}

type ResponseHeaderPane struct {
	viewport viewport.Model
	styles   *ResponseHeaderPaneStyles
	ready    bool
	raw      string
}

func NewResponseHeaderPane() ResponseHeaderPane {
	return ResponseHeaderPane{
		styles: NewResponseHeaderPaneStyles(),
	}
}

func (c *ResponseHeaderPane) SetSize(width, height int) {
	switch c.ready {
	case true:
		c.viewport.SetWidth(width)
		c.viewport.SetHeight(height)
	default:
		c.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		c.viewport.MouseWheelEnabled = true
		c.ready = true
	}
}

func (c *ResponseHeaderPane) SetHeaders(headers []valobj.Header) {
	if !c.ready {
		return
	}

	if len(headers) == 0 {
		c.viewport.SetContent("  No headers")
		return
	}

	// Sort headers by key for consistent display
	sorted := make([]valobj.Header, len(headers))
	copy(sorted, headers)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	var sb strings.Builder
	var rawSb strings.Builder

	for i := range sorted {
		sb.WriteString(fmt.Sprintf("  %s: %s\n",
			c.styles.keys.Render(sorted[i].Key),
			c.styles.values.Render(sorted[i].Value),
		))

		rawSb.WriteString(fmt.Sprintf("%s: %s\n", sorted[i].Key, sorted[i].Value))
	}

	c.raw = rawSb.String()

	c.viewport.SetContent(sb.String())
}

func (c ResponseHeaderPane) Raw() string {
	return c.raw
}

func (c *ResponseHeaderPane) Clear() {
	c.raw = ""
	if c.ready {
		c.viewport.SetContent("")
	}
}

func (c ResponseHeaderPane) Update(msg tea.Msg) (ResponseHeaderPane, tea.Cmd) {
	if !c.ready {
		return c, nil
	}

	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)

	return c, cmd
}

func (c ResponseHeaderPane) View() string {
	if !c.ready {
		return ""
	}

	return c.viewport.View()
}

type ResponseBodyPane struct {
	adapter  ports.TUI
	viewport viewport.Model
	ready    bool
	raw      string
}

func NewResponseBodyPane(adapter ports.TUI) ResponseBodyPane {
	return ResponseBodyPane{
		adapter: adapter,
	}
}

func (c *ResponseBodyPane) SetSize(width, height int) {
	switch c.ready {
	case true:
		c.viewport.SetWidth(width)
		c.viewport.SetHeight(height)
	default:
		c.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		c.viewport.MouseWheelEnabled = true
		c.ready = true
	}
}

func (c *ResponseBodyPane) SetContent(resp *entity.Response) {
	if !c.ready {
		return
	}

	c.raw = resp.FormatBody()

	c.viewport.SetContent(c.adapter.ColorizeJSON(c.raw))
}

func (c ResponseBodyPane) Raw() string {
	return c.raw
}

func (c *ResponseBodyPane) Clear() {
	c.raw = ""
	if c.ready {
		c.viewport.SetContent("")
	}
}

func (c ResponseBodyPane) Update(msg tea.Msg) (ResponseBodyPane, tea.Cmd) {
	if !c.ready {
		return c, nil
	}

	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)

	return c, cmd
}

func (c ResponseBodyPane) View() string {
	if !c.ready {
		return ""
	}

	return c.viewport.View()
}

type ResponseStatusLineStyles struct {
	muted     lipgloss.Style
	label     lipgloss.Style
	status2XX lipgloss.Style
	status3XX lipgloss.Style
	status4XX lipgloss.Style
	status5XX lipgloss.Style
}

func NewResponseStatusLineStyles() *ResponseStatusLineStyles {
	return &ResponseStatusLineStyles{
		muted:     lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
		label:     lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
		status2XX: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#73D216")),
		status3XX: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F5A623")),
		status4XX: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF6F61")),
		status5XX: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#FF4444")),
	}
}

type ResponseStatusLine struct {
	response *entity.Response
	styles   *ResponseStatusLineStyles
}

func NewResponseStatusLine() ResponseStatusLine {
	return ResponseStatusLine{
		styles: NewResponseStatusLineStyles(),
	}
}

func (c *ResponseStatusLine) SetResponse(resp *entity.Response) {
	c.response = resp
}

func (c *ResponseStatusLine) Clear() {
	c.response = nil
}

func (c ResponseStatusLine) View() string {
	if c.response == nil {
		return ""
	}

	status := c.renderStatus()
	totalTiming := c.styles.muted.Render(fmt.Sprintf("%dms", c.response.Timing.Total.Milliseconds()))
	size := c.styles.muted.Render(formatSize(len(c.response.Body.Bytes())))

	summary := fmt.Sprintf("%s  ·  %s  ·  %s", status, totalTiming, size)
	breakdown := c.formatTimingBreakdown(c.response.Timing)

	if len(breakdown) != 0 {
		summary += "\n" + breakdown
	}

	return summary
}

func (c ResponseStatusLine) renderStatus() string {
	switch {
	case c.response.StatusCode >= 200 && c.response.StatusCode < 300:
		return c.styles.status2XX.Render(c.response.Status)
	case c.response.StatusCode >= 300 && c.response.StatusCode < 400:
		return c.styles.status3XX.Render(c.response.Status)
	case c.response.StatusCode >= 400 && c.response.StatusCode < 500:
		return c.styles.status4XX.Render(c.response.Status)
	default:
		return c.styles.status5XX.Render(c.response.Status)
	}
}

func (c ResponseStatusLine) formatTimingBreakdown(t valobj.Timing) string {
	var parts []string

	if t.DNSLookup > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			c.styles.label.Render("DNS:"),
			c.styles.muted.Render(fmt.Sprintf("%dms", t.DNSLookup.Milliseconds())),
		))
	}

	if t.TCPConnect > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			c.styles.label.Render("TCP:"),
			c.styles.muted.Render(fmt.Sprintf("%dms", t.TCPConnect.Milliseconds())),
		))
	}

	if t.TLSHandshake > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			c.styles.label.Render("TLS:"),
			c.styles.muted.Render(fmt.Sprintf("%dms", t.TLSHandshake.Milliseconds())),
		))
	}

	if t.TTFB > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			c.styles.label.Render("TTFB:"),
			c.styles.muted.Render(fmt.Sprintf("%dms", t.TTFB.Milliseconds())),
		))
	}

	if len(parts) == 0 {
		return ""
	}

	return "  " + strings.Join(parts, "  ·  ")
}

func formatSize(bytes int) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/1024/1024)
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

type ResponseViewerKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
}

func defaultResponseViewerKeys() ResponseViewerKeyMap {
	return ResponseViewerKeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "prev tab"),
		),
	}
}

type ResponseViewer struct {
	statusLine ResponseStatusLine
	tabs       ResponseTabs
	bodyPane   ResponseBodyPane
	headerPane ResponseHeaderPane

	focused bool
	hasResp bool
	width   int
	height  int

	keys ResponseViewerKeyMap
}

func NewResponseViewer(adapter ports.TUI) ResponseViewer {
	return ResponseViewer{
		statusLine: NewResponseStatusLine(),
		tabs:       NewResponseTabs(),
		bodyPane:   NewResponseBodyPane(adapter),
		headerPane: NewResponseHeaderPane(),
		keys:       defaultResponseViewerKeys(),
	}
}

func (c *ResponseViewer) Focus() {
	c.focused = true
}

func (c *ResponseViewer) Blur() {
	c.focused = false
}

func (c *ResponseViewer) SetSize(width, height int) {
	c.width = width
	c.height = height

	// Reserve space for status line (2 lines) + tabs (1 line) + separator
	paneHeight := height - 4
	if paneHeight < 1 {
		paneHeight = 1
	}

	c.bodyPane.SetSize(width, paneHeight)
	c.headerPane.SetSize(width, paneHeight)
}

func (c *ResponseViewer) SetResponse(resp *entity.Response) {
	c.hasResp = true
	c.statusLine.SetResponse(resp)
	c.bodyPane.SetContent(resp)
	c.headerPane.SetHeaders(resp.Headers)
}

func (c *ResponseViewer) Clear() {
	c.hasResp = false
	c.statusLine.Clear()
	c.bodyPane.Clear()
	c.headerPane.Clear()
}

func (c ResponseViewer) HasResponse() bool {
	return c.hasResp
}

func (c ResponseViewer) CopyContent() string {
	if !c.hasResp {
		return ""
	}

	switch c.tabs.Active() {
	case ResponseTabHeaders:
		return c.headerPane.Raw()
	case ResponseTabBody:
		return c.bodyPane.Raw()
	default:
		return ""
	}
}

func (c ResponseViewer) Update(msg tea.Msg) (ResponseViewer, tea.Cmd) {
	if !c.focused || !c.hasResp {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keys.NextTab):
			c.tabs.Next()
			return c, nil
		case key.Matches(msg, c.keys.PrevTab):
			c.tabs.Prev()
			return c, nil
		}
	}

	// Route scroll events to active pane
	var cmd tea.Cmd
	switch c.tabs.Active() {
	case ResponseTabBody:
		c.bodyPane, cmd = c.bodyPane.Update(msg)
	case ResponseTabHeaders:
		c.headerPane, cmd = c.headerPane.Update(msg)
	}

	return c, cmd
}

func (c ResponseViewer) View() string {
	if !c.hasResp {
		return ""
	}

	statusView := c.statusLine.View()
	tabsView := c.tabs.View(c.width)

	var paneView string
	switch c.tabs.Active() {
	case ResponseTabBody:
		paneView = c.bodyPane.View()
	case ResponseTabHeaders:
		paneView = c.headerPane.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusView,
		tabsView,
		paneView,
	)
}
