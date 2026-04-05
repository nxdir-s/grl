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
)

type ResponseHeaderPane struct {
	viewport viewport.Model
	ready    bool
	keyStyle lipgloss.Style
	valStyle lipgloss.Style
}

func NewResponseHeaderPane() *ResponseHeaderPane {
	return &ResponseHeaderPane{
		keyStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
		valStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")),
	}
}

func (c *ResponseHeaderPane) SetSize(width, height int) {
	if !c.ready {
		c.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		c.viewport.MouseWheelEnabled = true
		c.ready = true
	} else {
		c.viewport.SetWidth(width)
		c.viewport.SetHeight(height)
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
	for i := range sorted {
		sb.WriteString(fmt.Sprintf("  %s: %s\n",
			c.keyStyle.Render(sorted[i].Key),
			c.valStyle.Render(sorted[i].Value),
		))
	}

	c.viewport.SetContent(sb.String())
}

func (c *ResponseHeaderPane) Clear() {
	if c.ready {
		c.viewport.SetContent("")
	}
}

func (c *ResponseHeaderPane) Update(msg tea.Msg) tea.Cmd {
	if !c.ready {
		return nil
	}

	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)

	return cmd
}

func (c *ResponseHeaderPane) View() string {
	if !c.ready {
		return ""
	}

	return c.viewport.View()
}

type ResponseBodyPane struct {
	viewport viewport.Model
	ready    bool
}

func NewResponseBodyPane() *ResponseBodyPane {
	return &ResponseBodyPane{}
}

func (c *ResponseBodyPane) SetSize(width, height int) {
	if !c.ready {
		c.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		c.viewport.MouseWheelEnabled = true
		c.ready = true
	} else {
		c.viewport.SetWidth(width)
		c.viewport.SetHeight(height)
	}
}

func (c *ResponseBodyPane) SetContent(resp *entity.Response) {
	if !c.ready {
		return
	}

	c.viewport.SetContent(resp.FormatBody())
}

func (c *ResponseBodyPane) Clear() {
	if c.ready {
		c.viewport.SetContent("")
	}
}

func (c *ResponseBodyPane) Update(msg tea.Msg) tea.Cmd {
	if !c.ready {
		return nil
	}

	var cmd tea.Cmd
	c.viewport, cmd = c.viewport.Update(msg)

	return cmd
}

func (c *ResponseBodyPane) View() string {
	if !c.ready {
		return ""
	}

	return c.viewport.View()
}

type ResponseStatusLine struct {
	response   *entity.Response
	mutedStyle lipgloss.Style
	labelStyle lipgloss.Style
}

func NewResponseStatusLine() *ResponseStatusLine {
	return &ResponseStatusLine{
		mutedStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
		labelStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("#888888")),
	}
}

func (c *ResponseStatusLine) SetResponse(resp *entity.Response) {
	c.response = resp
}

func (c *ResponseStatusLine) Clear() {
	c.response = nil
}

func (c *ResponseStatusLine) View() string {
	if c.response == nil {
		return ""
	}

	status := c.styleForCode().Render(c.response.Status)
	totalTiming := c.mutedStyle.Render(fmt.Sprintf("%dms", c.response.Timing.Total.Milliseconds()))
	size := c.mutedStyle.Render(formatSize(len(c.response.Body.Bytes())))

	summary := fmt.Sprintf("%s  ·  %s  ·  %s", status, totalTiming, size)
	breakdown := formatTimingBreakdown(c.response.Timing, c.labelStyle, c.mutedStyle)

	if len(breakdown) != 0 {
		summary += "\n" + breakdown
	}

	return summary
}

func (c *ResponseStatusLine) styleForCode() lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)

	switch {
	case c.response.StatusCode >= 200 && c.response.StatusCode < 300:
		return style.Foreground(lipgloss.Color("#73D216"))
	case c.response.StatusCode >= 300 && c.response.StatusCode < 400:
		return style.Foreground(lipgloss.Color("#F5A623"))
	case c.response.StatusCode >= 400 && c.response.StatusCode < 500:
		return style.Foreground(lipgloss.Color("#FF6F61"))
	default:
		return style.Foreground(lipgloss.Color("#FF4444"))
	}
}

func formatTimingBreakdown(t valobj.Timing, labelStyle lipgloss.Style, valStyle lipgloss.Style) string {
	var parts []string

	if t.DNSLookup > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			labelStyle.Render("DNS:"),
			valStyle.Render(fmt.Sprintf("%dms", t.DNSLookup.Milliseconds())),
		))
	}

	if t.TCPConnect > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			labelStyle.Render("TCP:"),
			valStyle.Render(fmt.Sprintf("%dms", t.TCPConnect.Milliseconds())),
		))
	}

	if t.TLSHandshake > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			labelStyle.Render("TLS:"),
			valStyle.Render(fmt.Sprintf("%dms", t.TLSHandshake.Milliseconds())),
		))
	}

	if t.TTFB > 0 {
		parts = append(parts, fmt.Sprintf("%s %s",
			labelStyle.Render("TTFB:"),
			valStyle.Render(fmt.Sprintf("%dms", t.TTFB.Milliseconds())),
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
	statusLine *ResponseStatusLine
	tabs       *ResponseTabs
	bodyPane   *ResponseBodyPane
	headerPane *ResponseHeaderPane

	focused bool
	hasResp bool
	width   int
	height  int

	keys ResponseViewerKeyMap
}

func NewResponseViewer() *ResponseViewer {
	return &ResponseViewer{
		statusLine: NewResponseStatusLine(),
		tabs:       NewResponseTabs(),
		bodyPane:   NewResponseBodyPane(),
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

func (c *ResponseViewer) HasResponse() bool {
	return c.hasResp
}

func (c *ResponseViewer) Update(msg tea.Msg) tea.Cmd {
	if !c.focused || !c.hasResp {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keys.NextTab):
			c.tabs.Next()
			return nil
		case key.Matches(msg, c.keys.PrevTab):
			c.tabs.Prev()
			return nil
		}
	}

	// Route scroll events to active pane
	var cmd tea.Cmd
	switch c.tabs.Active() {
	case ResponseTabBody:
		cmd = c.bodyPane.Update(msg)
	case ResponseTabHeaders:
		cmd = c.headerPane.Update(msg)
	}

	return cmd
}

func (c *ResponseViewer) View() string {
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
