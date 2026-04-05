package components

import (
	"strings"

	"charm.land/lipgloss/v2"
)

type RequestTab int

const (
	RequestTabHeaders RequestTab = iota
	RequestTabParams
	RequestTabBody
)

const (
	RequestTabHeadersName string = "Headers"
	RequestTabParamsName  string = "Params"
	RequestTabBodyName    string = "Body"
)

func (c RequestTab) String() string {
	switch c {
	case RequestTabHeaders:
		return RequestTabHeadersName
	case RequestTabParams:
		return RequestTabParamsName
	case RequestTabBody:
		return RequestTabBodyName
	default:
		return ""
	}
}

type RequestTabs struct {
	active   RequestTab
	tabNames []string

	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
}

func NewRequestTabs() *RequestTabs {
	return &RequestTabs{
		active: RequestTabHeaders,
		tabNames: []string{
			RequestTabHeadersName,
			RequestTabParamsName,
			RequestTabBodyName,
		},
		activeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 2),
		inactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 2),
	}
}

func (c *RequestTabs) Next() {
	c.active = (c.active + 1) % RequestTab(len(c.tabNames))
}

func (c *RequestTabs) Prev() {
	if c.active == 0 {
		c.active = RequestTab(len(c.tabNames) - 1)
	} else {
		c.active--
	}
}

func (c *RequestTabs) Active() RequestTab {
	return c.active
}

func (c *RequestTabs) View(width int) string {
	var tabs []string

	for i := range c.tabNames {
		switch RequestTab(i) == c.active {
		case true:
			tabs = append(tabs, c.activeStyle.Render(c.tabNames[i]))
		case false:
			tabs = append(tabs, c.inactiveStyle.Render(c.tabNames[i]))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

	return row + separator
}

type ResponseTab int

const (
	ResponseTabBody ResponseTab = iota
	ResponseTabHeaders
)

const (
	ResponseTabBodyName    string = "Body"
	ResponseTabHeadersName string = "Headers"
)

func (c ResponseTab) String() string {
	switch c {
	case ResponseTabBody:
		return ResponseTabBodyName
	case ResponseTabHeaders:
		return ResponseTabHeadersName
	default:
		return ""
	}
}

type ResponseTabs struct {
	active   ResponseTab
	tabNames []string

	activeStyle   lipgloss.Style
	inactiveStyle lipgloss.Style
}

func NewResponseTabs() *ResponseTabs {
	return &ResponseTabs{
		active: ResponseTabBody,
		tabNames: []string{
			ResponseTabBodyName,
			ResponseTabHeadersName,
		},
		activeStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 2),
		inactiveStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 2),
	}
}

func (c *ResponseTabs) Next() {
	c.active = (c.active + 1) % ResponseTab(len(c.tabNames))
}

func (c *ResponseTabs) Prev() {
	if c.active == 0 {
		c.active = ResponseTab(len(c.tabNames) - 1)
	} else {
		c.active--
	}
}

func (c *ResponseTabs) Active() ResponseTab {
	return c.active
}

func (c *ResponseTabs) View(width int) string {
	var tabs []string
	for i := range c.tabNames {
		switch ResponseTab(i) == c.active {
		case true:
			tabs = append(tabs, c.activeStyle.Render(c.tabNames[i]))
		case false:
			tabs = append(tabs, c.inactiveStyle.Render(c.tabNames[i]))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

	return row + separator
}
