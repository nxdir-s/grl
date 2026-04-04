package tui

import (
	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
	"github.com/nxdir-s/grl/internal/core/valobj"
)

type RequestBuilder struct {
	tabs    *Tabs
	headers KVEditor
	params  KVEditor
	body    *RequestBodyEditor

	focused bool
	width   int
	height  int

	keys RequestBuilderKeyMap
}

type RequestBuilderKeyMap struct {
	NextTab key.Binding
	PrevTab key.Binding
}

var defaultBuilderKeys = RequestBuilderKeyMap{
	NextTab: key.NewBinding(
		key.WithKeys("ctrl+x"),
		key.WithHelp("ctrl+x", "next tab"),
	),
	PrevTab: key.NewBinding(
		key.WithKeys("ctrl+q"),
		key.WithHelp("ctrl+q", "prev tab"),
	),
}

func NewRequestBuilder() *RequestBuilder {
	return &RequestBuilder{
		tabs:    NewTabs(),
		headers: NewKVEditor("Header name", "Header value"),
		params:  NewKVEditor("Parameter name", "Parameter value"),
		body:    NewRequestBodyEditor(),
		keys:    defaultBuilderKeys,
	}
}

func (b *RequestBuilder) Focus() tea.Cmd {
	b.focused = true
	return b.focusActiveTab()
}

func (b *RequestBuilder) Blur() {
	b.focused = false
	b.headers.Blur()
	b.params.Blur()
	b.body.Blur()
}

func (b *RequestBuilder) SetSize(width, height int) {
	b.width = width
	b.height = height

	contentHeight := height - 2 // tabs row + separator
	if contentHeight < 1 {
		contentHeight = 1
	}

	b.headers.SetWidth(width)
	b.params.SetWidth(width)
	b.body.SetWidth(width)
	b.body.SetHeight(contentHeight)
}

func (b *RequestBuilder) GetHeaders() []valobj.Header {
	return b.headers.Headers()
}

func (b *RequestBuilder) GetParams() []valobj.QueryParam {
	return b.params.QueryParams()
}

func (b *RequestBuilder) GetBody() string {
	return b.body.Value()
}

func (b *RequestBuilder) ActiveTab() Tab {
	return b.tabs.Active()
}

func (b *RequestBuilder) Update(msg tea.Msg) tea.Cmd {
	if !b.focused {
		return nil
	}

	switch msg := msg.(type) {
	case tea.KeyPressMsg:
		switch {
		case key.Matches(msg, b.keys.NextTab):
			b.blurActiveTab()
			b.tabs.Next()

			return b.focusActiveTab()
		case key.Matches(msg, b.keys.PrevTab):
			b.blurActiveTab()
			b.tabs.Prev()

			return b.focusActiveTab()
		}
	}

	var cmd tea.Cmd
	switch b.tabs.Active() {
	case TabHeaders:
		cmd = b.headers.Update(msg)
	case TabParams:
		cmd = b.params.Update(msg)
	case TabBody:
		cmd = b.body.Update(msg)
	}

	return cmd
}

func (b *RequestBuilder) View() string {
	tabsView := b.tabs.View(b.width)

	var content string
	switch b.tabs.Active() {
	case TabHeaders:
		content = b.headers.View()
	case TabParams:
		content = b.params.View()
	case TabBody:
		content = b.body.View()
	}

	return tabsView + "\n" + content
}

func (b *RequestBuilder) focusActiveTab() tea.Cmd {
	switch b.tabs.Active() {
	case TabHeaders:
		return b.headers.Focus()
	case TabParams:
		return b.params.Focus()
	case TabBody:
		return b.body.Focus()
	}
	return nil
}

func (b *RequestBuilder) blurActiveTab() {
	b.headers.Blur()
	b.params.Blur()
	b.body.Blur()
}

type RequestBodyEditor struct {
	textarea textarea.Model
	focused  bool
}

func NewRequestBodyEditor() *RequestBodyEditor {
	ta := textarea.New()
	ta.Placeholder = "Request body (JSON, XML, text...)"
	ta.ShowLineNumbers = false
	ta.SetHeight(6)

	return &RequestBodyEditor{
		textarea: ta,
	}
}

func (b *RequestBodyEditor) Focus() tea.Cmd {
	b.focused = true
	return b.textarea.Focus()
}

func (b *RequestBodyEditor) Blur() {
	b.focused = false
	b.textarea.Blur()
}

func (b *RequestBodyEditor) SetWidth(w int) {
	b.textarea.SetWidth(w)
}

func (b *RequestBodyEditor) SetHeight(h int) {
	b.textarea.SetHeight(h)
}

func (b *RequestBodyEditor) Value() string {
	return b.textarea.Value()
}

func (b *RequestBodyEditor) Update(msg tea.Msg) tea.Cmd {
	if !b.focused {
		return nil
	}

	var cmd tea.Cmd
	b.textarea, cmd = b.textarea.Update(msg)

	return cmd
}

func (b *RequestBodyEditor) View() string {
	return b.textarea.View()
}
