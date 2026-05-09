package components

import (
	"fmt"
	"sort"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/muesli/reflow/wordwrap"
	"github.com/muesli/reflow/wrap"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
)

const (
	ResponseHeaderPlaceholder string = "  No headers"
	ResponseBodyIndent        int    = 2
)

type ResponseHeaderPaneStyles struct {
	keys                  lipgloss.Style
	values                lipgloss.Style
	matchHighlight        lipgloss.Style
	currentMatchHighlight lipgloss.Style
}

func NewResponseHeaderPaneStyles() *ResponseHeaderPaneStyles {
	return &ResponseHeaderPaneStyles{
		keys:   lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
		values: lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")),
		matchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("#3a3a00")).
			Foreground(lipgloss.Color("#FAFAFA")),
		currentMatchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true),
	}
}

type ResponseHeaderPane struct {
	viewport viewport.Model
	styles   *ResponseHeaderPaneStyles
	ready    bool
	raw      string
	styled   string
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
		c.viewport.HighlightStyle = c.styles.matchHighlight
		c.viewport.SelectedHighlightStyle = c.styles.currentMatchHighlight
		c.ready = true
	}
}

func (c *ResponseHeaderPane) SetHeaders(headers []valobj.Header) {
	if !c.ready {
		return
	}

	if len(headers) == 0 {
		c.raw = ""
		c.styled = ResponseHeaderPlaceholder
		c.viewport.SetContent(c.styled)
		return
	}

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
	c.styled = sb.String()

	c.viewport.SetContent(c.styled)
}

func (c ResponseHeaderPane) Raw() string {
	return c.raw
}

func (c *ResponseHeaderPane) ShowPlain() {
	if c.ready {
		c.viewport.SetContent(c.raw)
	}
}

func (c *ResponseHeaderPane) ShowStyled() {
	if c.ready {
		c.viewport.SetContent(c.styled)
	}
}

func (c *ResponseHeaderPane) Viewport() *viewport.Model {
	return &c.viewport
}

func (c *ResponseHeaderPane) Clear() {
	c.raw = ""
	c.styled = ""

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

type ResponseBodyPaneStyles struct {
	matchHighlight        lipgloss.Style
	currentMatchHighlight lipgloss.Style
}

func NewResponseBodyPaneStyles() *ResponseBodyPaneStyles {
	return &ResponseBodyPaneStyles{
		matchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("#3a3a00")).
			Foreground(lipgloss.Color("#FAFAFA")),
		currentMatchHighlight: lipgloss.NewStyle().
			Background(lipgloss.Color("#7D56F4")).
			Foreground(lipgloss.Color("#FAFAFA")).
			Bold(true),
	}
}

type ResponseBodyPane struct {
	adapter  ports.TUI
	styles   *ResponseBodyPaneStyles
	viewport viewport.Model
	ready    bool
	styled   bool
	width    int
	raw      string
}

func NewResponseBodyPane(adapter ports.TUI) ResponseBodyPane {
	return ResponseBodyPane{
		adapter: adapter,
		styles:  NewResponseBodyPaneStyles(),
		styled:  true,
	}
}

func (c *ResponseBodyPane) SetSize(width, height int) {
	widthChanged := width != c.width
	c.width = width

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
		c.viewport.HighlightStyle = c.styles.matchHighlight
		c.viewport.SelectedHighlightStyle = c.styles.currentMatchHighlight
		c.ready = true
	}

	if widthChanged && len(c.raw) != 0 {
		c.refresh()
	}
}

func (c *ResponseBodyPane) refresh() {
	content := c.raw
	if c.styled {
		content = c.adapter.ColorizeJSON(c.raw)
	}

	c.viewport.SetContent(c.wrapForViewport(content, c.width))
}

func (c *ResponseBodyPane) SetContent(resp *entity.Response) {
	if !c.ready {
		return
	}

	c.raw = resp.FormatBody()
	c.refresh()
}

func (c ResponseBodyPane) Raw() string {
	return c.raw
}

func (c *ResponseBodyPane) ShowPlain() {
	if c.ready {
		c.styled = false
		c.refresh()
	}
}

func (c *ResponseBodyPane) ShowStyled() {
	if c.ready {
		c.styled = true
		c.refresh()
	}
}

func (c *ResponseBodyPane) Viewport() *viewport.Model {
	return &c.viewport
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

func (c ResponseBodyPane) wrapForViewport(s string, width int) string {
	if width <= ResponseBodyIndent+1 {
		return s
	}

	inner := width - ResponseBodyIndent
	pad := strings.Repeat(" ", ResponseBodyIndent)

	var out strings.Builder
	out.Grow(len(s) + len(s)/inner*ResponseBodyIndent)

	lines := strings.Split(s, "\n")
	for i := range lines {
		wrapped := wrap.String(wordwrap.String(lines[i], inner), inner)
		parts := strings.Split(wrapped, "\n")

		for j := range parts {
			if j > 0 {
				out.WriteString("\n")
				out.WriteString(pad)
			}

			out.WriteString(parts[j])
		}

		if i < len(lines)-1 {
			out.WriteString("\n")
		}
	}

	return out.String()
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
	NextTab   key.Binding
	PrevTab   key.Binding
	NextMatch key.Binding
	PrevMatch key.Binding
	CloseFind key.Binding
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
		NextMatch: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "next match"),
		),
		PrevMatch: key.NewBinding(
			key.WithKeys("shift+enter"),
			key.WithHelp("shift+enter", "prev match"),
		),
		CloseFind: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close find"),
		),
	}
}

type ResponseSearch struct {
	active  bool
	input   textinput.Model
	pattern string
	total   int
	current int
}

func NewResponseSearch() ResponseSearch {
	input := textinput.New()
	input.Prompt = "/ "
	input.CharLimit = 256
	input.Placeholder = "find in response"

	return ResponseSearch{
		input: input,
	}
}

type ResponseViewerStyles struct {
	searchBar     lipgloss.Style
	searchCounter lipgloss.Style
}

func NewResponseViewerStyles() *ResponseViewerStyles {
	return &ResponseViewerStyles{
		searchBar: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")),
		searchCounter: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Italic(true),
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

	keys   ResponseViewerKeyMap
	search ResponseSearch

	styles *ResponseViewerStyles
}

func NewResponseViewer(adapter ports.TUI) ResponseViewer {
	return ResponseViewer{
		statusLine: NewResponseStatusLine(),
		tabs:       NewResponseTabs(),
		bodyPane:   NewResponseBodyPane(adapter),
		headerPane: NewResponseHeaderPane(),
		keys:       defaultResponseViewerKeys(),
		search:     NewResponseSearch(),
		styles:     NewResponseViewerStyles(),
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

	paneHeight := height - 4

	if c.search.active {
		paneHeight--
	}

	if paneHeight < 1 {
		paneHeight = 1
	}

	c.bodyPane.SetSize(width, paneHeight)
	c.headerPane.SetSize(width, paneHeight)
}

func (c *ResponseViewer) SetResponse(resp *entity.Response) {
	if c.search.active {
		c.CloseSearch()
	}

	c.hasResp = true
	c.statusLine.SetResponse(resp)
	c.bodyPane.SetContent(resp)
	c.headerPane.SetHeaders(resp.Headers)
}

func (c *ResponseViewer) Clear() {
	if c.search.active {
		c.CloseSearch()
	}

	c.hasResp = false
	c.statusLine.Clear()
	c.bodyPane.Clear()
	c.headerPane.Clear()
}

func (c *ResponseViewer) SearchActive() bool {
	return c.search.active
}

func (c *ResponseViewer) OpenSearch() tea.Cmd {
	if !c.hasResp || c.search.active {
		return nil
	}

	c.search.active = true
	c.search.input.SetValue("")
	c.search.pattern = ""
	c.search.total = 0
	c.search.current = 0

	c.showActivePlain()
	c.SetSize(c.width, c.height)

	return c.search.input.Focus()
}

func (c *ResponseViewer) CloseSearch() {
	if !c.search.active {
		return
	}

	c.search.active = false
	c.search.input.Blur()
	c.search.input.SetValue("")
	c.search.pattern = ""
	c.search.total = 0
	c.search.current = 0

	if viewport := c.activeViewport(); viewport != nil {
		viewport.ClearHighlights()
	}

	c.showActiveStyled()
	c.SetSize(c.width, c.height)
}

func (c *ResponseViewer) activeViewport() *viewport.Model {
	switch c.tabs.Active() {
	case ResponseTabBody:
		return c.bodyPane.Viewport()
	case ResponseTabHeaders:
		return c.headerPane.Viewport()
	default:
		return nil
	}
}

func (c *ResponseViewer) activePaneRaw() string {
	switch c.tabs.Active() {
	case ResponseTabBody:
		return c.bodyPane.Raw()
	case ResponseTabHeaders:
		return c.headerPane.Raw()
	default:
		return ""
	}
}

func (c *ResponseViewer) showActivePlain() {
	switch c.tabs.Active() {
	case ResponseTabBody:
		c.bodyPane.ShowPlain()
	case ResponseTabHeaders:
		c.headerPane.ShowPlain()
	}
}

func (c *ResponseViewer) showActiveStyled() {
	switch c.tabs.Active() {
	case ResponseTabBody:
		c.bodyPane.ShowStyled()
	case ResponseTabHeaders:
		c.headerPane.ShowStyled()
	}
}

func (c *ResponseViewer) recomputeMatches() {
	pattern := strings.ToLower(c.search.input.Value())
	c.search.pattern = pattern

	viewport := c.activeViewport()
	if viewport == nil {
		c.search.total = 0
		c.search.current = 0
		return
	}

	if pattern == "" {
		viewport.ClearHighlights()
		c.search.total = 0
		c.search.current = 0
		return
	}

	haystack := strings.ToLower(c.activePaneRaw())
	matches := [][]int{}

	for off := 0; off <= len(haystack); {
		idx := strings.Index(haystack[off:], pattern)
		if idx < 0 {
			break
		}

		start := off + idx
		end := start + len(pattern)

		matches = append(matches, []int{start, end})

		off = end
		if off == start {
			off++
		}
	}

	if len(matches) == 0 {
		viewport.ClearHighlights()
		c.search.total = 0
		c.search.current = 0
		return
	}

	viewport.SetHighlights(matches)

	c.search.total = len(matches)
	c.search.current = 1
}

func (c *ResponseViewer) nextMatch() {
	if c.search.total == 0 {
		return
	}

	if vp := c.activeViewport(); vp != nil {
		vp.HighlightNext()
	}

	c.search.current++

	if c.search.current > c.search.total {
		c.search.current = 1
	}
}

func (c *ResponseViewer) prevMatch() {
	if c.search.total == 0 {
		return
	}

	if vp := c.activeViewport(); vp != nil {
		vp.HighlightPrevious()
	}

	c.search.current--

	if c.search.current < 1 {
		c.search.current = c.search.total
	}
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
	if !c.hasResp {
		return c, nil
	}

	if _, isWheel := msg.(tea.MouseWheelMsg); isWheel {
		var cmd tea.Cmd
		switch c.tabs.Active() {
		case ResponseTabBody:
			c.bodyPane, cmd = c.bodyPane.Update(msg)
		case ResponseTabHeaders:
			c.headerPane, cmd = c.headerPane.Update(msg)
		}

		return c, cmd
	}

	if c.search.active {
		return c.updateSearch(msg)
	}

	if !c.focused {
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

	var cmd tea.Cmd
	switch c.tabs.Active() {
	case ResponseTabBody:
		c.bodyPane, cmd = c.bodyPane.Update(msg)
	case ResponseTabHeaders:
		c.headerPane, cmd = c.headerPane.Update(msg)
	}

	return c, cmd
}

func (c ResponseViewer) updateSearch(msg tea.Msg) (ResponseViewer, tea.Cmd) {
	km, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return c, nil
	}

	switch {
	case key.Matches(km, c.keys.CloseFind):
		c.CloseSearch()
		return c, nil
	case key.Matches(km, c.keys.NextMatch):
		c.nextMatch()
		return c, nil
	case key.Matches(km, c.keys.PrevMatch):
		c.prevMatch()
		return c, nil
	case key.Matches(km, c.keys.NextTab):
		c.showActiveStyled()
		c.tabs.Next()
		c.showActivePlain()
		c.recomputeMatches()
		return c, nil
	case key.Matches(km, c.keys.PrevTab):
		c.showActiveStyled()
		c.tabs.Prev()
		c.showActivePlain()
		c.recomputeMatches()
		return c, nil
	}

	prev := c.search.input.Value()

	var cmd tea.Cmd
	c.search.input, cmd = c.search.input.Update(msg)

	if c.search.input.Value() != prev {
		c.recomputeMatches()
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

	parts := []string{statusView, tabsView, paneView}
	if c.search.active {
		parts = append(parts, c.searchBarView())
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

func (c ResponseViewer) searchBarView() string {
	var counter string
	switch {
	case c.search.total > 0:
		counter = fmt.Sprintf(" (%d/%d)", c.search.current, c.search.total)
	case c.search.input.Value() != "":
		counter = " (no matches)"
	}

	return c.styles.searchBar.Render(c.search.input.View()) + c.styles.searchCounter.Render(counter)
}
