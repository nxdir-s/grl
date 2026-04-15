package components

import (
	"fmt"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

const (
	AuthPassword string = "password"
	AuthUsername string = "username"
	AuthToken    string = "token"
	AuthKeyName  string = "X-API-Key"
	AuthKeyValue string = "value"

	AuthCharLimit int = 512

	AuthEchoCharacter rune = '•'
)

type AuthEditorKeyMap struct {
	NextField  key.Binding
	PrevField  key.Binding
	CycleLeft  key.Binding
	CycleRight key.Binding
}

func defaultAuthEditorKeys() AuthEditorKeyMap {
	return AuthEditorKeyMap{
		NextField: key.NewBinding(
			key.WithKeys("ctrl+n"),
			key.WithHelp("ctrl+n", "next field"),
		),
		PrevField: key.NewBinding(
			key.WithKeys("ctrl+p"),
			key.WithHelp("ctrl+p", "prev field"),
		),
		CycleLeft: key.NewBinding(
			key.WithKeys("left"),
			key.WithHelp("←", "prev"),
		),
		CycleRight: key.NewBinding(
			key.WithKeys("right"),
			key.WithHelp("→", "next"),
		),
	}
}

type AuthStyles struct {
	label  lipgloss.Style
	dim    lipgloss.Style
	active lipgloss.Style
}

func NewAuthStyles() *AuthStyles {
	return &AuthStyles{
		label:  lipgloss.NewStyle().Foreground(lipgloss.Color("#7D56F4")).Bold(true),
		dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
		active: lipgloss.NewStyle().Foreground(lipgloss.Color("#E5E5E5")).Bold(true),
	}
}

type AuthEditor struct {
	typeIndex int
	apiKeyIn  valobj.APIKeyLocation

	username textinput.Model
	password textinput.Model
	token    textinput.Model
	keyName  textinput.Model
	keyValue textinput.Model

	focusIdx int
	focused  bool
	width    int

	keys   AuthEditorKeyMap
	styles *AuthStyles

	authTypes []valobj.AuthType
}

func NewAuthEditor() AuthEditor {
	mk := func(placeholder string) textinput.Model {
		ti := textinput.New()
		ti.Placeholder = placeholder
		ti.Prompt = ""
		ti.CharLimit = AuthCharLimit
		return ti
	}

	password := mk(AuthPassword)
	password.EchoMode = textinput.EchoPassword
	password.EchoCharacter = AuthEchoCharacter

	return AuthEditor{
		typeIndex: 0,
		apiKeyIn:  valobj.APIKeyInHeader,
		username:  mk(AuthUsername),
		password:  password,
		token:     mk(AuthToken),
		keyName:   mk(AuthKeyName),
		keyValue:  mk(AuthKeyValue),
		keys:      defaultAuthEditorKeys(),
		styles:    NewAuthStyles(),
		authTypes: []valobj.AuthType{
			valobj.AuthNone,
			valobj.AuthBasic,
			valobj.AuthBearer,
			valobj.AuthAPIKey,
		},
	}
}

func (c *AuthEditor) Focus() tea.Cmd {
	c.focused = true
	c.focusIdx = 0
	c.blurAll()

	return nil
}

func (c *AuthEditor) Blur() {
	c.focused = false
	c.blurAll()
}

func (c *AuthEditor) SetWidth(w int) {
	c.width = w
}

func (c AuthEditor) ToAuth() valobj.Auth {
	return valobj.Auth{
		Type:        c.authTypes[c.typeIndex],
		Username:    c.username.Value(),
		Password:    c.password.Value(),
		Token:       c.token.Value(),
		APIKeyName:  c.keyName.Value(),
		APIKeyValue: c.keyValue.Value(),
		APIKeyIn:    c.apiKeyIn,
	}
}

func (c *AuthEditor) SetAuth(auth valobj.Auth) {
	c.typeIndex = 0

	for i := range c.authTypes {
		if c.authTypes[i] == auth.Type {
			c.typeIndex = i
			break
		}
	}

	c.apiKeyIn = auth.APIKeyIn

	if len(c.apiKeyIn) == 0 {
		c.apiKeyIn = valobj.APIKeyInHeader
	}

	c.username.SetValue(auth.Username)
	c.password.SetValue(auth.Password)
	c.token.SetValue(auth.Token)
	c.keyName.SetValue(auth.APIKeyName)
	c.keyValue.SetValue(auth.APIKeyValue)

	c.focusIdx = 0
	c.blurAll()
}

func (c AuthEditor) currentType() valobj.AuthType {
	return c.authTypes[c.typeIndex]
}

func (c AuthEditor) fieldCount() int {
	switch c.currentType() {
	case valobj.AuthBasic:
		return 3 // type, username, password
	case valobj.AuthBearer:
		return 2 // type, token
	case valobj.AuthAPIKey:
		return 4 // type, name, value, location
	default:
		return 1 // type only
	}
}

func (c *AuthEditor) blurAll() {
	c.username.Blur()
	c.password.Blur()
	c.token.Blur()
	c.keyName.Blur()
	c.keyValue.Blur()
}

// focusField focuses the textinput corresponding to focusIdx (if any).
// focusIdx 0 is the type selector; non-input indices (like location toggle)
// also result in no textinput focused.
func (c *AuthEditor) focusField() tea.Cmd {
	c.blurAll()

	switch c.currentType() {
	case valobj.AuthBasic:
		switch c.focusIdx {
		case 1:
			return c.username.Focus()
		case 2:
			return c.password.Focus()
		}
	case valobj.AuthBearer:
		if c.focusIdx == 1 {
			return c.token.Focus()
		}
	case valobj.AuthAPIKey:
		switch c.focusIdx {
		case 1:
			return c.keyName.Focus()
		case 2:
			return c.keyValue.Focus()
		}
	}

	return nil
}

func (c AuthEditor) Update(msg tea.Msg) (AuthEditor, tea.Cmd) {
	if !c.focused {
		return c, nil
	}

	if km, ok := msg.(tea.KeyPressMsg); ok {
		switch {
		case key.Matches(km, c.keys.NextField):
			n := c.fieldCount()
			c.focusIdx = (c.focusIdx + 1) % n
			return c, c.focusField()
		case key.Matches(km, c.keys.PrevField):
			n := c.fieldCount()
			c.focusIdx = (c.focusIdx - 1 + n) % n

			return c, c.focusField()
		}

		// Type-selector cycling (focusIdx == 0) or location toggle
		if c.focusIdx == 0 {
			switch {
			case key.Matches(km, c.keys.CycleLeft):
				n := len(c.authTypes)

				c.typeIndex = (c.typeIndex - 1 + n) % n
				if c.focusIdx >= c.fieldCount() {
					c.focusIdx = 0
				}

				return c, nil
			case key.Matches(km, c.keys.CycleRight):
				n := len(c.authTypes)

				c.typeIndex = (c.typeIndex + 1) % n
				if c.focusIdx >= c.fieldCount() {
					c.focusIdx = 0
				}

				return c, nil
			}
		}

		// API key location toggle (focusIdx == 3 for APIKey type)
		if c.currentType() == valobj.AuthAPIKey && c.focusIdx == 3 {
			switch {
			case key.Matches(km, c.keys.CycleLeft), key.Matches(km, c.keys.CycleRight):
				switch c.apiKeyIn == valobj.APIKeyInHeader {
				case true:
					c.apiKeyIn = valobj.APIKeyInQuery
				default:
					c.apiKeyIn = valobj.APIKeyInHeader
				}

				return c, nil
			}
		}
	}

	// Forward to the focused textinput
	var cmd tea.Cmd
	switch c.currentType() {
	case valobj.AuthBasic:
		switch c.focusIdx {
		case 1:
			c.username, cmd = c.username.Update(msg)
		case 2:
			c.password, cmd = c.password.Update(msg)
		}
	case valobj.AuthBearer:
		if c.focusIdx == 1 {
			c.token, cmd = c.token.Update(msg)
		}
	case valobj.AuthAPIKey:
		switch c.focusIdx {
		case 1:
			c.keyName, cmd = c.keyName.Update(msg)
		case 2:
			c.keyValue, cmd = c.keyValue.Update(msg)
		}
	}

	return c, cmd
}

func (c AuthEditor) View() string {
	typeLabel := "Type: "
	typeVal := c.currentType().String()

	var typeRow string
	switch {
	case c.focusIdx == 0:
		typeRow = c.styles.label.Render(typeLabel) + c.styles.active.Render("< "+typeVal+" >")
	default:
		typeRow = c.styles.label.Render(typeLabel) + c.styles.dim.Render("< "+typeVal+" >")
	}

	fieldRow := func(label string, val string, idx int) string {
		l := c.styles.label.Render(fmt.Sprintf("%-10s", label))

		if c.focusIdx == idx {
			return l + c.styles.active.Render("▸ "+val)
		}

		return l + c.styles.dim.Render("  "+val)
	}

	lines := []string{typeRow, ""}

	switch c.currentType() {
	case valobj.AuthNone:
		lines = append(lines, c.styles.dim.Render("No authentication will be sent."))
	case valobj.AuthBasic:
		lines = append(lines,
			fieldRow("Username:", c.username.View(), 1),
			fieldRow("Password:", c.password.View(), 2),
		)
	case valobj.AuthBearer:
		lines = append(lines,
			fieldRow("Token:", c.token.View(), 1),
		)
	case valobj.AuthAPIKey:
		locStr := "< " + c.apiKeyIn.String() + " >"
		lines = append(lines,
			fieldRow("Name:", c.keyName.View(), 1),
			fieldRow("Value:", c.keyValue.View(), 2),
			fieldRow("In:", locStr, 3),
		)
	}

	if c.focused {
		lines = append(lines, "", c.styles.dim.Render("ctrl+n/p: field · ←/→: cycle type or location"))
	}

	return lipgloss.JoinVertical(lipgloss.Left, lines...)
}
