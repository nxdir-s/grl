package components

import (
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type MethodSelector struct {
	methods []valobj.HTTPMethod
	index   int
}

func NewMethodSelector() *MethodSelector {
	return &MethodSelector{
		methods: []valobj.HTTPMethod{
			valobj.MethodGet,
			valobj.MethodPost,
			valobj.MethodPut,
			valobj.MethodDelete,
			valobj.MethodPatch,
			valobj.MethodHead,
			valobj.MethodOptions,
		},
	}
}

func (c *MethodSelector) Next() {
	c.index = (c.index + 1) % len(c.methods)
}

func (c *MethodSelector) Current() valobj.HTTPMethod {
	return c.methods[c.index]
}

func (c *MethodSelector) View() string {
	style := lipgloss.NewStyle().Bold(true)

	switch c.Current() {
	case valobj.MethodGet:
		style = style.Foreground(lipgloss.Color("#73D216"))
	case valobj.MethodPost:
		style = style.Foreground(lipgloss.Color("#F5A623"))
	case valobj.MethodPut:
		style = style.Foreground(lipgloss.Color("#4A9EF7"))
	case valobj.MethodDelete:
		style = style.Foreground(lipgloss.Color("#FF4444"))
	case valobj.MethodPatch:
		style = style.Foreground(lipgloss.Color("#C678DD"))
	default:
		style = style.Foreground(lipgloss.Color("#626262"))
	}

	return style.Render(c.Current().String())
}
