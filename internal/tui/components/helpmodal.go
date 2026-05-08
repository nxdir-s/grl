package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	HelpModalTitle string = "Keybindings"
	HelpModalCmds  string = "esc/?: close · ↑/↓: scroll · pgup/pgdn: page"

	HelpModalWidth      int = 72
	HelpModalKeyCol     int = 18
	HelpModalPadding    int = 4
	HelpModalMaxRows    int = 18
	HelpModalChromeRows int = 8

	HelpSectionSidebar                 string = "Sidebar"
	HelpSectionRequestBuilderTabs      string = "Request Builder Tabs"
	HelpSectionHeadersParamsFormEditor string = "Headers / Params / Form Editor"
	HelpSectionBodyEditor              string = "Body Editor"
	HelpSectionAuthEditor              string = "Auth Editor"
	HelpSectionResponseViewer          string = "Response Viewer"
	HelpSectionSaveRequest             string = "Save Request Modal"
	HelpSectionEnvironments            string = "Environments Modal"
	HelpSectionSettings                string = "Settings Modal"
)

type HelpSection struct {
	Title    string
	Bindings []key.Binding
}

type HelpModalKeyMap struct {
	Close    key.Binding
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
}

func defaultHelpModalKeys() HelpModalKeyMap {
	return HelpModalKeyMap{
		Close:    key.NewBinding(key.WithKeys("esc", "?"), key.WithHelp("esc/?", "close")),
		Up:       key.NewBinding(key.WithKeys("up", "k"), key.WithHelp("↑/k", "scroll up")),
		Down:     key.NewBinding(key.WithKeys("down", "j"), key.WithHelp("↓/j", "scroll down")),
		PageUp:   key.NewBinding(key.WithKeys("pgup", "ctrl+u"), key.WithHelp("pgup", "page up")),
		PageDown: key.NewBinding(key.WithKeys("pgdown", "ctrl+d"), key.WithHelp("pgdn", "page down")),
	}
}

type HelpModalStyles struct {
	border   lipgloss.Style
	title    lipgloss.Style
	section  lipgloss.Style
	keyCol   lipgloss.Style
	descCol  lipgloss.Style
	cmdsHelp lipgloss.Style
}

func NewHelpModalStyles(width int) *HelpModalStyles {
	return &HelpModalStyles{
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(width),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
		section: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F5A623")),
		keyCol: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#7D56F4")).
			Bold(true),
		descCol: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
		cmdsHelp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
	}
}

type HelpModal struct {
	open     bool
	width    int
	height   int
	sections []HelpSection
	viewport viewport.Model
	ready    bool

	keys   HelpModalKeyMap
	styles *HelpModalStyles
}

func NewHelpModal() HelpModal {
	return HelpModal{
		keys:   defaultHelpModalKeys(),
		styles: NewHelpModalStyles(HelpModalWidth),
	}
}

func (m *HelpModal) Open() {
	m.open = true
	m.refreshContent()

	if m.ready {
		m.viewport.GotoTop()
	}
}

func (m *HelpModal) Close() {
	m.open = false
}

func (m HelpModal) IsOpen() bool {
	return m.open
}

func (m *HelpModal) SetSize(w, h int) {
	m.width = w
	m.height = h

	innerWidth := HelpModalWidth - HelpModalPadding

	innerHeight := HelpModalMaxRows
	if maxFit := h - HelpModalChromeRows; maxFit < innerHeight {
		innerHeight = maxFit
	}

	if innerHeight < 5 {
		innerHeight = 5
	}

	switch m.ready {
	case true:
		m.viewport.SetWidth(innerWidth)
		m.viewport.SetHeight(innerHeight)
	default:
		m.viewport = viewport.New(
			viewport.WithWidth(innerWidth),
			viewport.WithHeight(innerHeight),
		)

		m.ready = true
	}

	m.refreshContent()
}

func (m *HelpModal) SetSections(sections []HelpSection) {
	m.sections = sections
	m.refreshContent()
}

func (m *HelpModal) refreshContent() {
	if !m.ready {
		return
	}

	m.viewport.SetContent(m.renderSections())
}

func (m HelpModal) renderSections() string {
	var b strings.Builder

	for i := range m.sections {
		b.WriteString(m.styles.section.Render(m.sections[i].Title))
		b.WriteString("\n")

		for j := range m.sections[i].Bindings {
			if len(m.sections[i].Bindings[j].Help().Key) == 0 && len(m.sections[i].Bindings[j].Help().Desc) == 0 {
				continue
			}

			b.WriteString(fmt.Sprintf(
				"  %s  %s\n",
				m.styles.keyCol.Render(padRight(m.sections[i].Bindings[j].Help().Key, HelpModalKeyCol)),
				m.styles.descCol.Render(m.sections[i].Bindings[j].Help().Desc),
			))
		}

		if i < len(m.sections)-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

func padRight(s string, width int) string {
	rs := []rune(s)
	if len(rs) >= width {
		return s
	}

	return s + strings.Repeat(" ", width-len(rs))
}

func (m HelpModal) Update(msg tea.Msg) (HelpModal, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	if km, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(km, m.keys.Close):
			m.Close()
			return m, nil
		case key.Matches(km, m.keys.Up):
			m.viewport.ScrollUp(1)
			return m, nil
		case key.Matches(km, m.keys.Down):
			m.viewport.ScrollDown(1)
			return m, nil
		case key.Matches(km, m.keys.PageUp):
			m.viewport.PageUp()
			return m, nil
		case key.Matches(km, m.keys.PageDown):
			m.viewport.PageDown()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.viewport, cmd = m.viewport.Update(msg)

	return m, cmd
}

func (m HelpModal) View() string {
	if !m.open {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.styles.title.Render(HelpModalTitle))
	b.WriteString("\n\n")

	switch m.ready {
	case true:
		b.WriteString(m.viewport.View())
	default:
		b.WriteString(m.renderSections())
	}

	b.WriteString("\n\n")
	b.WriteString(m.styles.cmdsHelp.Render(HelpModalCmds))

	return m.styles.border.Render(b.String())
}

func BuildHelpSections(
	global HelpSection,
	sidebar Sidebar,
	builder RequestBuilder,
	response ResponseViewer,
	save SaveModal,
	env EnvModal,
	cfg ConfigModal,
) []HelpSection {
	return []HelpSection{
		global,
		{
			Title: HelpSectionSidebar,
			Bindings: []key.Binding{
				sidebar.keys.Up,
				sidebar.keys.Down,
				sidebar.keys.Select,
				sidebar.keys.Rename,
				sidebar.keys.Delete,
				sidebar.keys.Yes,
				sidebar.keys.No,
				sidebar.keys.Cancel,
			},
		},
		{
			Title: HelpSectionRequestBuilderTabs,
			Bindings: []key.Binding{
				builder.keys.NextTab,
				builder.keys.PrevTab,
			},
		},
		{
			Title: HelpSectionHeadersParamsFormEditor,
			Bindings: []key.Binding{
				builder.headers.keys.Up,
				builder.headers.keys.Down,
				builder.headers.keys.TabField,
				builder.headers.keys.AddRow,
				builder.headers.keys.DelRow,
				builder.headers.keys.Toggle,
			},
		},
		{
			Title: HelpSectionBodyEditor,
			Bindings: []key.Binding{
				builder.body.keys.CycleType,
			},
		},
		{
			Title: HelpSectionAuthEditor,
			Bindings: []key.Binding{
				builder.auth.keys.NextField,
				builder.auth.keys.PrevField,
				builder.auth.keys.CycleLeft,
				builder.auth.keys.CycleRight,
			},
		},
		{
			Title: HelpSectionResponseViewer,
			Bindings: []key.Binding{
				response.keys.NextTab,
				response.keys.PrevTab,
				response.keys.NextMatch,
				response.keys.PrevMatch,
				response.keys.CloseFind,
			},
		},
		{
			Title: HelpSectionSaveRequest,
			Bindings: []key.Binding{
				save.keys.Up,
				save.keys.Down,
				save.keys.Tab,
				save.keys.Select,
				save.keys.Close,
			},
		},
		{
			Title: HelpSectionEnvironments,
			Bindings: []key.Binding{
				env.keys.Up,
				env.keys.Down,
				env.keys.Select,
				env.keys.New,
				env.keys.Edit,
				env.keys.Delete,
				env.keys.Add,
				env.keys.Field,
				env.keys.Save,
				env.keys.Close,
			},
		},
		{
			Title: HelpSectionSettings,
			Bindings: []key.Binding{
				cfg.keys.Next,
				cfg.keys.Prev,
				cfg.keys.Toggle,
				cfg.keys.Save,
				cfg.keys.Close,
			},
		},
	}
}
