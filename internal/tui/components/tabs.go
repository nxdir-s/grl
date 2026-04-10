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

type RequestTabsStyles struct {
	active    lipgloss.Style
	inactive  lipgloss.Style
	separator lipgloss.Style
}

func NewRequestTabsStyles() *RequestTabsStyles {
	return &RequestTabsStyles{
		active: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 2),
		inactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 2),
		separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")),
	}
}

type RequestTabs struct {
	active   RequestTab
	tabNames []string
	styles   *RequestTabsStyles
}

func NewRequestTabs() RequestTabs {
	return RequestTabs{
		active: RequestTabHeaders,
		tabNames: []string{
			RequestTabHeadersName,
			RequestTabParamsName,
			RequestTabBodyName,
		},
		styles: NewRequestTabsStyles(),
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

func (c RequestTabs) Active() RequestTab {
	return c.active
}

func (c RequestTabs) View(width int) string {
	var tabs []string

	for i := range c.tabNames {
		switch RequestTab(i) == c.active {
		case true:
			tabs = append(tabs, c.styles.active.Render(c.tabNames[i]))
		case false:
			tabs = append(tabs, c.styles.inactive.Render(c.tabNames[i]))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	separator := c.styles.separator.Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

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

type ResponseTabsStyles struct {
	active    lipgloss.Style
	inactive  lipgloss.Style
	separator lipgloss.Style
}

func NewResponseTabsStyles() *ResponseTabsStyles {
	return &ResponseTabsStyles{
		active: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")).
			Border(lipgloss.NormalBorder(), false, false, true, false).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 2),
		inactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Padding(0, 2),
		separator: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#444444")),
	}
}

type ResponseTabs struct {
	active   ResponseTab
	tabNames []string
	styles   *ResponseTabsStyles
}

func NewResponseTabs() ResponseTabs {
	return ResponseTabs{
		active: ResponseTabBody,
		tabNames: []string{
			ResponseTabBodyName,
			ResponseTabHeadersName,
		},
		styles: NewResponseTabsStyles(),
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

func (c ResponseTabs) Active() ResponseTab {
	return c.active
}

func (c ResponseTabs) View(width int) string {
	var tabs []string
	for i := range c.tabNames {
		switch ResponseTab(i) == c.active {
		case true:
			tabs = append(tabs, c.styles.active.Render(c.tabNames[i]))
		case false:
			tabs = append(tabs, c.styles.inactive.Render(c.tabNames[i]))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	separator := c.styles.separator.Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

	return row + separator
}
