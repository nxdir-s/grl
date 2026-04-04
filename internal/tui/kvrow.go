package tui

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

func NewKVRow(keyPlaceholder, valuePlaceholder string) *KVRow {
	k := textinput.New()
	k.Placeholder = keyPlaceholder
	k.Prompt = ""
	k.CharLimit = 256

	v := textinput.New()
	v.Placeholder = valuePlaceholder
	v.Prompt = ""
	v.CharLimit = 1024

	return &KVRow{
		key:     k,
		value:   v,
		enabled: true,
	}
}

func (r *KVRow) focusField(field KVField) tea.Cmd {
	r.key.Blur()
	r.value.Blur()

	switch field {
	case KVFieldKey:
		return r.key.Focus()
	case KVFieldValue:
		return r.value.Focus()
	default:
		return nil
	}
}

func (r *KVRow) blur() {
	r.key.Blur()
	r.value.Blur()
}

func (r *KVRow) update(msg tea.Msg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if r.key.Focused() {
		r.key, cmd = r.key.Update(msg)
		cmds = append(cmds, cmd)
	}

	if r.value.Focused() {
		r.value, cmd = r.value.Update(msg)
		cmds = append(cmds, cmd)
	}

	return tea.Batch(cmds...)
}
