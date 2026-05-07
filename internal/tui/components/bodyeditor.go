package components

import (
	"charm.land/bubbles/v2/key"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type BodyEditorKeyMap struct {
	CycleType key.Binding
}

func defaultBodyEditorKeys() BodyEditorKeyMap {
	return BodyEditorKeyMap{
		CycleType: key.NewBinding(
			key.WithKeys("ctrl+b"),
			key.WithHelp("ctrl+b", "cycle body type"),
		),
	}
}

type BodyEditorStyles struct {
	chipActive   lipgloss.Style
	chipInactive lipgloss.Style
	hint         lipgloss.Style
}

func NewBodyEditorStyles() *BodyEditorStyles {
	return &BodyEditorStyles{
		chipActive: lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1),
		chipInactive: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888")).
			Padding(0, 1),
		hint: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#626262")).
			Italic(true),
	}
}

type BodyEditor struct {
	bodyType valobj.BodyType
	raw      RequestBodyEditor
	formURL  KVEditor
	formData KVEditor

	keys   BodyEditorKeyMap
	styles *BodyEditorStyles

	width   int
	height  int
	focused bool
}

func NewBodyEditor() BodyEditor {
	return BodyEditor{
		raw:      NewRequestBodyEditor(),
		formURL:  NewKVEditor("Field name", "Field value"),
		formData: NewKVEditor("Field name", "Value (or @/path/to/file)"),
		keys:     defaultBodyEditorKeys(),
		styles:   NewBodyEditorStyles(),
	}
}

func (c *BodyEditor) Focus() tea.Cmd {
	c.focused = true
	return c.focusActive()
}

func (c *BodyEditor) Blur() {
	c.focused = false
	c.raw.Blur()
	c.formURL.Blur()
	c.formData.Blur()
}

func (c *BodyEditor) SetWidth(w int) {
	c.width = w

	c.raw.SetWidth(w)
	c.formURL.SetWidth(w)
	c.formData.SetWidth(w)
}

func (c *BodyEditor) SetHeight(h int) {
	c.height = h

	// Chip row + spacer occupies 2 rows; sub-editor gets the rest.
	subHeight := h - 2
	if subHeight < 1 {
		subHeight = 1
	}

	c.raw.SetHeight(subHeight)
}

func (c BodyEditor) Value() string {
	return c.raw.Value()
}

func (c *BodyEditor) SetValue(s string) {
	c.raw.SetValue(s)
}

func (c BodyEditor) BodyType() valobj.BodyType {
	return c.bodyType
}

func (c *BodyEditor) SetBodyType(t valobj.BodyType) {
	c.bodyType = t
}

func (c BodyEditor) FormFields() []valobj.FormField {
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		return c.formURL.FormFields()
	case valobj.BodyTypeFormData:
		return c.formData.FormFields()
	default:
		return nil
	}
}

func (c *BodyEditor) SetFormFields(fields []valobj.FormField) {
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		c.formURL.SetFormFields(fields)
	case valobj.BodyTypeFormData:
		c.formData.SetFormFields(fields)
	}
}

func (c BodyEditor) Update(msg tea.Msg) (BodyEditor, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		if key.Matches(msg, c.keys.CycleType) {
			return c.cycleType()
		}
	}

	var cmd tea.Cmd
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		c.formURL, cmd = c.formURL.Update(msg)
	case valobj.BodyTypeFormData:
		c.formData, cmd = c.formData.Update(msg)
	default:
		c.raw, cmd = c.raw.Update(msg)
	}

	return c, cmd
}

func (c BodyEditor) View() string {
	chips := c.chipRow()

	var content string
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		content = c.formURL.View()
	case valobj.BodyTypeFormData:
		content = c.formData.View() + "\n" + c.styles.hint.Render("  tip: prefix value with @ to attach a file")
	default:
		content = c.raw.View()
	}

	return chips + "\n\n" + content
}

func (c BodyEditor) chipRow() string {
	return "  " +
		c.renderChip("Raw", valobj.BodyTypeRaw) + "  " +
		c.renderChip("URL-encoded", valobj.BodyTypeFormURL) + "  " +
		c.renderChip("Form-data", valobj.BodyTypeFormData) + "  " +
		c.styles.hint.Render("(ctrl+b to cycle)")
}

func (c BodyEditor) renderChip(label string, t valobj.BodyType) string {
	if c.bodyType == t {
		return c.styles.chipActive.Render(label)
	}

	return c.styles.chipInactive.Render(label)
}

func (c *BodyEditor) cycleType() (BodyEditor, tea.Cmd) {
	c.blurActive()

	switch c.bodyType {
	case valobj.BodyTypeRaw:
		c.bodyType = valobj.BodyTypeFormURL
	case valobj.BodyTypeFormURL:
		c.bodyType = valobj.BodyTypeFormData
	default:
		c.bodyType = valobj.BodyTypeRaw
	}

	return *c, c.focusActive()
}

func (c *BodyEditor) focusActive() tea.Cmd {
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		return c.formURL.Focus()
	case valobj.BodyTypeFormData:
		return c.formData.Focus()
	default:
		return c.raw.Focus()
	}
}

func (c *BodyEditor) blurActive() {
	switch c.bodyType {
	case valobj.BodyTypeFormURL:
		c.formURL.Blur()
	case valobj.BodyTypeFormData:
		c.formData.Blur()
	default:
		c.raw.Blur()
	}
}
