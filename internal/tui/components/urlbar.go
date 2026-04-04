package tui

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type URLBar struct {
	input textinput.Model
}

func NewURLBar() *URLBar {
	ti := textinput.New()
	ti.Placeholder = "https://api.example.com/endpoint"
	ti.Prompt = ""
	ti.CharLimit = 2048

	return &URLBar{
		input: ti,
	}
}

func (u *URLBar) Focus() tea.Cmd {
	return u.input.Focus()
}

func (u *URLBar) Blur() {
	u.input.Blur()
}

func (u *URLBar) Value() string {
	return u.input.Value()
}

func (u *URLBar) SetValue(s string) {
	u.input.SetValue(s)
}

func (u *URLBar) SetWidth(w int) {
	u.input.CharLimit = 2048
	// textinput doesn't have a Width setter in v2, but the prompt width is handled by the parent
}

func (u *URLBar) Update(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	u.input, cmd = u.input.Update(msg)

	return cmd
}

func (u *URLBar) View() string {
	return u.input.View()
}
