package components

import (
	"charm.land/bubbles/v2/help"
)

type StatusBar struct {
	help help.Model
}

func NewStatusBar() StatusBar {
	return StatusBar{
		help: help.New(),
	}
}

func (c *StatusBar) SetWidth(w int) {
	c.help.SetWidth(w)
}

func (c StatusBar) View(km help.KeyMap) string {
	return c.help.View(km)
}
