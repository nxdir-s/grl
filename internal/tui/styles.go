package tui

import "charm.land/lipgloss/v2"

type Styles struct {
	base            lipgloss.Style
	method          lipgloss.Style
	urlBar          lipgloss.Style
	statusSuccess   lipgloss.Style
	statusRedirect  lipgloss.Style
	statusClientErr lipgloss.Style
	statusServerErr lipgloss.Style
	timing          lipgloss.Style
	responseBody    lipgloss.Style
	responseSection lipgloss.Style
	statusBar       lipgloss.Style
	focusIndicator  lipgloss.Style
	title           lipgloss.Style
	loading         lipgloss.Style
	error           lipgloss.Style
	envBadge        lipgloss.Style
	flash           lipgloss.Style
}

func NewStyles(themes *Themes) *Styles {
	return &Styles{
		base: lipgloss.NewStyle(),
		method: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorPrimary).
			Padding(0, 1),
		urlBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(themes.colorBorder).
			Padding(0, 1),
		statusSuccess: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorSuccess),
		statusRedirect: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorWarning),
		statusClientErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorSecondary),
		statusServerErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorError),
		timing: lipgloss.NewStyle().
			Foreground(themes.colorMuted),
		responseBody: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(themes.colorBorder),
		responseSection: lipgloss.NewStyle().
			Foreground(themes.colorMuted),
		statusBar: lipgloss.NewStyle().
			Foreground(themes.colorMuted).
			Padding(0, 1),
		focusIndicator: lipgloss.NewStyle().
			Foreground(themes.colorPrimary).
			Bold(true),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(themes.colorPrimary),
		loading: lipgloss.NewStyle().
			Foreground(themes.colorMuted).
			Italic(true),
		error: lipgloss.NewStyle().
			Foreground(themes.colorError),
		envBadge: lipgloss.NewStyle().
			Foreground(themes.colorPrimary).
			Bold(true),
		flash: lipgloss.NewStyle().
			Foreground(themes.colorSuccess).
			Bold(true),
	}
}
