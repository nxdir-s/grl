package components

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/nxdir-s/grl/internal/core/entity"
)

type Section int

const (
	SectionHistory Section = iota
	SectionCollections
)

type SidebarMode int

const (
	SidebarModeNormal SidebarMode = iota
	SidebarModeRenaming
	SidebarModeConfirmDelete
)

const (
	SidebarHistoryLabel      string = "─ History ─"
	SidebarEmptyHistoryLabel string = "  (empty)"

	SidebarCollectionsLabel      string = "─ Collections ─"
	SidebarEmptyCollectionsLabel string = "  (empty)"
)

type SidebarKeyMap struct {
	Up     key.Binding
	Down   key.Binding
	Select key.Binding
	Rename key.Binding
	Delete key.Binding
	Cancel key.Binding
	Yes    key.Binding
	No     key.Binding
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
		Rename: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "rename"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d"),
			key.WithHelp("d", "delete"),
		),
		Cancel: key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "cancel"),
		),
		Yes: key.NewBinding(
			key.WithKeys("y"),
		),
		No: key.NewBinding(
			key.WithKeys("n"),
		),
	}
}

type SidebarStyles struct {
	title    lipgloss.Style
	header   lipgloss.Style
	muted    lipgloss.Style
	danger   lipgloss.Style
	selected lipgloss.Style
	normal   lipgloss.Style
	border   lipgloss.Style
}

func NewSidebarStyles() *SidebarStyles {
	return &SidebarStyles{
		title:    lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
		header:   lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Bold(true),
		muted:    lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")),
		danger:   lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4444")).Bold(true),
		selected: lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA")).Background(lipgloss.Color("#7D56F4")),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
		border:   lipgloss.NewStyle().Border(lipgloss.RoundedBorder(), false, true, false, false).BorderForeground(lipgloss.Color("#444444")),
	}
}

type Entry struct {
	label              string
	section            Section
	request            *entity.Request
	collection         *entity.Collection
	historyEntry       *entity.HistoryEntry
	parentCollectionID string
}

type Sidebar struct {
	entries []Entry
	cursor  int
	offset  int
	focused bool
	width   int
	height  int

	mode          SidebarMode
	renameInput   textinput.Model
	confirmTarget int

	keys SidebarKeyMap

	styles *SidebarStyles
}

func NewSidebar() Sidebar {
	ti := textinput.New()
	ti.Prompt = ""
	ti.CharLimit = 128

	return Sidebar{
		renameInput: ti,
		keys:        defaultSidebarKeys(),
		styles:      NewSidebarStyles(),
	}
}

func (c *Sidebar) Focus() {
	c.focused = true
}

func (c *Sidebar) Blur() {
	c.focused = false
	c.cancelMode()
}

func (c *Sidebar) SetSize(width int, height int) {
	c.width = width
	c.height = height
}

func (c *Sidebar) SetData(history []entity.HistoryEntry, collections []entity.Collection) {
	c.cancelMode()
	c.entries = make([]Entry, 0)

	c.setCollections(collections)
	c.setHistory(history)

	if c.cursor >= len(c.entries) {
		c.cursor = max(0, len(c.entries)-1)
	}
}

func (c *Sidebar) setCollections(collections []entity.Collection) {
	c.entries = append(c.entries, Entry{
		label:   SidebarCollectionsLabel,
		section: SectionCollections,
	})

	switch len(collections) == 0 {
	case true:
		c.entries = append(c.entries, Entry{
			label:   SidebarEmptyCollectionsLabel,
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
					section:            SectionCollections,
					request:            &collections[i].Requests[j],
					parentCollectionID: collections[i].ID,
				})
			}
		}
	}
}

func (c *Sidebar) setHistory(history []entity.HistoryEntry) {
	c.entries = append(c.entries, Entry{
		label:   SidebarHistoryLabel,
		section: SectionHistory,
	})

	switch len(history) == 0 {
	case true:
		c.entries = append(c.entries, Entry{
			label:   SidebarEmptyHistoryLabel,
			section: SectionHistory,
		})
	default:
		for i := range history {
			c.entries = append(c.entries, Entry{
				label: fmt.Sprintf("%s %s",
					history[i].Request.Method,
					truncate(history[i].Request.URL, c.width-12),
				),
				section:      SectionHistory,
				request:      history[i].Request,
				historyEntry: &history[i],
			})
		}
	}
}

func (c Sidebar) SelectedRequest() *entity.Request {
	if c.cursor < 0 || c.cursor >= len(c.entries) {
		return nil
	}

	return c.entries[c.cursor].request
}

func (c *Sidebar) cancelMode() {
	c.mode = SidebarModeNormal
	c.renameInput.Blur()
	c.renameInput.SetValue("")
}

func (c *Sidebar) currentEntry() *Entry {
	if c.cursor < 0 || c.cursor >= len(c.entries) {
		return nil
	}

	return &c.entries[c.cursor]
}

func (c Sidebar) Update(msg tea.Msg) (Sidebar, tea.Cmd) {
	if !c.focused || len(c.entries) == 0 {
		return c, nil
	}

	km, isKey := msg.(tea.KeyPressMsg)
	if !isKey {
		return c, nil
	}

	switch c.mode {
	case SidebarModeRenaming:
		return c.updateRenaming(km)
	case SidebarModeConfirmDelete:
		return c.updateConfirmDelete(km)
	}

	switch {
	case key.Matches(km, c.keys.Up):
		c.moveCursor(-1)
	case key.Matches(km, c.keys.Down):
		c.moveCursor(1)
	case key.Matches(km, c.keys.Select):
		if req := c.SelectedRequest(); req != nil {
			return c, func() tea.Msg {
				return LoadRequestMsg{
					Request: req,
				}
			}
		}
	case key.Matches(km, c.keys.Rename):
		return c.beginRename()
	case key.Matches(km, c.keys.Delete):
		return c.beginConfirmDelete()
	}

	return c, nil
}

func (c Sidebar) beginRename() (Sidebar, tea.Cmd) {
	e := c.currentEntry()
	if e == nil || e.collection == nil {
		return c, nil
	}

	c.mode = SidebarModeRenaming
	c.confirmTarget = c.cursor
	c.renameInput.SetValue(e.collection.Name)

	return c, c.renameInput.Focus()
}

func (c Sidebar) beginConfirmDelete() (Sidebar, tea.Cmd) {
	e := c.currentEntry()
	if e == nil {
		return c, nil
	}

	if e.collection == nil && e.request == nil && e.historyEntry == nil {
		return c, nil
	}

	c.mode = SidebarModeConfirmDelete
	c.confirmTarget = c.cursor

	return c, nil
}

func (c Sidebar) updateRenaming(km tea.KeyPressMsg) (Sidebar, tea.Cmd) {
	switch {
	case key.Matches(km, c.keys.Cancel):
		c.cancelMode()
		return c, nil
	case key.Matches(km, c.keys.Select):
		newName := strings.TrimSpace(c.renameInput.Value())
		e := &c.entries[c.confirmTarget]

		if len(newName) == 0 || e.collection == nil {
			c.cancelMode()
			return c, nil
		}

		id := e.collection.ID
		c.cancelMode()

		return c, func() tea.Msg {
			return RenameCollectionMsg{
				ID:      id,
				NewName: newName,
			}
		}
	}

	var cmd tea.Cmd
	c.renameInput, cmd = c.renameInput.Update(km)

	return c, cmd
}

func (c Sidebar) updateConfirmDelete(km tea.KeyPressMsg) (Sidebar, tea.Cmd) {
	switch {
	case key.Matches(km, c.keys.Yes):
		e := &c.entries[c.confirmTarget]
		c.cancelMode()

		switch {
		case e.collection != nil:
			return c, func() tea.Msg {
				return DeleteCollectionMsg{
					ID: e.collection.ID,
				}
			}
		case e.request != nil && len(e.parentCollectionID) != 0:
			return c, func() tea.Msg {
				return RemoveRequestMsg{
					CollectionID: e.parentCollectionID,
					RequestID:    e.request.ID,
				}
			}
		case e.historyEntry != nil:
			return c, func() tea.Msg {
				return DeleteHistoryEntryMsg{
					EntryID: e.historyEntry.ID,
				}
			}
		}

		return c, nil
	case key.Matches(km, c.keys.No), key.Matches(km, c.keys.Cancel):
		c.cancelMode()
		return c, nil
	}

	return c, nil
}

func (c *Sidebar) moveCursor(dir int) {
	next := c.cursor + dir

	for next >= 0 && next < len(c.entries) {
		if c.entries[next].request != nil || c.entries[next].collection != nil || c.entries[next].historyEntry != nil {
			break
		}

		next += dir
	}

	if next >= 0 && next < len(c.entries) {
		c.cursor = next
	}

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

	var lines []string
	visible := c.height - 1
	end := min(c.offset+visible, len(c.entries))

	for i := c.offset; i < end; i++ {
		if strings.HasPrefix(c.entries[i].label, "─") {
			lines = append(lines, c.styles.header.Render(c.entries[i].label))
			continue
		}

		if c.entries[i].request == nil && c.entries[i].collection == nil && c.entries[i].historyEntry == nil {
			lines = append(lines, c.styles.muted.Render(c.entries[i].label))
			continue
		}

		if c.mode == SidebarModeRenaming && i == c.confirmTarget {
			lines = append(lines, c.styles.selected.Render("✎ "+c.renameInput.View()))
			continue
		}

		switch {
		case c.focused && i == c.cursor:
			lines = append(lines, c.styles.selected.Width(c.width-2).Render(c.entries[i].label))
		default:
			lines = append(lines, c.styles.normal.Width(c.width-2).Render(c.entries[i].label))
		}
	}

	for len(lines) < visible {
		lines = append(lines, "")
	}

	content := strings.Join(lines, "\n")

	if c.mode == SidebarModeConfirmDelete {
		prompt := c.styles.danger.Render("Delete? (y/n)")
		content += "\n" + prompt
	}

	return c.styles.border.Width(c.width).Height(c.height).Render(content)
}

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
