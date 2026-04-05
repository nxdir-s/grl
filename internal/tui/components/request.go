package components

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type RequestBuilderKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
}

func defaultBuilderKeys() RequestBuilderKeyMap {
	return RequestBuilderKeyMap{
		NextTab: key.NewBinding(
			key.WithKeys("ctrl+x"),
			key.WithHelp("ctrl+x", "next tab"),
		),
		PrevTab: key.NewBinding(
			key.WithKeys("ctrl+q"),
			key.WithHelp("ctrl+q", "prev tab"),
		),
	}
}

type RequestBuilder struct {
	tabs    *RequestTabs
	headers KVEditor
	params  KVEditor
	body    *RequestBodyEditor

	focused bool
	width   int
	height  int

	keys RequestBuilderKeyMap
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		tabs:    NewRequestTabs(),
		headers: NewKVEditor("Header name", "Header value"),
		params:  NewKVEditor("Parameter name", "Parameter value"),
		body:    NewRequestBodyEditor(),
		keys:    defaultBuilderKeys(),
	}
}

func (c *RequestBuilder) Focus() tea.Cmd {
	c.focused = true
	return c.focusActiveTab()
}

func (c *RequestBuilder) Blur() {
	c.focused = false
	c.headers.Blur()
	c.params.Blur()
	c.body.Blur()
}

func (c *RequestBuilder) SetSize(width, height int) {
	c.width = width
	c.height = height

	contentHeight := height - 2 // tabs row + separator
	if contentHeight < 1 {
		contentHeight = 1
	}

	c.headers.SetWidth(width)
	c.params.SetWidth(width)
	c.body.SetWidth(width)
	c.body.SetHeight(contentHeight)
}

func (c *RequestBuilder) GetHeaders() []valobj.Header {
	return c.headers.Headers()
}

func (c *RequestBuilder) GetParams() []valobj.QueryParam {
	return c.params.QueryParams()
}

func (c *RequestBuilder) GetBody() string {
	return c.body.Value()
}

func (c *RequestBuilder) ActiveTab() RequestTab {
	return c.tabs.Active()
}

func (c *RequestBuilder) Update(msg tea.Msg) tea.Cmd {
	if !c.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, c.keys.NextTab):
			c.blurActiveTab()
			c.tabs.Next()

			return c.focusActiveTab()
		case key.Matches(msg, c.keys.PrevTab):
			c.blurActiveTab()
			c.tabs.Prev()

			return c.focusActiveTab()
		}
	}

	var cmd tea.Cmd
	switch c.tabs.Active() {
	case RequestTabHeaders:
		cmd = c.headers.Update(msg)
	case RequestTabParams:
		cmd = c.params.Update(msg)
	case RequestTabBody:
		cmd = c.body.Update(msg)
	}

	return cmd
}

func (c *RequestBuilder) View() string {
	tabsView := c.tabs.View(c.width)

	var content string
	switch c.tabs.Active() {
	case RequestTabHeaders:
		content = c.headers.View()
	case RequestTabParams:
		content = c.params.View()
	case RequestTabBody:
		content = c.body.View()
	}

	return tabsView + "\n" + content
}

func (c *RequestBuilder) focusActiveTab() tea.Cmd {
	switch c.tabs.Active() {
	case RequestTabHeaders:
		return c.headers.Focus()
	case RequestTabParams:
		return c.params.Focus()
	case RequestTabBody:
		return c.body.Focus()
	default:
		return nil
	}
}

func (c *RequestBuilder) blurActiveTab() {
	c.headers.Blur()
	c.params.Blur()
	c.body.Blur()
}

type RequestBodyEditor struct {
	textarea textarea.Model
	focused  bool
}

func NewRequestBodyEditor() *RequestBodyEditor {
	text := textarea.New()

	text.Placeholder = "Request body (JSON, XML, text...)"
	text.ShowLineNumbers = false
	text.SetHeight(6)

	return &RequestBodyEditor{
		textarea: text,
	}
}

func (c *RequestBodyEditor) Focus() tea.Cmd {
	c.focused = true
	return c.textarea.Focus()
}

func (c *RequestBodyEditor) Blur() {
	c.focused = false
	c.textarea.Blur()
}

func (c *RequestBodyEditor) SetWidth(w int) {
	c.textarea.SetWidth(w)
}

func (c *RequestBodyEditor) SetHeight(h int) {
	c.textarea.SetHeight(h)
}

func (c *RequestBodyEditor) Value() string {
	return c.textarea.Value()
}

func (c *RequestBodyEditor) Update(msg tea.Msg) tea.Cmd {
	if !c.focused {
		return nil
	}

	var cmd tea.Cmd
	c.textarea, cmd = c.textarea.Update(msg)

	return cmd
}

func (c *RequestBodyEditor) View() string {
	return c.textarea.View()
}
