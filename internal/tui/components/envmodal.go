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

type Mode int

const (
	ModeList Mode = iota
	ModeCreate
	ModeEdit
)

type EditField int

const (
	EditKey EditField = iota
	EditValue
)

type ModalKeyMap struct {
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

func defaultKeys() ModalKeyMap {
	return ModalKeyMap{
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

type Modal struct {
	mode     Mode
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
	editField   EditField
	keyInput    textinput.Model
	valueInput  textinput.Model
	editingPair bool

	keys ModalKeyMap
}

func NewModal() Modal {
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

	return Modal{
		mode:       ModeList,
		nameInput:  nameInput,
		keyInput:   keyInput,
		valueInput: valInput,
		keys:       defaultKeys(),
	}
}

func (m *Modal) Open() tea.Cmd {
	m.open = true
	m.focused = true
	m.mode = ModeList
	m.cursor = 0

	return nil
}

func (m *Modal) Close() {
	m.open = false
	m.focused = false
	m.mode = ModeList

	m.nameInput.Blur()
	m.nameInput.SetValue("")

	m.keyInput.Blur()
	m.valueInput.Blur()

	m.editing = nil
	m.editingPair = false
}

func (m Modal) IsOpen() bool {
	return m.open
}

func (m *Modal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m *Modal) SetData(envs []entity.Environment, activeID string) {
	m.envs = envs
	m.activeID = activeID
	if m.cursor >= len(envs) {
		m.cursor = max(0, len(envs)-1)
	}
}

func (m Modal) Update(msg tea.Msg) (Modal, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	switch m.mode {
	case ModeList:
		return m.updateList(msg)
	case ModeCreate:
		return m.updateCreate(msg)
	case ModeEdit:
		return m.updateEdit(msg)
	}

	return m, nil
}

func (m Modal) updateList(msg tea.Msg) (Modal, tea.Cmd) {
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
		m.mode = ModeCreate
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
		m.mode = ModeEdit

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

func (m *Modal) refreshEditKeys() {
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

func (m Modal) updateCreate(msg tea.Msg) (Modal, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if ok {
		switch {
		case key.Matches(km, m.keys.Close):
			m.nameInput.Blur()
			m.mode = ModeList

			return m, nil
		case km.String() == "enter":
			name := strings.TrimSpace(m.nameInput.Value())

			if len(name) == 0 {
				return m, nil
			}

			m.nameInput.Blur()
			m.nameInput.SetValue("")
			m.mode = ModeList

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

func (m Modal) updateEdit(msg tea.Msg) (Modal, tea.Cmd) {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return m.routeEditInput(msg)
	}

	// If editing a pair, route keys to inputs unless finalizing
	if m.editingPair {
		switch km.String() {
		case "esc":
			m.keyInput.Blur()
			m.valueInput.Blur()
			m.editingPair = false

			return m, nil
		case "tab":
			if m.editField == EditKey {
				m.keyInput.Blur()
				m.editField = EditValue

				return m, m.valueInput.Focus()
			}

			m.valueInput.Blur()
			m.editField = EditKey

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
		m.mode = ModeList

		return m, nil
	case key.Matches(km, m.keys.Save):
		env := m.editing
		m.editing = nil
		m.mode = ModeList

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
		m.editField = EditKey
		m.keyInput.SetValue("")
		m.valueInput.SetValue("")

		return m, m.keyInput.Focus()
	case key.Matches(km, m.keys.Edit):
		if len(m.editKeys) == 0 {
			return m, nil
		}

		k := m.editKeys[m.editCursor]
		m.editingPair = true
		m.editField = EditValue
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

func (m Modal) routeEditInput(msg tea.Msg) (Modal, tea.Cmd) {
	var cmd tea.Cmd

	switch {
	case m.keyInput.Focused():
		m.keyInput, cmd = m.keyInput.Update(msg)
	case m.valueInput.Focused():
		m.valueInput, cmd = m.valueInput.Update(msg)
	}

	return m, cmd
}

func (m Modal) View() string {
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

	border := m.borderStyle(modalW)
	title := m.titleStyle()

	var content string
	switch m.mode {
	case ModeList:
		content = m.viewList(title)
	case ModeCreate:
		content = m.viewCreate(title)
	case ModeEdit:
		content = m.viewEdit(title)
	}

	return border.Render(content)
}

func (m Modal) borderStyle(width int) lipgloss.Style {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(1, 2).
		Width(width)
}

func (m Modal) titleStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
}

func (m Modal) viewList(title lipgloss.Style) string {
	var b strings.Builder
	b.WriteString(title.Render("Environments"))
	b.WriteString("\n\n")

	switch len(m.envs) == 0 {
	case true:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("(no environments)"))
	default:
		sel := m.selStyle()
		muted := m.mutedStyle()
		active := m.activeStyle()

		for i := range m.envs {
			marker := "  "

			if m.envs[i].ID == m.activeID {
				marker = active.Render("● ")
			}

			line := fmt.Sprintf("%s%s (%d vars)", marker, m.envs[i].Name, len(m.envs[i].Variables))

			switch i == m.cursor {
			case true:
				b.WriteString(sel.Render(line))
			default:
				b.WriteString(muted.Render(line))
			}

			b.WriteString("\n")
		}
	}

	b.WriteString("\n")
	b.WriteString(m.keyCmdsHelp())

	return b.String()
}

func (m Modal) keyCmdsHelp() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		"enter: activate/clear · ctrl+o: new · ctrl+y: edit · ctrl+g: delete · esc: close",
	)
}

func (m Modal) selStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4"))
}

func (m Modal) mutedStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA"))
}

func (m Modal) activeStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#73D216")).Bold(true)
}

func (m Modal) keyStyle() lipgloss.Style {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623"))
}

func (m Modal) viewCreate(title lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(title.Render("New Environment"))
	b.WriteString("\n\n")
	b.WriteString("Name: ")
	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")
	b.WriteString(m.createEnvCmdsHelp())

	return b.String()
}

func (m Modal) createEnvCmdsHelp() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		"enter: create · esc: cancel",
	)
}

func (m Modal) viewEdit(title lipgloss.Style) string {
	var b strings.Builder

	b.WriteString(title.Render("Edit " + m.editing.Name))
	b.WriteString("\n\n")

	switch len(m.editKeys) == 0 {
	case true:
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render("(no variables)"))
	default:
		sel := m.selStyle()
		muted := m.mutedStyle()
		keyStyle := m.keyStyle()

		for i := range m.editKeys {
			v := m.editing.Variables[m.editKeys[i]]
			if len(v) > 30 {
				v = v[:27] + "..."
			}

			line := fmt.Sprintf("%s = %s", keyStyle.Render(m.editKeys[i]), v)

			switch i == m.editCursor && !m.editingPair {
			case true:
				b.WriteString(sel.Render(fmt.Sprintf("%-40s", line)))
			default:
				b.WriteString(muted.Render(line))
			}

			b.WriteString("\n")
		}
	}

	switch m.editingPair {
	case true:
		b.WriteString("\n")
		b.WriteString("Key:   " + m.keyInput.View() + "\n")
		b.WriteString("Value: " + m.valueInput.View() + "\n")
		b.WriteString(m.editPairCmdsHelp())
	default:
		b.WriteString("\n")
		b.WriteString(m.editCmdsHelp())
	}

	return b.String()
}

func (m Modal) editPairCmdsHelp() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		"tab: switch field · enter: confirm · esc: cancel",
	)
}

func (m Modal) editCmdsHelp() string {
	return lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render(
		"ctrl+o: add · ctrl+y: edit · ctrl+g: delete · ctrl+s: save · esc: cancel",
	)
}
