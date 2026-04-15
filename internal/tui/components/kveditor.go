package components

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type KVEditorKeyMap struct {
	Up       key.Binding
	Down     key.Binding
	TabField key.Binding
	AddRow   key.Binding
	DelRow   key.Binding
	Toggle   key.Binding
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

type KVEditorStyles struct {
	noEntries lipgloss.Style
	header    lipgloss.Style
	row       lipgloss.Style
	disabled  lipgloss.Style
	toggle    lipgloss.Style
}

func NewKVEditorStyles() *KVEditorStyles {
	return &KVEditorStyles{
		noEntries: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
		header: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#626262")),
		row: lipgloss.NewStyle(),
		disabled: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")),
		toggle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")),
	}
}

type KVEditor struct {
	rows           []KVRow
	cursor         int
	focusedField   KVField
	focused        bool
	keyPlaceholder string
	valPlaceholder string
	width          int

	keys KVEditorKeyMap

	styles *KVEditorStyles
}

func NewKVEditor(keyPlaceholder, valPlaceholder string) KVEditor {
	editor := KVEditor{
		keyPlaceholder: keyPlaceholder,
		valPlaceholder: valPlaceholder,
		keys:           defaultKVEditorKeys(),
		styles:         NewKVEditorStyles(),
	}

	editor.addRow()

	return editor
}

func (c *KVEditor) addRow() {
	c.rows = append(c.rows, NewKVRow(c.keyPlaceholder, c.valPlaceholder))
}

func (c *KVEditor) Focus() tea.Cmd {
	c.focused = true

	if len(c.rows) == 0 {
		c.addRow()
	}

	c.cursor = 0
	c.focusedField = KVFieldKey

	return c.rows[0].focusField(KVFieldKey)
}

func (c *KVEditor) Blur() {
	c.focused = false

	for i := range c.rows {
		c.rows[i].blur()
	}
}

func (c *KVEditor) SetWidth(w int) {
	c.width = w
}

func (c KVEditor) Headers() []valobj.Header {
	var headers []valobj.Header

	for i := range c.rows {
		k := c.rows[i].key.Value()
		v := c.rows[i].value.Value()

		if k != "" || v != "" {
			headers = append(headers, valobj.Header{
				Key:     k,
				Value:   v,
				Enabled: c.rows[i].enabled,
			})
		}
	}

	return headers
}

func (c *KVEditor) SetHeaders(headers []valobj.Header) {
	c.rows = make([]KVRow, 0, len(headers))

	for i := range headers {
		row := NewKVRow(c.keyPlaceholder, c.valPlaceholder)

		row.key.SetValue(headers[i].Key)
		row.value.SetValue(headers[i].Value)
		row.enabled = headers[i].Enabled

		c.rows = append(c.rows, row)
	}

	if len(c.rows) == 0 {
		c.addRow()
	}

	c.cursor = 0
	c.focusedField = KVFieldKey
}

func (c KVEditor) QueryParams() []valobj.QueryParam {
	var params []valobj.QueryParam

	for i := range c.rows {
		k := c.rows[i].key.Value()
		v := c.rows[i].value.Value()

		if k != "" || v != "" {
			params = append(params, valobj.QueryParam{
				Key:     k,
				Value:   v,
				Enabled: c.rows[i].enabled,
			})
		}
	}

	return params
}

func (c *KVEditor) SetParams(params []valobj.QueryParam) {
	c.rows = make([]KVRow, 0, len(params))

	for i := range params {
		row := NewKVRow(c.keyPlaceholder, c.valPlaceholder)

		row.key.SetValue(params[i].Key)
		row.value.SetValue(params[i].Value)
		row.enabled = params[i].Enabled

		c.rows = append(c.rows, row)
	}

	if len(c.rows) == 0 {
		c.addRow()
	}

	c.cursor = 0
	c.focusedField = KVFieldKey
}

func (c KVEditor) Update(msg tea.Msg) (KVEditor, tea.Cmd) {
	if !c.focused || len(c.rows) == 0 {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keys.Up):
			if c.cursor > 0 {
				c.rows[c.cursor].blur()
				c.cursor--

				return c, c.rows[c.cursor].focusField(c.focusedField)
			}

			return c, nil
		case key.Matches(msg, c.keys.Down):
			if c.cursor < len(c.rows)-1 {
				c.rows[c.cursor].blur()
				c.cursor++

				return c, c.rows[c.cursor].focusField(c.focusedField)
			}

			return c, nil
		case key.Matches(msg, c.keys.TabField):
			if c.focusedField == KVFieldKey {
				c.focusedField = KVFieldValue
			} else {
				c.focusedField = KVFieldKey
			}

			return c, c.rows[c.cursor].focusField(c.focusedField)
		case key.Matches(msg, c.keys.AddRow):
			c.rows[c.cursor].blur()
			c.addRow()
			c.cursor = len(c.rows) - 1
			c.focusedField = KVFieldKey

			return c, c.rows[c.cursor].focusField(KVFieldKey)
		case key.Matches(msg, c.keys.DelRow):
			if len(c.rows) <= 1 {
				return c, nil
			}

			c.rows[c.cursor].blur()
			c.rows = append(c.rows[:c.cursor], c.rows[c.cursor+1:]...)

			if c.cursor >= len(c.rows) {
				c.cursor = len(c.rows) - 1
			}

			return c, c.rows[c.cursor].focusField(c.focusedField)
		case key.Matches(msg, c.keys.Toggle):
			c.rows[c.cursor].enabled = !c.rows[c.cursor].enabled
			return c, nil
		}
	}

	var cmd tea.Cmd
	c.rows[c.cursor], cmd = c.rows[c.cursor].update(msg)

	return c, cmd
}

func (c KVEditor) View() string {
	if len(c.rows) == 0 {
		return c.styles.noEntries.Render("  No entries. Press ctrl+n to add.")
	}

	colWidth := c.width / 2
	if colWidth < 10 {
		colWidth = 10
	}

	headerStyle := c.styles.header.Width(colWidth)

	header := fmt.Sprintf("  %s%s",
		headerStyle.Render("Key"),
		headerStyle.Render("Value"),
	)

	rowStyle := c.styles.row.Width(colWidth)
	disabledStyle := c.styles.disabled.Width(colWidth)

	cursorStr := "▸ "
	noCursor := "  "

	var rows string
	for i := range c.rows {
		prefix := noCursor
		if c.focused && i == c.cursor {
			prefix = cursorStr
		}

		style := rowStyle
		if !c.rows[i].enabled {
			style = disabledStyle
		}

		toggleMark := "●"
		if !c.rows[i].enabled {
			toggleMark = "○"
		}

		row := fmt.Sprintf("%s%s %s%s",
			prefix,
			c.styles.toggle.Render(toggleMark),
			style.Render(c.rows[i].key.View()),
			style.Render(c.rows[i].value.View()),
		)

		rows += row + "\n"
	}

	return header + "\n" + rows
}
