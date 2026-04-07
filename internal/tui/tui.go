package tui

import (
	"context"
	"fmt"
	"os"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/ports"
	"github.com/nxdir-s/grl/internal/tui/components"
)

type KeyMap struct {
	Send        key.Binding
	CycleMethod key.Binding
	FocusNext   key.Binding
	Quit        key.Binding
}

func defaultKeyMap() KeyMap {
	return KeyMap{
		Send: key.NewBinding(
			key.WithKeys("ctrl+s"),
			key.WithHelp("ctrl+s", "send"),
		),
		CycleMethod: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "cycle method"),
		),
		FocusNext: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "switch panel"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c"),
			key.WithHelp("ctrl+c", "quit"),
		),
	}
}

func (m KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{
		m.Send,
		m.CycleMethod,
		m.FocusNext,
		m.Quit,
	}
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		[]key.Binding{
			m.Send,
			m.CycleMethod,
			m.FocusNext,
			m.Quit,
		},
	}
}

type FocusPanel int

const (
	FocusSidebar FocusPanel = iota
	FocusURLBar
	FocusRequestBuilder
	FocusResponseViewer
)

const (
	PanelCount   int = 4
	SidebarWidth int = 28
)

type TUIOpts func(t *TUI)

type TUI struct {
	adapter ports.CLI

	sidebar   components.Sidebar
	method    components.MethodSelector
	urlBar    components.URLBar
	builder   components.RequestBuilder
	response  components.ResponseViewer
	statusBar components.StatusBar

	focus  FocusPanel
	keys   KeyMap
	width  int
	height int

	panelCount   int
	sidebarWidth int

	styles Styles
	theme  Theme

	loading bool
	err     error
}

func New(adapter ports.CLI, opts ...TUIOpts) *TUI {
	theme := NewTheme()

	ui := &TUI{
		adapter:      adapter,
		sidebar:      components.NewSidebar(),
		method:       components.NewMethodSelector(),
		urlBar:       components.NewURLBar(),
		builder:      components.NewRequestBuilder(),
		response:     components.NewResponseViewer(),
		statusBar:    components.NewStatusBar(),
		focus:        FocusURLBar,
		keys:         defaultKeyMap(),
		panelCount:   PanelCount,
		sidebarWidth: SidebarWidth,
		styles:       NewStyles(theme),
		theme:        theme,
	}

	for _, opt := range opts {
		opt(ui)
	}

	return ui
}

func (t *TUI) Run(ctx context.Context) error {
	p := tea.NewProgram(t)

	if _, err := p.Run(); err != nil {
		return err
	}

	return nil
}

func (t *TUI) Init() tea.Cmd {
	return tea.Batch(
		t.urlBar.Focus(),
		t.loadSidebarData(),
	)
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.statusBar.SetWidth(t.width)
		t.recalcLayout()

		return t, nil
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, t.keys.Quit):
			return t, tea.Quit
		case key.Matches(msg, t.keys.CycleMethod):
			t.method.Next()
			return t, nil
		case key.Matches(msg, t.keys.Send):
			if t.loading {
				return t, nil
			}

			return t, t.sendRequest()
		case key.Matches(msg, t.keys.FocusNext):
			return t, t.cycleFocus(1)
		}
	case components.ResponseReceivedMsg:
		t.loading = false
		t.err = nil
		t.response.SetResponse(msg.Response)

		return t, t.recordHistory(msg.Request, msg.Response)
		// return t, nil
	case components.RequestErrorMsg:
		t.loading = false
		t.err = msg.Err
		t.response.Clear()

		return t, nil
	case components.HistoryUpdatedMsg:
		t.updateSidebarFromMsg(msg)
		return t, nil
	case components.CollectionsUpdatedMsg:
		t.updateSidebarCollections(msg)
		return t, nil
	case components.LoadRequestMsg:
		t.loadRequest(msg.Request)
		return t, nil
	case components.ErrorMsg:
		t.err = msg.Err
		return t, nil
	}

	// Route input to focused panel
	var cmd tea.Cmd
	switch t.focus {
	case FocusSidebar:
		t.sidebar, cmd = t.sidebar.Update(msg)
	case FocusURLBar:
		t.urlBar, cmd = t.urlBar.Update(msg)
	case FocusRequestBuilder:
		t.builder, cmd = t.builder.Update(msg)
	case FocusResponseViewer:
		t.response, cmd = t.response.Update(msg)
	}

	return t, cmd
}

func (t *TUI) View() tea.View {
	if t.width == 0 {
		v := tea.NewView("Initializing...")
		v.AltScreen = true

		return v
	}

	mainWidth := t.width - t.sidebarWidth - 1

	responseWidth := mainWidth / 2
	requestWidth := mainWidth - responseWidth - 1 // -1 for separator

	responseSection := t.responseSection()
	requestRow := t.requestRow()

	switch {
	case t.focus == FocusURLBar:
		requestRow = t.focusIndicator("▸ ") + requestRow
	default:
		requestRow = "  " + requestRow
	}

	builderView := t.requestBuilder()

	// Right column: request area (URL + builder stacked vertically)
	requestContent := lipgloss.JoinVertical(
		lipgloss.Left,
		requestRow,
		"",
		builderView,
	)

	// Pad columns to fill height
	helpView := t.styles.statusBar.Render(t.statusBar.View(t.keys))
	mainHeight := t.height - lipgloss.Height(helpView) - 1

	responseCol := lipgloss.NewStyle().
		Width(responseWidth).
		Height(mainHeight).
		Render(responseSection)

	requestCol := lipgloss.NewStyle().
		Width(requestWidth).
		Height(mainHeight).
		Render(requestContent)

	// Vertical separator between response and request columns
	colSeparator := lipgloss.NewStyle().
		Foreground(t.theme.colorBorder).
		Height(mainHeight).
		Render(strings.Repeat("│\n", mainHeight))

	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, requestCol, colSeparator, responseCol)

	// Sidebar
	sidebarView := t.sidebar.View()

	// Compose sidebar + main
	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainContent)

	view := lipgloss.JoinVertical(lipgloss.Left, body, helpView)

	v := tea.NewView(view)
	v.AltScreen = true

	return v
}

func (t *TUI) responseSection() string {
	// Response section (left column)
	var responseSection string

	switch {
	case t.loading:
		responseSection = t.styles.loading.Render("  Sending request...")
	case t.err != nil:
		responseSection = t.styles.error.Render(fmt.Sprintf("  Error: %s", t.err.Error()))
	case t.response.HasResponse():
		respView := t.response.View()

		switch t.focus == FocusResponseViewer {
		case true:
			responseSection = t.focusIndicator("▸ ") + respView
		case false:
			responseSection = "  " + respView
		}
	default:
		responseSection = lipgloss.NewStyle().
			Foreground(t.theme.colorMuted).
			Render("  Press ctrl+s to send a request")
	}

	return responseSection
}

// Request row: [METHOD] [URL input]
func (t *TUI) requestRow() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		t.method.View(),
		" ",
		t.urlBar.View(),
	)
}

func (t *TUI) requestBuilder() string {
	builderView := t.builder.View()

	switch {
	case t.focus == FocusRequestBuilder:
		builderView = t.focusIndicator("▸ ") + builderView
	default:
		builderView = "  " + builderView
	}

	return builderView
}

func (t *TUI) cycleFocus(dir int) tea.Cmd {
	t.blurCurrent()
	t.focus = FocusPanel((int(t.focus) + dir + t.panelCount) % t.panelCount)

	return t.focusCurrent()
}

func (t *TUI) blurCurrent() {
	switch t.focus {
	case FocusSidebar:
		t.sidebar.Blur()
	case FocusURLBar:
		t.urlBar.Blur()
	case FocusRequestBuilder:
		t.builder.Blur()
	case FocusResponseViewer:
		t.response.Blur()
	}
}

func (t *TUI) focusCurrent() tea.Cmd {
	switch t.focus {
	case FocusSidebar:
		t.sidebar.Focus()
		return nil
	case FocusURLBar:
		return t.urlBar.Focus()
	case FocusRequestBuilder:
		return t.builder.Focus()
	case FocusResponseViewer:
		t.response.Focus()
		return nil
	default:
		return nil
	}
}

func (t *TUI) recalcLayout() {
	mainWidth := t.width - t.sidebarWidth - 1
	helpHeight := 2
	responseWidth := mainWidth/2 - 4
	requestWidth := mainWidth - mainWidth/2 - 5 // account for separator + padding
	contentHeight := t.height - helpHeight - 1

	builderHeight := 10
	responseHeight := contentHeight - 2
	if responseHeight < 5 {
		responseHeight = 5
	}

	t.sidebar.SetSize(t.sidebarWidth, t.height-helpHeight)
	t.builder.SetSize(requestWidth, builderHeight)
	t.response.SetSize(responseWidth, responseHeight)
}

func (t *TUI) sendRequest() tea.Cmd {
	url := t.urlBar.Value()
	method := t.method.Current()

	t.loading = true
	t.err = nil
	t.response.Clear()

	req := &entity.Request{
		Method:  method,
		URL:     url,
		Headers: t.builder.GetHeaders(),
		Params:  t.builder.GetParams(),
		Body:    t.builder.GetBody(),
	}

	return func() tea.Msg {
		resp, err := t.adapter.SendRequest(context.Background(), req)
		if err != nil {
			return components.RequestErrorMsg{
				Err: err,
			}
		}

		return components.ResponseReceivedMsg{
			Response: resp,
			Request:  req,
		}
	}
}

func (t *TUI) recordHistory(req *entity.Request, resp *entity.Response) tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		if err := t.adapter.RecordHistory(ctx, req, resp); err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		history, err := t.adapter.GetHistory(ctx, 50)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.HistoryUpdatedMsg{
			History: history,
		}
	}
}

func (t *TUI) loadSidebarData() tea.Cmd {
	return func() tea.Msg {
		ctx := context.Background()

		history, err := t.adapter.GetHistory(ctx, 50)
		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to get history: %s\n", err.Error())
		}

		collections, err := t.adapter.ListCollections(ctx)
		if err != nil {
			fmt.Fprintf(os.Stdout, "failed to list collections: %s\n", err.Error())
		}

		// Use a batch to send both updates
		t.sidebar.SetData(history, collections)

		return components.HistoryUpdatedMsg{
			History: history,
		}
	}
}

func (t *TUI) updateSidebarFromMsg(msg components.HistoryUpdatedMsg) {
	ctx := context.Background()

	collections, err := t.adapter.ListCollections(ctx)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to list collections: %s\n", err.Error())
	}

	t.sidebar.SetData(msg.History, collections)
}

func (t *TUI) updateSidebarCollections(msg components.CollectionsUpdatedMsg) {
	ctx := context.Background()

	history, err := t.adapter.GetHistory(ctx, 50)
	if err != nil {
		fmt.Fprintf(os.Stdout, "failed to get history: %s\n", err.Error())
	}

	t.sidebar.SetData(history, msg.Collections)
}

func (t *TUI) loadRequest(req *entity.Request) {
	for req.Method != t.method.Current() {
		t.method.Next()
	}

	// Set URL — we need to re-create the URL bar with the value
	t.urlBar.SetValue(req.URL)
	// TODO: Load headers, params, body into the builder in future iteration
}

func (t *TUI) focusIndicator(s string) string {
	return lipgloss.NewStyle().
		Foreground(t.theme.colorPrimary).
		Bold(true).
		Render(s)
}
