package tui

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
)

type Tab int

const (
	TabHeaders Tab = iota
	TabParams
	TabBody
)

var tabNames = []string{"Headers", "Params", "Body"}

func (t Tab) String() string {
	return tabNames[t]
}

type Tabs struct {
	active Tab
}

func NewTabs() *Tabs {
	return &Tabs{
		active: TabHeaders,
	}
}

func (t *Tabs) Next() {
	t.active = (t.active + 1) % Tab(len(tabNames))
}

func (t *Tabs) Prev() {
	if t.active == 0 {
		t.active = Tab(len(tabNames) - 1)
	} else {
		t.active--
	}
}

func (t *Tabs) Active() Tab {
	return t.active
}

func (t *Tabs) View(width int) string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(0, 2)

	var tabs []string

	for i, name := range tabNames {
		if Tab(i) == t.active {
			tabs = append(tabs, activeStyle.Render(name))
		} else {
			tabs = append(tabs, inactiveStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	// Add separator line under tabs
	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

	return fmt.Sprintf("%s%s", row, separator)
}

type ResponseTab int

const (
	ResponseTabBody ResponseTab = iota
	ResponseTabHeaders
)

var responseTabNames = []string{"Body", "Headers"}

func (t ResponseTab) String() string {
	return responseTabNames[t]
}

type ResponseTabs struct {
	active ResponseTab
}

func NewResponseTabs() *ResponseTabs {
	return &ResponseTabs{
		active: ResponseTabBody,
	}
}

func (t *ResponseTabs) Next() {
	t.active = (t.active + 1) % ResponseTab(len(responseTabNames))
}

func (t *ResponseTabs) Prev() {
	if t.active == 0 {
		t.active = ResponseTab(len(responseTabNames) - 1)
	} else {
		t.active--
	}
}

func (t *ResponseTabs) Active() ResponseTab {
	return t.active
}

func (t *ResponseTabs) View(width int) string {
	activeStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("#7D56F4")).
		Border(lipgloss.NormalBorder(), false, false, true, false).
		BorderForeground(lipgloss.Color("#7D56F4")).
		Padding(0, 2)

	inactiveStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#626262")).
		Padding(0, 2)

	var tabs []string
	for i, name := range tabNames {
		if ResponseTab(i) == t.active {
			tabs = append(tabs, activeStyle.Render(name))
		} else {
			tabs = append(tabs, inactiveStyle.Render(name))
		}
	}

	row := lipgloss.JoinHorizontal(lipgloss.Bottom, tabs...)

	separator := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#444444")).
		Render(strings.Repeat("─", max(0, width-lipgloss.Width(row))))

	return fmt.Sprintf("%s%s", row, separator)
}
