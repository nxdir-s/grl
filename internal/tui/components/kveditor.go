package tui

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type KVEditor struct {
	rows           []*KVRow
	cursor         int
	focusedField   KVField
	focused        bool
	keyPlaceholder string
	valPlaceholder string
	width          int

	keys KVEditorKeyMap
}

type KVEditorKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	TabField key.Binding
	AddRow   key.Binding
	DelRow   key.Binding
	Toggle   key.Binding
}

func NewKVEditor(keyPlaceholder, valPlaceholder string) KVEditor {
	editor := KVEditor{
		keyPlaceholder: keyPlaceholder,
		valPlaceholder: valPlaceholder,
		keys:           defaultKVEditorKeys(),
	}

	editor.addRow()

	return editor
}

func (e *KVEditor) addRow() {
	e.rows = append(e.rows, NewKVRow(e.keyPlaceholder, e.valPlaceholder))
}

func defaultKVEditorKeys() KVEditorKeyMap {
	return KVEditorKeyMap{
		Up: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "prev row"),
		),
		Down: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "next row"),
		),
		TabField: key.NewBinding(
			key.WithKeys("tab"),
			key.WithHelp("tab", "next field"),
		),
		AddRow: key.NewBinding(
			key.WithKeys("ctrl+o"),
			key.WithHelp("ctrl+o", "add row"),
		),
		DelRow: key.NewBinding(
			key.WithKeys("ctrl+g"),
			key.WithHelp("ctrl+g", "delete row"),
		),
		Toggle: key.NewBinding(
			key.WithKeys("ctrl+y"),
			key.WithHelp("ctrl+y", "toggle row"),
		),
	}
}

func (e *KVEditor) Focus() tea.Cmd {
	e.focused = true

	if len(e.rows) == 0 {
		e.addRow()
	}

	e.cursor = 0
	e.focusedField = KVFieldKey

	return e.rows[0].focusField(KVFieldKey)
}

func (e *KVEditor) Blur() {
	e.focused = false

	for i := range e.rows {
		e.rows[i].blur()
	}
}

func (e *KVEditor) SetWidth(w int) {
	e.width = w
}

func (e KVEditor) Headers() []valobj.Header {
	var headers []valobj.Header

	for _, r := range e.rows {
		k := r.key.Value()
		v := r.value.Value()
		if k != "" || v != "" {
			headers = append(headers, valobj.Header{
				Key:     k,
				Value:   v,
				Enabled: r.enabled,
			})
		}
	}

	return headers
}

func (e KVEditor) QueryParams() []valobj.QueryParam {
	var params []valobj.QueryParam

	for _, r := range e.rows {
		k := r.key.Value()
		v := r.value.Value()
		if k != "" || v != "" {
			params = append(params, valobj.QueryParam{
				Key:     k,
				Value:   v,
				Enabled: r.enabled,
			})
		}
	}

	return params
}

func (e *KVEditor) Update(msg tea.Msg) tea.Cmd {
	if !e.focused || len(e.rows) == 0 {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, e.keys.Up):
			if e.cursor > 0 {
				e.rows[e.cursor].blur()
				e.cursor--

				return e.rows[e.cursor].focusField(e.focusedField)
			}

			return nil
		case key.Matches(msg, e.keys.Down):
			if e.cursor < len(e.rows)-1 {
				e.rows[e.cursor].blur()
				e.cursor++

				return e.rows[e.cursor].focusField(e.focusedField)
			}

			return nil
		case key.Matches(msg, e.keys.TabField):
			if e.focusedField == KVFieldKey {
				e.focusedField = KVFieldValue
			} else {
				e.focusedField = KVFieldKey
			}

			return e.rows[e.cursor].focusField(e.focusedField)
		case key.Matches(msg, e.keys.AddRow):
			e.rows[e.cursor].blur()
			e.addRow()
			e.cursor = len(e.rows) - 1
			e.focusedField = KVFieldKey

			return e.rows[e.cursor].focusField(KVFieldKey)
		case key.Matches(msg, e.keys.DelRow):
			if len(e.rows) <= 1 {
				return nil
			}

			e.rows[e.cursor].blur()
			e.rows = append(e.rows[:e.cursor], e.rows[e.cursor+1:]...)

			if e.cursor >= len(e.rows) {
				e.cursor = len(e.rows) - 1
			}

			return e.rows[e.cursor].focusField(e.focusedField)
		case key.Matches(msg, e.keys.Toggle):
			e.rows[e.cursor].enabled = !e.rows[e.cursor].enabled
			return nil
		}
	}

	return e.rows[e.cursor].update(msg)
}

func (e *KVEditor) View() string {
	if len(e.rows) == 0 {
		return lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Render("  No entries. Press ctrl+n to add.")
	}

	colWidth := e.width / 2
	if colWidth < 10 {
		colWidth = 10
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#626262")).
		Width(colWidth)

	header := fmt.Sprintf("  %s%s",
		headerStyle.Render("Key"),
		headerStyle.Render("Value"),
	)

	rowStyle := lipgloss.NewStyle().Width(colWidth)

	disabledStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Width(colWidth)

	cursorStr := "▸ "
	noCursor := "  "

	var rows string
	for i, r := range e.rows {
		prefix := noCursor
		if e.focused && i == e.cursor {
			prefix = cursorStr
		}

		style := rowStyle
		if !r.enabled {
			style = disabledStyle
		}

		toggleMark := "●"
		if !r.enabled {
			toggleMark = "○"
		}

		toggleStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#626262"))

		row := fmt.Sprintf("%s%s %s%s",
			prefix,
			toggleStyle.Render(toggleMark),
			style.Render(r.key.View()),
			style.Render(r.value.View()),
		)

		rows += row + "\n"
	}

	return header + "\n" + rows
}
