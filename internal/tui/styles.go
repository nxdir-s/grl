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

func NewStyles(theme *Theme) *Styles {
	return &Styles{
		base: lipgloss.NewStyle(),
		method: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorPrimary).
			Padding(0, 1),
		urlBar: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.colorBorder).
			Padding(0, 1),
		statusSuccess: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorSuccess),
		statusRedirect: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorWarning),
		statusClientErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorSecondary),
		statusServerErr: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorError),
		timing: lipgloss.NewStyle().
			Foreground(theme.colorMuted),
		responseBody: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.colorBorder),
		responseSection: lipgloss.NewStyle().
			Foreground(theme.colorMuted),
		statusBar: lipgloss.NewStyle().
			Foreground(theme.colorMuted).
			Padding(0, 1),
		focusIndicator: lipgloss.NewStyle().
			Foreground(theme.colorPrimary).
			Bold(true),
		title: lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.colorPrimary),
		loading: lipgloss.NewStyle().
			Foreground(theme.colorMuted).
			Italic(true),
		error: lipgloss.NewStyle().
			Foreground(theme.colorError),
		envBadge: lipgloss.NewStyle().
			Foreground(theme.colorPrimary).
			Bold(true),
		flash: lipgloss.NewStyle().
			Foreground(theme.colorSuccess).
			Bold(true),
	}
}
