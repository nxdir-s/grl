package tui

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
}

func NewResponseHeaderPane() *ResponseHeaderPane {
	return &ResponseHeaderPane{}
}

func (h *ResponseHeaderPane) SetSize(width, height int) {
	if !h.ready {
		h.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		h.viewport.MouseWheelEnabled = true
		h.ready = true
	} else {
		h.viewport.SetWidth(width)
		h.viewport.SetHeight(height)
	}
}

func (h *ResponseHeaderPane) SetHeaders(headers []valobj.Header) {
	if !h.ready {
		return
	}

	if len(headers) == 0 {
		h.viewport.SetContent("  No headers")
		return
	}

	// Sort headers by key for consistent display
	sorted := make([]valobj.Header, len(headers))
	copy(sorted, headers)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Key < sorted[j].Key
	})

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4"))

	valStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA"))

	var sb strings.Builder
	for _, hdr := range sorted {
		sb.WriteString(fmt.Sprintf("  %s: %s\n",
			keyStyle.Render(hdr.Key),
			valStyle.Render(hdr.Value),
		))
	}

	h.viewport.SetContent(sb.String())
}

func (h *ResponseHeaderPane) Clear() {
	if h.ready {
		h.viewport.SetContent("")
	}
}

func (h *ResponseHeaderPane) Update(msg tea.Msg) tea.Cmd {
	if !h.ready {
		return nil
	}

	var cmd tea.Cmd
	h.viewport, cmd = h.viewport.Update(msg)

	return cmd
}

func (h *ResponseHeaderPane) View() string {
	if !h.ready {
		return ""
	}

	return h.viewport.View()
}

type ResponseBodyPane struct {
	viewport viewport.Model
	ready    bool
}

func NewResponseBodyPane() *ResponseBodyPane {
	return &ResponseBodyPane{}
}

func (b *ResponseBodyPane) SetSize(width, height int) {
	if !b.ready {
		b.viewport = viewport.New(
			viewport.WithWidth(width),
			viewport.WithHeight(height),
		)

		b.viewport.MouseWheelEnabled = true
		b.ready = true
	} else {
		b.viewport.SetWidth(width)
		b.viewport.SetHeight(height)
	}
}

func (b *ResponseBodyPane) SetContent(resp *entity.Response) {
	if !b.ready {
		return
	}

	b.viewport.SetContent(resp.FormatBody())
}

func (b *ResponseBodyPane) Clear() {
	if b.ready {
		b.viewport.SetContent("")
	}
}

func (b *ResponseBodyPane) Update(msg tea.Msg) tea.Cmd {
	if !b.ready {
		return nil
	}

	var cmd tea.Cmd
	b.viewport, cmd = b.viewport.Update(msg)

	return cmd
}

func (b *ResponseBodyPane) View() string {
	if !b.ready {
		return ""
	}

	return b.viewport.View()
}

type ResponseStatusLine struct {
	response *entity.Response
}

func NewResponseStatusLine() *ResponseStatusLine {
	return &ResponseStatusLine{}
}

func (s *ResponseStatusLine) SetResponse(resp *entity.Response) {
	s.response = resp
}

func (s *ResponseStatusLine) Clear() {
	s.response = nil
}

func (s *ResponseStatusLine) View() string {
	if s.response == nil {
		return ""
	}

	statusStyle := statusStyleForCode(s.response.StatusCode)
	status := statusStyle.Render(s.response.Status)

	mutedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#888888"))

	totalTiming := mutedStyle.Render(fmt.Sprintf("%dms", s.response.Timing.Total.Milliseconds()))
	size := mutedStyle.Render(formatSize(len(s.response.Body.Bytes())))

	summary := fmt.Sprintf("%s  ·  %s  ·  %s", status, totalTiming, size)

	breakdown := formatTimingBreakdown(s.response.Timing, labelStyle, mutedStyle)
	if breakdown != "" {
		summary += "\n" + breakdown
	}

	return summary
}

func formatTimingBreakdown(t valobj.Timing, labelStyle, valStyle lipgloss.Style) string {
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

func statusStyleForCode(code int) lipgloss.Style {
	style := lipgloss.NewStyle().Bold(true)
	switch {
	case code >= 200 && code < 300:
		return style.Foreground(lipgloss.Color("#73D216"))
	case code >= 300 && code < 400:
		return style.Foreground(lipgloss.Color("#F5A623"))
	case code >= 400 && code < 500:
		return style.Foreground(lipgloss.Color("#FF6F61"))
	default:
		return style.Foreground(lipgloss.Color("#FF4444"))
	}
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

type ResponseViewer struct {
	statusLine *ResponseStatusLine
	tabs       *Tabs
	bodyPane   *ResponseBodyPane
	headerPane *ResponseHeaderPane

	focused bool
	hasResp bool
	width   int
	height  int

	keys ResponseViewerKeyMap
}

type ResponseViewerKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
}

var defaultViewerKeys = ResponseViewerKeyMap{
	NextTab: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "prev tab"),
	),
}

func NewResponseViewer() *ResponseViewer {
	return &ResponseViewer{
		statusLine: NewResponseStatusLine(),
		tabs:       NewTabs(),
		bodyPane:   NewResponseBodyPane(),
		headerPane: NewResponseHeaderPane(),
		keys:       defaultViewerKeys,
	}
}

func (v *ResponseViewer) Focus() {
	v.focused = true
}

func (v *ResponseViewer) Blur() {
	v.focused = false
}

func (v *ResponseViewer) SetSize(width, height int) {
	v.width = width
	v.height = height

	// Reserve space for status line (2 lines) + tabs (1 line) + separator
	paneHeight := height - 4
	if paneHeight < 1 {
		paneHeight = 1
	}

	v.bodyPane.SetSize(width, paneHeight)
	v.headerPane.SetSize(width, paneHeight)
}

func (v *ResponseViewer) SetResponse(resp *entity.Response) {
	v.hasResp = true
	v.statusLine.SetResponse(resp)
	v.bodyPane.SetContent(resp)
	v.headerPane.SetHeaders(resp.Headers)
}

func (v *ResponseViewer) Clear() {
	v.hasResp = false
	v.statusLine.Clear()
	v.bodyPane.Clear()
	v.headerPane.Clear()
}

func (v *ResponseViewer) HasResponse() bool {
	return v.hasResp
}

func (v *ResponseViewer) Update(msg tea.Msg) tea.Cmd {
	if !v.focused || !v.hasResp {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, v.keys.NextTab):
			v.tabs.Next()
			return nil
		case key.Matches(msg, v.keys.PrevTab):
			v.tabs.Prev()
			return nil
		}
	}

	// Route scroll events to active pane
	var cmd tea.Cmd
	switch v.tabs.Active() {
	case TabBody:
		cmd = v.bodyPane.Update(msg)
	case TabHeaders:
		cmd = v.headerPane.Update(msg)
	}

	return cmd
}

func (v *ResponseViewer) View() string {
	if !v.hasResp {
		return ""
	}

	statusView := v.statusLine.View()
	tabsView := v.tabs.View(v.width)

	var paneView string
	switch v.tabs.Active() {
	case TabBody:
		paneView = v.bodyPane.View()
	case TabHeaders:
		paneView = v.headerPane.View()
	}

	return lipgloss.JoinVertical(
		lipgloss.Left,
		statusView,
		tabsView,
		paneView,
	)
}
