package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/entity"
)

type EnvMode int

const (
	EnvModeList EnvMode = iota
	EnvModeCreate
	EnvModeEdit
)

type EnvEditField int

const (
	EnvEditKey EnvEditField = iota
	EnvEditValue
)

type EnvModalKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	New    key.Binding
	Edit   key.Binding
	Delete key.Binding
	Close  key.Binding
	Save   key.Binding
	Add    key.Binding
	Field  key.Binding
}

func defaultEnvModalKeys() EnvModalKeyMap {
	return EnvModalKeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "ctrl+p"), key.WithHelp("↑", "up")),
		Down:   key.NewBinding(key.WithKeys("down", "ctrl+n"), key.WithHelp("↓", "down")),
		Select: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "activate")),
		New:    key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "new env")),
		Edit:   key.NewBinding(key.WithKeys("ctrl+y"), key.WithHelp("ctrl+y", "edit vars")),
		Delete: key.NewBinding(key.WithKeys("ctrl+g"), key.WithHelp("ctrl+g", "delete")),
		Close:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "close")),
		Save:   key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Add:    key.NewBinding(key.WithKeys("ctrl+o"), key.WithHelp("ctrl+o", "add var")),
		Field:  key.NewBinding(key.WithKeys("tab"), key.WithHelp("tab", "next field")),
	}
}

type EnvModalStyles struct {
	cmdsHelp lipgloss.Style
	sel      lipgloss.Style
	muted    lipgloss.Style
	active   lipgloss.Style
	key      lipgloss.Style
	noVars   lipgloss.Style
	noEnv    lipgloss.Style
	border   lipgloss.Style
	title    lipgloss.Style
}

func NewEnvModalStyles() *EnvModalStyles {
	return &EnvModalStyles{
		cmdsHelp: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		sel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")),
		muted: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#AAAAAA")),
		active: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#73D216")).
			Bold(true),
		key: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623")),
		noVars: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		noEnv: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
	}
}

type ActivateEnvMsg struct {
	ID string
}

type CreateEnvMsg struct {
	Name string
}

type DeleteEnvMsg struct {
	ID string
}

type SaveEnvMsg struct {
	Env *entity.Environment
}

type EnvModal struct {
	mode     EnvMode
	envs     []entity.Environment
	activeID string
	cursor   int
	width    int
	height   int
	open     bool
	focused  bool

	nameInput textinput.Model

	editing     *entity.Environment
	editKeys    []string
	editCursor  int
	editField   EnvEditField
	keyInput    textinput.Model
	valueInput  textinput.Model
	editingPair bool

	keys EnvModalKeyMap

	styles *EnvModalStyles
}

func NewEnvModal() EnvModal {
	nameInput := textinput.New()
	nameInput.Placeholder = "environment name"
	nameInput.Prompt = ""
	nameInput.CharLimit = 64

	keyInput := textinput.New()
	keyInput.Placeholder = "key"
	keyInput.Prompt = ""
	keyInput.CharLimit = 64

	valInput := textinput.New()
	valInput.Placeholder = "value"
	valInput.Prompt = ""
	valInput.CharLimit = 1024

	return EnvModal{
		mode:       EnvModeList,
		nameInput:  nameInput,
		keyInput:   keyInput,
		valueInput: valInput,
		keys:       defaultEnvModalKeys(),
		styles:     NewEnvModalStyles(),
	}
}

func (m *EnvModal) Open() tea.Cmd {
	m.open = true
	m.focused = true
	m.mode = EnvModeList
	m.cursor = 0

	return nil
}

func (m *EnvModal) Close() {
	m.open = false
	m.focused = false
	m.mode = EnvModeList

	m.nameInput.Blur()
	m.nameInput.SetValue("")

	m.keyInput.Blur()
	m.valueInput.Blur()

	m.editing = nil
	m.editingPair = false
}

func (m EnvModal) IsOpen() bool {
	return m.open
}

func (m *EnvModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *EnvModal) SetData(envs []entity.Environment, activeID string) {
	m.envs = envs
	m.activeID = activeID
	if m.cursor >= len(envs) {
		m.cursor = max(0, len(envs)-1)
	}
}

func (m EnvModal) Update(msg tea.Msg) (EnvModal, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	switch m.mode {
	case EnvModeList:
		return m.updateList(msg)
	case EnvModeCreate:
		return m.updateCreate(msg)
	case EnvModeEdit:
		return m.updateEdit(msg)
	}

	return m, nil
}

func (m EnvModal) updateList(msg tea.Msg) (EnvModal, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m, nil
	}

	switch {
	case key.Matches(km, m.keys.Close):
		m.Close()
		return m, nil
	case key.Matches(km, m.keys.Up):
		if m.cursor > 0 {
			m.cursor--
		}

		return m, nil
	case key.Matches(km, m.keys.Down):
		if m.cursor < len(m.envs)-1 {
			m.cursor++
		}

		return m, nil
	case key.Matches(km, m.keys.Select):
		if len(m.envs) == 0 {
			return m, nil
		}

		env := m.envs[m.cursor]

		id := env.ID
		if m.activeID == id {
			id = ""
		}

		return m, func() tea.Msg {
			return ActivateEnvMsg{
				ID: id,
			}
		}
	case key.Matches(km, m.keys.New):
		m.mode = EnvModeCreate
		m.nameInput.SetValue("")

		return m, m.nameInput.Focus()
	case key.Matches(km, m.keys.Edit):
		if len(m.envs) == 0 {
			return m, nil
		}

		env := m.envs[m.cursor]

		m.editing = &entity.Environment{
			ID:        env.ID,
			Name:      env.Name,
			Variables: make(map[string]string),
		}

		for k, v := range env.Variables {
			m.editing.Variables[k] = v
		}

		m.refreshEditKeys()
		m.editCursor = 0
		m.editingPair = false
		m.mode = EnvModeEdit

		return m, nil
	case key.Matches(km, m.keys.Delete):
		if len(m.envs) == 0 {
			return m, nil
		}

		env := m.envs[m.cursor]

		return m, func() tea.Msg {
			return DeleteEnvMsg{
				ID: env.ID,
			}
		}
	}

	return m, nil
}

func (m *EnvModal) refreshEditKeys() {
	m.editKeys = make([]string, 0, len(m.editing.Variables))

	for k := range m.editing.Variables {
		m.editKeys = append(m.editKeys, k)
	}

	for i := 1; i < len(m.editKeys); i++ {
		for j := i; j > 0 && m.editKeys[j] < m.editKeys[j-1]; j-- {
			m.editKeys[j], m.editKeys[j-1] = m.editKeys[j-1], m.editKeys[j]
		}
	}
}

func (m EnvModal) updateCreate(msg tea.Msg) (EnvModal, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(km, m.keys.Close):
			m.nameInput.Blur()
			m.mode = EnvModeList

			return m, nil
		case km.String() == "enter":
			name := strings.TrimSpace(m.nameInput.Value())

			if len(name) == 0 {
				return m, nil
			}

			m.nameInput.Blur()
			m.nameInput.SetValue("")
			m.mode = EnvModeList

			return m, func() tea.Msg {
				return CreateEnvMsg{
					Name: name,
				}
			}
		}
	}

	var cmd tea.Cmd
	m.nameInput, cmd = m.nameInput.Update(msg)

	return m, cmd
}

func (m EnvModal) updateEdit(msg tea.Msg) (EnvModal, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m.routeEditInput(msg)
	}

	if m.editingPair {
		switch km.String() {
		case "esc":
			m.keyInput.Blur()
			m.valueInput.Blur()
			m.editingPair = false

			return m, nil
		case "tab":
			if m.editField == EnvEditKey {
				m.keyInput.Blur()
				m.editField = EnvEditValue

				return m, m.valueInput.Focus()
			}

			m.valueInput.Blur()
			m.editField = EnvEditKey

			return m, m.keyInput.Focus()
		case "enter":
			k := strings.TrimSpace(m.keyInput.Value())
			v := m.valueInput.Value()

			if len(k) != 0 {
				m.editing.Variables[k] = v
				m.refreshEditKeys()
			}

			m.keyInput.Blur()
			m.valueInput.Blur()
			m.keyInput.SetValue("")
			m.valueInput.SetValue("")
			m.editingPair = false

			return m, nil
		}

		return m.routeEditInput(msg)
	}

	switch {
	case key.Matches(km, m.keys.Close):
		m.editing = nil
		m.mode = EnvModeList

		return m, nil
	case key.Matches(km, m.keys.Save):
		env := m.editing
		m.editing = nil
		m.mode = EnvModeList

		return m, func() tea.Msg {
			return SaveEnvMsg{
				Env: env,
			}
		}
	case key.Matches(km, m.keys.Up):
		if m.editCursor > 0 {
			m.editCursor--
		}

		return m, nil
	case key.Matches(km, m.keys.Down):
		if m.editCursor < len(m.editKeys)-1 {
			m.editCursor++
		}

		return m, nil
	case key.Matches(km, m.keys.Add):
		m.editingPair = true
		m.editField = EnvEditKey
		m.keyInput.SetValue("")
		m.valueInput.SetValue("")

		return m, m.keyInput.Focus()
	case key.Matches(km, m.keys.Edit):
		if len(m.editKeys) == 0 {
			return m, nil
		}

		k := m.editKeys[m.editCursor]
		m.editingPair = true
		m.editField = EnvEditValue
		m.keyInput.SetValue(k)
		m.valueInput.SetValue(m.editing.Variables[k])

		return m, m.valueInput.Focus()
	case key.Matches(km, m.keys.Delete):
		if len(m.editKeys) == 0 {
			return m, nil
		}

		k := m.editKeys[m.editCursor]
		delete(m.editing.Variables, k)
		m.refreshEditKeys()

		if m.editCursor >= len(m.editKeys) {
			m.editCursor = max(0, len(m.editKeys)-1)
		}

		return m, nil
	}

	return m, nil
}

func (m EnvModal) routeEditInput(msg tea.Msg) (EnvModal, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case m.keyInput.Focused():
		m.keyInput, cmd = m.keyInput.Update(msg)
	case m.valueInput.Focused():
		m.valueInput, cmd = m.valueInput.Update(msg)
	}

	return m, cmd
}

func (m EnvModal) View() string {
	if !m.open {
		return ""
	}

	modalW := m.width / 2

	if modalW < 40 {
		modalW = 40
	}

	if modalW > 80 {
		modalW = 80
	}

	var content string
	switch m.mode {
	case EnvModeList:
		content = m.viewList()
	case EnvModeCreate:
		content = m.viewCreate()
	case EnvModeEdit:
		content = m.viewEdit()
	}

	return m.styles.border.Width(modalW).Render(content)
}

func (m EnvModal) viewList() string {
	var b strings.Builder
	b.WriteString(m.styles.title.Render("Environments"))
	b.WriteString("\n\n")

	switch len(m.envs) == 0 {
	case true:
		b.WriteString(m.styles.noEnv.Render("(no environments)"))
	default:
		for i := range m.envs {
			marker := "  "

			if m.envs[i].ID == m.activeID {
				marker = m.styles.active.Render("● ")
			}

			line := fmt.Sprintf("%s%s (%d vars)", marker, m.envs[i].Name, len(m.envs[i].Variables))

			switch i == m.cursor {
			case true:
				b.WriteString(m.styles.sel.Render(line))
			default:
				b.WriteString(m.styles.muted.Render(line))
			}

			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.styles.cmdsHelp.Render(
		"enter: activate/clear · ctrl+o: new · ctrl+y: edit · ctrl+g: delete · esc: close",
	))

	return b.String()
}

func (m EnvModal) viewCreate() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("New Environment"))
	b.WriteString("\n\n")
	b.WriteString("Name: ")
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")
	b.WriteString(m.styles.cmdsHelp.Render(
		"enter: create · esc: cancel",
	))

	return b.String()
}

func (m EnvModal) viewEdit() string {
	var b strings.Builder

	b.WriteString(m.styles.title.Render("Edit " + m.editing.Name))
	b.WriteString("\n\n")

	switch len(m.editKeys) == 0 {
	case true:
		b.WriteString(m.styles.noVars.Render("(no variables)"))
	default:
		for i := range m.editKeys {
			v := m.editing.Variables[m.editKeys[i]]
			if len(v) > 30 {
				v = v[:27] + "..."
			}

			line := fmt.Sprintf("%s = %s", m.styles.key.Render(m.editKeys[i]), v)

			switch i == m.editCursor && !m.editingPair {
			case true:
				b.WriteString(m.styles.sel.Render(fmt.Sprintf("%-40s", line)))
			default:
				b.WriteString(m.styles.muted.Render(line))
			}

			b.WriteString("\n")
		}
	}

	switch m.editingPair {
	case true:
		b.WriteString("\n")
		b.WriteString("Key:   " + m.keyInput.View() + "\n")
		b.WriteString("Value: " + m.valueInput.View() + "\n")
		b.WriteString(m.styles.cmdsHelp.Render(
			"tab: switch field · enter: confirm · esc: cancel",
		))
	default:
		b.WriteString("\n")
		b.WriteString(m.styles.cmdsHelp.Render(
			"ctrl+o: add · ctrl+y: edit · ctrl+g: delete · ctrl+s: save · esc: cancel",
		))
	}

	return b.String()
}
