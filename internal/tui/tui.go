package tui

import (
	"context"
	"strings"
	"time"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nxdir-s/grl/internal/core/entity"
	"github.com/nxdir-s/grl/internal/core/valobj"
	"github.com/nxdir-s/grl/internal/ports"
	"github.com/nxdir-s/grl/internal/tui/components"
)

type KeyMap struct {
	Send        key.Binding
	CycleMethod key.Binding
	FocusNext   key.Binding
	EnvModal    key.Binding
	ConfigModal key.Binding
	Copy        key.Binding
	Search      key.Binding
	SaveModal   key.Binding
	Help        key.Binding
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
		EnvModal: key.NewBinding(
			key.WithKeys("ctrl+e"),
			key.WithHelp("ctrl+e", "environments"),
		),
		ConfigModal: key.NewBinding(
			key.WithKeys("ctrl+t"),
			key.WithHelp("ctrl+t", "settings"),
		),
		Copy: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "copy"),
		),
		Search: key.NewBinding(
			key.WithKeys("ctrl+f"),
			key.WithHelp("ctrl+f", "find"),
		),
		SaveModal: key.NewBinding(
			key.WithKeys("ctrl+w"),
			key.WithHelp("ctrl+w", "save request"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
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
		m.EnvModal,
		m.ConfigModal,
		m.Copy,
		m.Search,
		m.SaveModal,
		m.Help,
		m.Quit,
	}
}

func (m KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		[]key.Binding{
			m.Send,
			m.CycleMethod,
			m.FocusNext,
			m.EnvModal,
			m.ConfigModal,
			m.Copy,
			m.Search,
			m.SaveModal,
			m.Help,
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

type Rect struct {
	x, y, w, h int
}

func (r Rect) contains(x, y int) bool {
	return x >= r.x && x < r.x+r.w && y >= r.y && y < r.y+r.h
}

type TUIOpts func(t *TUI)

func WithContext(ctx context.Context) TUIOpts {
	return func(t *TUI) {
		t.ctx = ctx
	}
}

type TUI struct {
	ctx     context.Context
	adapter ports.TUI

	sidebar     components.Sidebar
	method      components.MethodSelector
	urlBar      components.URLBar
	builder     components.RequestBuilder
	response    components.ResponseViewer
	statusBar   components.StatusBar
	envModal    components.EnvModal
	configModal components.ConfigModal
	saveModal   components.SaveModal
	helpModal   components.HelpModal

	collections []entity.Collection

	focus  FocusPanel
	keys   KeyMap
	width  int
	height int

	sidebarRect  Rect
	urlBarRect   Rect
	builderRect  Rect
	responseRect Rect

	panelCount   int
	sidebarWidth int

	styles *Styles
	themes *Themes

	activeEnv *entity.Environment

	loading  bool
	err      error
	flashMsg string
}

func New(adapter ports.TUI, opts ...TUIOpts) *TUI {
	themes := NewThemes()

	ui := &TUI{
		ctx:          context.Background(),
		adapter:      adapter,
		sidebar:      components.NewSidebar(),
		method:       components.NewMethodSelector(),
		urlBar:       components.NewURLBar(),
		builder:      components.NewRequestBuilder(),
		response:     components.NewResponseViewer(adapter),
		statusBar:    components.NewStatusBar(),
		envModal:     components.NewEnvModal(),
		configModal:  components.NewConfigModal(),
		saveModal:    components.NewSaveModal(),
		helpModal:    components.NewHelpModal(),
		collections:  make([]entity.Collection, 0),
		focus:        FocusURLBar,
		keys:         defaultKeyMap(),
		panelCount:   PanelCount,
		sidebarWidth: SidebarWidth,
		styles:       NewStyles(themes),
		themes:       themes,
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
		t.loadEnvironments(),
		t.loadConfig(),
	)
}

func (t *TUI) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		t.width = msg.Width
		t.height = msg.Height
		t.statusBar.SetWidth(t.width)
		t.envModal.SetSize(t.width, t.height)
		t.helpModal.SetSize(t.width, t.height)
		t.recalcLayout()

		return t, nil
	case tea.MouseClickMsg:
		if t.modalOpen() {
			return t, nil
		}

		return t, t.handleMouseClick(msg.X, msg.Y)
	case tea.MouseWheelMsg:
		if t.modalOpen() {
			return t, nil
		}

		return t, t.handleMouseWheel(msg)
	case tea.KeyPressMsg:
		if t.helpModal.IsOpen() {
			var cmd tea.Cmd
			t.helpModal, cmd = t.helpModal.Update(msg)
			return t, cmd
		}

		if t.envModal.IsOpen() {
			var cmd tea.Cmd
			t.envModal, cmd = t.envModal.Update(msg)
			return t, cmd
		}

		if t.configModal.IsOpen() {
			var cmd tea.Cmd
			t.configModal, cmd = t.configModal.Update(msg)
			return t, cmd
		}

		if t.saveModal.IsOpen() {
			var cmd tea.Cmd
			t.saveModal, cmd = t.saveModal.Update(msg)
			return t, cmd
		}

		if key.Matches(msg, t.keys.Quit) {
			return t, tea.Quit
		}

		if t.response.SearchActive() {
			var cmd tea.Cmd
			t.response, cmd = t.response.Update(msg)
			return t, cmd
		}

		switch {
		case key.Matches(msg, t.keys.EnvModal):
			t.envModal.Open()
			return t, t.loadEnvironments()
		case key.Matches(msg, t.keys.ConfigModal):
			return t, t.openConfigModal()
		case key.Matches(msg, t.keys.SaveModal):
			return t, t.openSaveModal()
		case key.Matches(msg, t.keys.Help):
			t.openHelpModal()
			return t, nil
		case key.Matches(msg, t.keys.CycleMethod):
			t.method.Next()
			return t, nil
		case key.Matches(msg, t.keys.Send):
			if t.loading {
				return t, nil
			}

			return t, t.sendRequest()
		case key.Matches(msg, t.keys.Copy) && t.focus == FocusResponseViewer:
			return t, t.copyResponse()
		case key.Matches(msg, t.keys.Search):
			return t, t.openSearch()
		case key.Matches(msg, t.keys.FocusNext):
			return t, t.cycleFocus(1)
		}
	case components.ResponseReceivedMsg:
		t.loading = false
		t.err = nil
		t.response.SetResponse(msg.Response)

		return t, t.recordHistory(msg.Request, msg.Response)
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
	case components.EnvironmentsUpdatedMsg:
		t.activeEnv = msg.Active
		activeID := ""
		if msg.Active != nil {
			activeID = msg.Active.ID
		}

		t.envModal.SetData(msg.Environments, activeID)

		return t, nil
	case components.EnvironmentSwitchedMsg:
		t.activeEnv = msg.Active
		return t, t.loadEnvironments()
	case components.ActivateEnvMsg:
		return t, t.activateEnv(msg.ID)
	case components.CreateEnvMsg:
		return t, t.createEnv(msg.Name)
	case components.DeleteEnvMsg:
		return t, t.deleteEnv(msg.ID)
	case components.SaveEnvMsg:
		return t, t.saveEnv(msg.Env)
	case components.SaveConfigMsg:
		return t, t.saveConfig(msg.Cfg)
	case components.ConfigUpdatedMsg:
		t.applyConfig(msg.Cfg)
		return t, nil
	case components.LoadRequestMsg:
		t.loadRequest(msg.Request)
		return t, nil
	case components.RenameCollectionMsg:
		return t, t.renameCollection(msg.ID, msg.NewName)
	case components.DeleteCollectionMsg:
		return t, t.deleteCollectionCmd(msg.ID)
	case components.RemoveRequestMsg:
		return t, t.removeRequestCmd(msg.CollectionID, msg.RequestID)
	case components.DeleteHistoryEntryMsg:
		return t, t.deleteHistoryEntryCmd(msg.EntryID)
	case components.FlashMsg:
		t.flashMsg = msg.Text
		return t, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return components.ClearFlashMsg{}
		})
	case components.ClearFlashMsg:
		t.flashMsg = ""
		return t, nil
	case components.ErrorMsg:
		t.err = msg.Err
		return t, nil
	case components.SaveToCollectionMsg:
		return t, t.saveRequestToCollection(msg.CollectionID, msg.RequestName)

	case components.CreateAndSaveMsg:
		return t, t.createAndSaveRequest(msg.CollectionName, msg.RequestName)

	case components.CloseSaveModalMsg:
		return t, nil

	case components.RequestSavedMsg:
		t.updateSidebarCollections(components.CollectionsUpdatedMsg{
			Collections: msg.Collections,
		})

		t.flashMsg = msg.FlashText

		return t, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
			return components.ClearFlashMsg{}
		})
	}

	if t.envModal.IsOpen() {
		var cmd tea.Cmd
		t.envModal, cmd = t.envModal.Update(msg)
		return t, cmd
	}

	if t.configModal.IsOpen() {
		var cmd tea.Cmd
		t.configModal, cmd = t.configModal.Update(msg)
		return t, cmd
	}

	if t.saveModal.IsOpen() {
		var cmd tea.Cmd
		t.saveModal, cmd = t.saveModal.Update(msg)
		return t, cmd
	}

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

	badge := t.envBadge()
	if len(t.flashMsg) != 0 {
		badge = t.styles.flash.Render(t.flashMsg) + "  " + badge
	}

	helpView := t.styles.statusBar.Render(t.statusBar.View(t.keys) + "  " + badge)
	mainHeight := t.height - lipgloss.Height(helpView) - 1

	mainWidth := t.width - t.sidebarWidth
	responseWidth := mainWidth / 2
	requestWidth := mainWidth - responseWidth

	urlOuterH := 3
	builderOuterH := mainHeight - urlOuterH

	sidebarView := t.panelStyle(FocusSidebar).
		Width(t.sidebarWidth).
		Height(mainHeight).
		Render(t.sidebar.View())

	urlPanel := t.panelStyle(FocusURLBar).
		Width(requestWidth).
		Height(urlOuterH).
		Render(t.requestRow())

	builderPanel := t.panelStyle(FocusRequestBuilder).
		Width(requestWidth).
		Height(builderOuterH).
		Render(t.requestBuilder())

	responsePanel := t.panelStyle(FocusResponseViewer).
		Width(responseWidth).
		Height(mainHeight).
		Render(t.responseSection())

	requestCol := lipgloss.JoinVertical(lipgloss.Left, urlPanel, builderPanel)
	mainContent := lipgloss.JoinHorizontal(lipgloss.Top, requestCol, responsePanel)

	body := lipgloss.JoinHorizontal(lipgloss.Top, sidebarView, mainContent)

	switch {
	case t.envModal.IsOpen():
		body = lipgloss.Place(
			t.width, mainHeight,
			lipgloss.Center, lipgloss.Center,
			t.envModal.View(),
		)
	case t.configModal.IsOpen():
		body = lipgloss.Place(
			t.width, mainHeight,
			lipgloss.Center, lipgloss.Center,
			t.configModal.View(),
		)
	case t.saveModal.IsOpen():
		body = lipgloss.Place(
			t.width, mainHeight,
			lipgloss.Center, lipgloss.Center,
			t.saveModal.View(),
		)
	case t.helpModal.IsOpen():
		body = lipgloss.Place(
			t.width, mainHeight,
			lipgloss.Center, lipgloss.Center,
			t.helpModal.View(),
		)
	}

	view := lipgloss.JoinVertical(lipgloss.Left, body, helpView)

	v := tea.NewView(view)
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion

	return v
}

func (t *TUI) openSearch() tea.Cmd {
	if !t.response.HasResponse() {
		return func() tea.Msg {
			return components.FlashMsg{
				Text: "no response to search",
			}
		}
	}

	if t.focus != FocusResponseViewer {
		t.blurCurrent()
		t.focus = FocusResponseViewer
		t.focusCurrent()
	}

	return t.response.OpenSearch()
}

func (t *TUI) copyResponse() tea.Cmd {
	content := t.response.CopyContent()
	if len(content) == 0 {
		return nil
	}

	return func() tea.Msg {
		if err := t.adapter.CopyToClipboard(content); err != nil {
			return components.FlashMsg{
				Text: "copy failed: " + err.Error(),
			}
		}

		return components.FlashMsg{
			Text: "copied",
		}
	}
}

func (t *TUI) envBadge() string {
	name := "none"

	if t.activeEnv != nil {
		name = t.activeEnv.Name
	}

	return t.styles.envBadge.Render("env: " + name)
}

func (t *TUI) handleMouseClick(x, y int) tea.Cmd {
	var target FocusPanel

	switch {
	case t.sidebarRect.contains(x, y):
		target = FocusSidebar
	case t.urlBarRect.contains(x, y):
		target = FocusURLBar
	case t.builderRect.contains(x, y):
		target = FocusRequestBuilder
	case t.responseRect.contains(x, y):
		target = FocusResponseViewer
	default:
		return nil
	}

	if target == t.focus {
		return nil
	}

	t.blurCurrent()
	t.focus = target

	return t.focusCurrent()
}

func (t *TUI) handleMouseWheel(msg tea.MouseWheelMsg) tea.Cmd {
	x, y := msg.X, msg.Y

	var cmd tea.Cmd
	switch {
	case t.responseRect.contains(x, y):
		t.response, cmd = t.response.Update(msg)
	case t.builderRect.contains(x, y):
		t.builder, cmd = t.builder.Update(msg)
	case t.sidebarRect.contains(x, y):
		t.sidebar, cmd = t.sidebar.Update(msg)
	}

	return cmd
}

func (t *TUI) panelStyle(panel FocusPanel) lipgloss.Style {
	color := t.themes.colorBorder
	if t.focus == panel {
		color = t.themes.colorPrimary
	}

	return t.styles.panelBorder.BorderForeground(color)
}

func (t *TUI) responseSection() string {
	switch {
	case t.loading:
		return t.styles.loading.Render("Sending request...")
	case t.err != nil:
		return t.styles.error.Render("Error: " + t.err.Error())
	case t.response.HasResponse():
		return t.response.View()
	default:
		return t.styles.responseSection.Render("Press ctrl+s to send a request")
	}
}

func (t *TUI) requestRow() string {
	return lipgloss.JoinHorizontal(
		lipgloss.Center,
		t.method.View(),
		" ",
		t.urlBar.View(),
	)
}

func (t *TUI) requestBuilder() string {
	return t.builder.View()
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
	helpHeight := 1
	mainHeight := t.height - helpHeight - 1

	mainWidth := t.width - t.sidebarWidth
	responseWidth := mainWidth / 2
	requestWidth := mainWidth - responseWidth

	urlOuterH := 3
	builderOuterH := mainHeight - urlOuterH

	t.sidebar.SetSize(t.sidebarWidth-2, mainHeight-2)
	t.builder.SetSize(requestWidth-2, builderOuterH-2)
	t.response.SetSize(responseWidth-2, mainHeight-2)

	t.sidebarRect = Rect{
		x: 0,
		y: 0,
		w: t.sidebarWidth,
		h: mainHeight,
	}

	t.urlBarRect = Rect{
		x: t.sidebarWidth,
		y: 0,
		w: requestWidth,
		h: urlOuterH,
	}

	t.builderRect = Rect{
		x: t.sidebarWidth,
		y: urlOuterH,
		w: requestWidth,
		h: builderOuterH,
	}

	t.responseRect = Rect{
		x: t.sidebarWidth + requestWidth,
		y: 0,
		w: responseWidth,
		h: mainHeight,
	}
}

func (t *TUI) modalOpen() bool {
	if t.envModal.IsOpen() || t.configModal.IsOpen() || t.saveModal.IsOpen() || t.helpModal.IsOpen() {
		return true
	}

	return false
}

func (t *TUI) sendRequest() tea.Cmd {
	url := t.urlBar.Value()
	method := t.method.Current()

	t.loading = true
	t.err = nil
	t.response.Clear()

	req := &entity.Request{
		Method:     method,
		URL:        url,
		Headers:    t.builder.GetHeaders(),
		Params:     t.builder.GetParams(),
		Body:       t.builder.GetBody(),
		BodyType:   t.builder.GetBodyType(),
		FormFields: t.builder.GetFormFields(),
		Auth:       t.builder.GetAuth(),
	}

	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		resp, err := t.adapter.SendRequest(ctx, req)
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
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

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
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		history, err := t.adapter.GetHistory(ctx, 50)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		collections, err := t.adapter.ListCollections(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		t.sidebar.SetData(history, collections)

		return components.HistoryUpdatedMsg{
			History: history,
		}
	}
}

func (t *TUI) loadConfig() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		return components.ConfigUpdatedMsg{
			Cfg: t.adapter.GetConfig(ctx),
		}
	}
}

func (t *TUI) openConfigModal() tea.Cmd {
	ctx, cancel := context.WithCancel(t.ctx)
	defer cancel()

	return t.configModal.Open(t.adapter.GetConfig(ctx))
}

func (t *TUI) saveConfig(cfg *valobj.Config) tea.Cmd {
	return tea.Batch(
		func() tea.Msg {
			ctx, cancel := context.WithCancel(t.ctx)
			defer cancel()

			if err := t.adapter.SaveConfig(ctx, cfg); err != nil {
				return components.ErrorMsg{
					Err: err,
				}
			}

			return components.ConfigUpdatedMsg{
				Cfg: cfg,
			}
		},
		func() tea.Msg {
			return components.FlashMsg{
				Text: components.ConfigSavedAlert,
			}
		},
	)
}

func (t *TUI) applyConfig(cfg *valobj.Config) {
	if cfg == nil {
		return
	}

	t.method.SetCurrent(valobj.HTTPMethod(cfg.DefaultMethod))
}

func (t *TUI) loadEnvironments() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		envs, err := t.adapter.ListEnvironments(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		active, err := t.adapter.GetActiveEnvironment(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.EnvironmentsUpdatedMsg{
			Environments: envs,
			Active:       active,
		}
	}
}

func (t *TUI) activateEnv(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		if err := t.adapter.SetActiveEnvironment(ctx, id); err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		active, err := t.adapter.GetActiveEnvironment(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		envs, err := t.adapter.ListEnvironments(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.EnvironmentsUpdatedMsg{
			Environments: envs,
			Active:       active,
		}
	}
}

func (t *TUI) createEnv(name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		if _, err := t.adapter.CreateEnvironment(ctx, name); err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		envs, err := t.adapter.ListEnvironments(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		active, err := t.adapter.GetActiveEnvironment(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.EnvironmentsUpdatedMsg{
			Environments: envs,
			Active:       active,
		}
	}
}

func (t *TUI) deleteEnv(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		if err := t.adapter.DeleteEnvironment(ctx, id); err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		envs, err := t.adapter.ListEnvironments(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		active, err := t.adapter.GetActiveEnvironment(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.EnvironmentsUpdatedMsg{
			Environments: envs,
			Active:       active,
		}
	}
}

func (t *TUI) saveEnv(env *entity.Environment) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		if err := t.adapter.SaveEnvironment(ctx, env); err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		envs, err := t.adapter.ListEnvironments(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		active, err := t.adapter.GetActiveEnvironment(ctx)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.EnvironmentsUpdatedMsg{
			Environments: envs,
			Active:       active,
		}
	}
}

func (t *TUI) updateSidebarFromMsg(msg components.HistoryUpdatedMsg) {
	ctx, cancel := context.WithCancel(t.ctx)
	defer cancel()

	collections, err := t.adapter.ListCollections(ctx)
	if err != nil {
		collections = make([]entity.Collection, 0)
	}

	t.collections = collections

	t.sidebar.SetData(msg.History, collections)
}

func (t *TUI) updateSidebarCollections(msg components.CollectionsUpdatedMsg) {
	ctx, cancel := context.WithCancel(t.ctx)
	defer cancel()

	history, err := t.adapter.GetHistory(ctx, 50)
	if err != nil {
		history = make([]entity.HistoryEntry, 0)
	}

	t.collections = msg.Collections

	t.sidebar.SetData(history, msg.Collections)
}

func (t *TUI) loadRequest(req *entity.Request) {
	t.method.SetCurrent(req.Method)
	t.urlBar.SetValue(req.URL)
	t.builder.SetHeaders(req.Headers)
	t.builder.SetParams(req.Params)
	t.builder.SetBody(req.Body)
	t.builder.SetBodyType(req.BodyType)
	t.builder.SetFormFields(req.FormFields)
	t.builder.SetAuth(req.Auth)
}

func (t *TUI) openSaveModal() tea.Cmd {
	defaultName := t.method.Current().String() + " " + t.urlBar.Value()
	defaultName = strings.TrimSpace(defaultName)

	return t.saveModal.Open(t.collections, defaultName)
}

func (t *TUI) openHelpModal() {
	global := components.HelpSection{
		Title:    "Global",
		Bindings: t.keys.ShortHelp(),
	}

	t.helpModal.SetSections(components.BuildHelpSections(
		global,
		t.sidebar,
		t.builder,
		t.response,
		t.saveModal,
		t.envModal,
		t.configModal,
	))

	t.helpModal.Open()
}

func (t *TUI) snapshotRequest(name string) *entity.Request {
	return &entity.Request{
		Name:       name,
		Method:     t.method.Current(),
		URL:        t.urlBar.Value(),
		Headers:    t.builder.GetHeaders(),
		Params:     t.builder.GetParams(),
		Body:       t.builder.GetBody(),
		BodyType:   t.builder.GetBodyType(),
		FormFields: t.builder.GetFormFields(),
		Auth:       t.builder.GetAuth(),
	}
}

func (t *TUI) saveRequestToCollection(collectionID string, requestName string) tea.Cmd {
	req := t.snapshotRequest(requestName)

	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		collections, err := t.adapter.SaveRequestToCollection(ctx, req, requestName, collectionID)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.RequestSavedMsg{
			Collections: collections,
			FlashText:   "saved to collection",
		}
	}
}

func (t *TUI) createAndSaveRequest(collectionName, requestName string) tea.Cmd {
	req := t.snapshotRequest(requestName)

	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		collections, err := t.adapter.CreateCollection(ctx, collectionName, req)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.RequestSavedMsg{
			Collections: collections,
			FlashText:   "saved to new collection",
		}
	}
}

func (t *TUI) renameCollection(id string, name string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		collections, err := t.adapter.RenameCollection(ctx, id, name)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.RequestSavedMsg{
			Collections: collections,
			FlashText:   "renamed",
		}
	}
}

func (t *TUI) deleteCollectionCmd(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		collections, err := t.adapter.DeleteCollection(ctx, id)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.RequestSavedMsg{
			Collections: collections,
			FlashText:   "collection deleted",
		}
	}
}

func (t *TUI) removeRequestCmd(collectionID string, requestID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		collections, err := t.adapter.RemoveRequest(ctx, collectionID, requestID)
		if err != nil {
			return components.ErrorMsg{
				Err: err,
			}
		}

		return components.RequestSavedMsg{
			Collections: collections,
			FlashText:   "request removed",
		}
	}
}

func (t *TUI) deleteHistoryEntryCmd(entryID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithCancel(t.ctx)
		defer cancel()

		if err := t.adapter.DeleteHistoryEntry(ctx, entryID); err != nil {
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
