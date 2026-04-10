package tui

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

type Themes struct {
	colorPrimary    color.Color
	colorSecondary  color.Color
	colorSuccess    color.Color
	colorWarning    color.Color
	colorError      color.Color
	colorMuted      color.Color
	colorText       color.Color
	colorSubtle     color.Color
	colorBackground color.Color
	colorBorder     color.Color
}

func NewThemes() *Themes {
	return &Themes{
		colorPrimary:    lipgloss.Color("#7D56F4"),
		colorSecondary:  lipgloss.Color("#FF6F61"),
		colorSuccess:    lipgloss.Color("#73D216"),
		colorWarning:    lipgloss.Color("#F5A623"),
		colorError:      lipgloss.Color("#FF4444"),
		colorMuted:      lipgloss.Color("#626262"),
		colorText:       lipgloss.Color("#FAFAFA"),
		colorSubtle:     lipgloss.Color("#383838"),
		colorBackground: lipgloss.Color("#1A1A2E"),
		colorBorder:     lipgloss.Color("#444444"),
	}
}
