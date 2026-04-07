package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nxdir-s/grl/internal/core/entity"
)

type Section int

const (
	SectionHistory Section = iota
	SectionCollections
)

type SidebarKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
}

func defaultSidebarKeys() SidebarKeyMap {
	return SidebarKeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "ctrl+p"),
			key.WithHelp("↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "ctrl+n"),
			key.WithHelp("↓", "down"),
		),
		Select: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
	}
}

type Entry struct {
	label      string
	section    Section
	request    *entity.Request
	collection *entity.Collection
}

type Sidebar struct {
	entries []Entry
	cursor  int
	offset  int
	focused bool
	width   int
	height  int

	titleStyle    lipgloss.Style
	headerStyle   lipgloss.Style
	selectedStyle lipgloss.Style
	normalStyle   lipgloss.Style
	mutedStyle    lipgloss.Style
	borderStyle   lipgloss.Style

	keys SidebarKeyMap
}

func NewSidebar() Sidebar {
	return Sidebar{
		keys: defaultSidebarKeys(),
		titleStyle: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#7D56F4")),
		headerStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Bold(true),
		mutedStyle: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#555555")),
	}
}

func (c *Sidebar) Focus() {
	c.focused = true
}

func (c *Sidebar) Blur() {
	c.focused = false
}

func (c *Sidebar) SetSize(width int, height int) {
	c.width = width
	c.height = height
}

func (c *Sidebar) SetData(history []entity.HistoryEntry, collections []entity.Collection) {
	c.entries = make([]Entry, 0)

	c.setCollections(collections)
	c.setHistory(history)

	// Keep cursor in bounds
	if c.cursor >= len(c.entries) {
		c.cursor = max(0, len(c.entries)-1)
	}
}

func (c *Sidebar) setCollections(collections []entity.Collection) {
	c.entries = append(c.entries, Entry{
		label:   "─ Collections ─",
		section: SectionCollections,
	})

	switch len(collections) == 0 {
	case true:
		c.entries = append(c.entries, Entry{
			label:   "  (empty)",
			section: SectionCollections,
		})
	default:
		for i := range collections {
			c.entries = append(c.entries, Entry{
				label:      fmt.Sprintf("▸ %s (%d)", collections[i].Name, len(collections[i].Requests)),
				section:    SectionCollections,
				collection: &collections[i],
			})

			for j := range collections[i].Requests {
				c.entries = append(c.entries, Entry{
					label: fmt.Sprintf("  %s %s",
						collections[i].Requests[j].Method,
						truncate(collections[i].Requests[j].URL, c.width-14),
					),
					section: SectionCollections,
					request: &collections[i].Requests[j],
				})
			}
		}
	}
}

func (c *Sidebar) setHistory(history []entity.HistoryEntry) {
	c.entries = append(c.entries, Entry{
		label:   "─ History ─",
		section: SectionHistory,
	})

	switch len(history) == 0 {
	case true:
		c.entries = append(c.entries, Entry{
			label:   "  (empty)",
			section: SectionHistory,
		})
	default:
		for i := range history {
			c.entries = append(c.entries, Entry{
				label: fmt.Sprintf("%s %s",
					history[i].Request.Method,
					truncate(history[i].Request.URL, c.width-12),
				),
				section: SectionHistory,
				request: history[i].Request,
			})
		}
	}
}

// SelectedRequest returns the request at the current cursor, or nil if it's a section header.
func (c Sidebar) SelectedRequest() *entity.Request {
	if c.cursor < 0 || c.cursor >= len(c.entries) {
		return nil
	}

	return c.entries[c.cursor].request
}

func (c Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	if !c.focused || len(c.entries) == 0 {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keys.Up):
			c.moveCursor(-1)
			return c, nil
		case key.Matches(msg, c.keys.Down):
			c.moveCursor(1)
			return c, nil
		case key.Matches(msg, c.keys.Select):
			if req := c.SelectedRequest(); req != nil {
				return c, func() tea.Msg {
					return NewLoadRequestMsg(req)
				}
			}

			return c, nil
		}
	}

	return c, nil
}

func (c *Sidebar) moveCursor(dir int) {
	next := c.cursor + dir

	// Skip section headers
	for next >= 0 && next < len(c.entries) {
		if c.entries[next].request != nil || c.entries[next].collection != nil {
			break
		}

		next += dir
	}

	if next >= 0 && next < len(c.entries) {
		c.cursor = next
	}

	// Scroll viewport
	visible := c.height - 2
	if visible < 1 {
		visible = 1
	}

	if c.cursor < c.offset {
		c.offset = c.cursor
	}

	if c.cursor >= c.offset+visible {
		c.offset = c.cursor - visible + 1
	}
}

func (c Sidebar) View() string {
	if c.width == 0 || c.height == 0 {
		return ""
	}

	c.setSelectedStyle()
	c.setNormalStyle()

	var lines []string
	visible := c.height - 1
	end := min(c.offset+visible, len(c.entries))

	for i := c.offset; i < end; i++ {
		// Section headers
		if strings.HasPrefix(c.entries[i].label, "─") {
			lines = append(lines, c.headerStyle.Render(c.entries[i].label))
			continue
		}

		// Muted entries like "(empty)"
		if c.entries[i].request == nil && c.entries[i].collection == nil {
			lines = append(lines, c.mutedStyle.Render(c.entries[i].label))
			continue
		}

		if c.focused && i == c.cursor {
			lines = append(lines, c.selectedStyle.Render(c.entries[i].label))
		} else {
			lines = append(lines, c.normalStyle.Render(c.entries[i].label))
		}
	}

	// Pad to fill height
	for len(lines) < visible {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	c.setBorderStyle()

	return c.borderStyle.Render(content)
}

func (c *Sidebar) setSelectedStyle() {
	c.selectedStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Width(c.width - 2)
}

func (c *Sidebar) setNormalStyle() {
	c.normalStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#AAAAAA")).
		Width(c.width - 2)
}

func (c *Sidebar) setBorderStyle() {
	c.borderStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), false, true, false, false).
		BorderForeground(lipgloss.Color("#444444")).
		Width(c.width).
		Height(c.height)
}

// LoadRequestMsg is sent when the user selects a request from the sidebar.
type LoadRequestMsg struct {
	Request *entity.Request
}

func NewLoadRequestMsg(req *entity.Request) LoadRequestMsg {
	return LoadRequestMsg{
		Request: req,
	}
}

func truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}

	if len(s) <= maxLen {
		return s
	}

	if maxLen <= 3 {
		return s[:maxLen]
	}

	return s[:maxLen-3] + "..."
}
