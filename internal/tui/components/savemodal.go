package components

import (
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/entity"
)

type SaveModalFocus int

const (
	SaveModalFocusName SaveModalFocus = iota
	SaveModalFocusList
	SaveModalFocusNewCol
)

const (
	SaveModalCmds string = "tab: next field · enter: save · esc: cancel"

	SaveModalLabel       string = "Save Request"
	SaveModalNewColLabel string = "New name: "
	SaveModalNameLabel   string = "Name: "

	SaveModalCollectionsLabel   string = "Collection:"
	SaveModalNewCollectionLabel string = "+ New collection"

	SaveModalRowPrefix       string = "  "
	SaveModalActiveRowPrefix string = "▸ "
)

type SaveModalKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Tab    key.Binding
	Select key.Binding
	Close  key.Binding
}

func defaultSaveModalKeys() SaveModalKeyMap {
	return SaveModalKeyMap{
		Up:     key.NewBinding(key.WithKeys("up", "ctrl+p")),
		Down:   key.NewBinding(key.WithKeys("down", "ctrl+n")),
		Tab:    key.NewBinding(key.WithKeys("tab")),
		Select: key.NewBinding(key.WithKeys("enter")),
		Close:  key.NewBinding(key.WithKeys("esc")),
	}
}

type SaveModalStyles struct {
	label  lipgloss.Style
	dim    lipgloss.Style
	active lipgloss.Style
	normal lipgloss.Style
	border lipgloss.Style
}

func NewSaveModalStyles() *SaveModalStyles {
	return &SaveModalStyles{
		label: lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true),
		dim:   lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
		active: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")),
		normal: lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(50),
	}
}

type SaveModal struct {
	open        bool
	collections []entity.Collection
	cursor      int
	focus       SaveModalFocus

	nameInput   textinput.Model
	newColInput textinput.Model

	keys   SaveModalKeyMap
	styles *SaveModalStyles
}

func NewSaveModal() SaveModal {
	name := textinput.New()
	name.Placeholder = "My request"
	name.Prompt = ""
	name.CharLimit = 128

	newCol := textinput.New()
	newCol.Placeholder = "New collection name"
	newCol.Prompt = ""
	newCol.CharLimit = 128

	return SaveModal{
		nameInput:   name,
		newColInput: newCol,
		keys:        defaultSaveModalKeys(),
		styles:      NewSaveModalStyles(),
	}
}

func (m SaveModal) IsOpen() bool { return m.open }

func (m *SaveModal) Open(collections []entity.Collection, defaultName string) tea.Cmd {
	m.open = true
	m.collections = collections
	m.cursor = 0
	m.focus = SaveModalFocusName
	m.nameInput.SetValue(defaultName)
	m.newColInput.SetValue("")
	m.newColInput.Blur()

	return m.nameInput.Focus()
}

func (m *SaveModal) Close() {
	m.open = false
	m.nameInput.Blur()
	m.newColInput.Blur()
}

func (m SaveModal) Update(msg tea.Msg) (SaveModal, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	if km, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(km, m.keys.Close):
			m.Close()

			return m, func() tea.Msg {
				return CloseSaveModalMsg{}
			}
		case key.Matches(km, m.keys.Tab):
			return m.cycleFocus()
		}

		switch m.focus {
		case SaveModalFocusList:
			switch {
			case key.Matches(km, m.keys.Up):
				if m.cursor > 0 {
					m.cursor--
				}

				return m, nil
			case key.Matches(km, m.keys.Down):
				if m.cursor < len(m.collections) {
					m.cursor++
				}

				return m, nil
			case key.Matches(km, m.keys.Select):
				return m.commit()
			}
		case SaveModalFocusName, SaveModalFocusNewCol:
			if key.Matches(km, m.keys.Select) {
				// Commit from name field: if list has selection, save into it
				if m.focus == SaveModalFocusName {
					return m.commit()
				}

				// From newCol input: must have a name
				return m.commit()
			}
		}
	}

	var cmd tea.Cmd
	switch m.focus {
	case SaveModalFocusName:
		m.nameInput, cmd = m.nameInput.Update(msg)
	case SaveModalFocusNewCol:
		m.newColInput, cmd = m.newColInput.Update(msg)
	}

	return m, cmd
}

func (m SaveModal) cycleFocus() (SaveModal, tea.Cmd) {
	m.nameInput.Blur()
	m.newColInput.Blur()

	switch m.focus {
	case SaveModalFocusName:
		m.focus = SaveModalFocusList
		return m, nil
	case SaveModalFocusList:
		if m.cursor == 0 {
			m.focus = SaveModalFocusNewCol
			return m, m.newColInput.Focus()
		}

		m.focus = SaveModalFocusName
		return m, m.nameInput.Focus()
	case SaveModalFocusNewCol:
		m.focus = SaveModalFocusName
		return m, m.nameInput.Focus()
	}

	return m, nil
}

func (m SaveModal) commit() (SaveModal, tea.Cmd) {
	reqName := strings.TrimSpace(m.nameInput.Value())

	if len(reqName) == 0 {
		return m, nil
	}

	if m.cursor == 0 {
		return m.commitNewCollection(reqName)
	}

	col := m.collections[m.cursor-1]
	m.Close()

	return m, func() tea.Msg {
		return SaveToCollectionMsg{
			CollectionID: col.ID,
			RequestName:  reqName,
		}
	}
}

func (m SaveModal) commitNewCollection(reqName string) (SaveModal, tea.Cmd) {
	colName := strings.TrimSpace(m.newColInput.Value())

	if len(colName) == 0 {
		m.focus = SaveModalFocusNewCol
		return m, m.newColInput.Focus()
	}

	m.Close()

	return m, func() tea.Msg {
		return CreateAndSaveMsg{
			CollectionName: colName,
			RequestName:    reqName,
		}
	}
}

func (m SaveModal) View() string {
	if !m.open {
		return ""
	}

	var b strings.Builder
	b.WriteString(m.styles.label.Render(SaveModalLabel))
	b.WriteString("\n\n")

	switch {
	case m.focus == SaveModalFocusName:
		b.WriteString(m.styles.label.Render(SaveModalNameLabel))
	default:
		b.WriteString(m.styles.dim.Render(SaveModalNameLabel))
	}

	b.WriteString(m.nameInput.View())
	b.WriteString("\n\n")

	b.WriteString(m.styles.label.Render(SaveModalCollectionsLabel))
	b.WriteString("\n")

	rows := []string{SaveModalNewCollectionLabel}
	for i := range m.collections {
		rows = append(rows, m.collections[i].Name)
	}

	for i := range rows {
		prefix := SaveModalRowPrefix

		switch {
		case m.focus == SaveModalFocusList && i == m.cursor:
			prefix = SaveModalActiveRowPrefix
			b.WriteString(m.styles.active.Render(prefix + rows[i]))
		case m.focus != SaveModalFocusList && i == m.cursor:
			b.WriteString(m.styles.label.Render(prefix + rows[i]))
		default:
			b.WriteString(m.styles.normal.Render(prefix + rows[i]))
		}

		b.WriteString("\n")
	}

	if m.cursor == 0 {
		b.WriteString("\n")

		switch {
		case m.focus == SaveModalFocusNewCol:
			b.WriteString(m.styles.label.Render(SaveModalNewColLabel))
		default:
			b.WriteString(m.styles.dim.Render(SaveModalNewColLabel))
		}

		b.WriteString(m.newColInput.View())
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.dim.Render(SaveModalCmds))

	return m.styles.border.Render(b.String())
}
