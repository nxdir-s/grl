package components

import (
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type MethodSelectorStyles struct {
	get      lipgloss.Style
	post     lipgloss.Style
	put      lipgloss.Style
	delete   lipgloss.Style
	patch    lipgloss.Style
	fallback lipgloss.Style
}

func NewMethodSelectorStyles() *MethodSelectorStyles {
	return &MethodSelectorStyles{
		get: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#73D216")),
		post: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F5A623")),
		put: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#4A9EF7")),
		delete: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FF4444")),
		patch: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#C678DD")),
		fallback: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#626262")),
	}
}

type MethodSelector struct {
	methods []valobj.HTTPMethod
	index   int
	styles  *MethodSelectorStyles
}

func NewMethodSelector() MethodSelector {
	return MethodSelector{
		methods: []valobj.HTTPMethod{
			valobj.MethodGet,
			valobj.MethodPost,
			valobj.MethodPut,
			valobj.MethodDelete,
			valobj.MethodPatch,
			valobj.MethodHead,
			valobj.MethodOptions,
		},
		styles: NewMethodSelectorStyles(),
	}
}

func (c *MethodSelector) Next() {
	c.index = (c.index + 1) % len(c.methods)
}

func (m *MethodSelector) SetCurrent(method valobj.HTTPMethod) {
	for i := range m.methods {
		if m.methods[i] == method {
			m.index = i
			return
		}
	}
}

func (c MethodSelector) Current() valobj.HTTPMethod {
	return c.methods[c.index]
}

func (c MethodSelector) View() string {
	switch c.Current() {
	case valobj.MethodGet:
		return c.styles.get.Render(c.Current().String())
	case valobj.MethodPost:
		return c.styles.post.Render(c.Current().String())
	case valobj.MethodPut:
		return c.styles.put.Render(c.Current().String())
	case valobj.MethodDelete:
		return c.styles.delete.Render(c.Current().String())
	case valobj.MethodPatch:
		return c.styles.patch.Render(c.Current().String())
	default:
		return c.styles.fallback.Render(c.Current().String())
	}
}
