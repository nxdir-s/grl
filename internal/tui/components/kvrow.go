package components

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type KVField int

const (
	KVFieldKey KVField = iota
	KVFieldValue
)

type KVRow struct {
	key     textinput.Model
	value   textinput.Model
	enabled bool
}

func NewKVRow(keyPlaceholder string, valuePlaceholder string) KVRow {
	key := textinput.New()
	key.Placeholder = keyPlaceholder
	key.Prompt = ""
	key.CharLimit = 256

	value := textinput.New()
	value.Placeholder = valuePlaceholder
	value.Prompt = ""
	value.CharLimit = 1024

	return KVRow{
		key:     key,
		value:   value,
		enabled: true,
	}
}

func (c *KVRow) focusField(field KVField) tea.Cmd {
	c.key.Blur()
	c.value.Blur()

	switch field {
	case KVFieldKey:
		return c.key.Focus()
	case KVFieldValue:
		return c.value.Focus()
	default:
		return nil
	}
}

func (c *KVRow) blur() {
	c.key.Blur()
	c.value.Blur()
}

func (c KVRow) update(msg tea.Msg) (KVRow, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if c.key.Focused() {
		c.key, cmd = c.key.Update(msg)
		cmds = append(cmds, cmd)
	}

	if c.value.Focused() {
		c.value, cmd = c.value.Update(msg)
		cmds = append(cmds, cmd)
	}

	return c, tea.Batch(cmds...)
}
