package components

import (
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
)

type URLBar struct {
	input textinput.Model
}

func NewURLBar() URLBar {
	input := textinput.New()

	input.Placeholder = "https://api.example.com/endpoint"
	input.Prompt = ""
	input.CharLimit = 2048

	return URLBar{
		input: input,
	}
}

func (c *URLBar) Focus() tea.Cmd {
	return c.input.Focus()
}

func (c *URLBar) Blur() {
	c.input.Blur()
}

func (c URLBar) Value() string {
	return c.input.Value()
}

func (c *URLBar) SetValue(s string) {
	c.input.SetValue(s)
}

func (c *URLBar) SetWidth(w int) {
	c.input.CharLimit = 2048
	// textinput doesn't have a Width setter in v2, but the prompt width is handled by the parent
}

func (c URLBar) Update(msg tea.Msg) (URLBar, tea.Cmd) {
	var cmd tea.Cmd
	c.input, cmd = c.input.Update(msg)

	return c, cmd
}

func (c URLBar) View() string {
	return c.input.View()
}
