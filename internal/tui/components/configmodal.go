package components

import (
	"fmt"
	"strconv"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/nxdir-s/grl/internal/core/valobj"
)

type ConfigField int

const (
	ConfigFieldMethod ConfigField = iota
	ConfigFieldTimeout
	ConfigFieldRedirects
	ConfigFieldHistoryLimit
	ConfigFieldCount
)

const (
	ConfigFieldMethodName       string = "Default Method"
	ConfigFieldTimeoutName      string = "Timeout (seconds)"
	ConfigFieldRedirectsName    string = "Follow Redirects"
	ConfigFieldHistoryLimitName string = "History Limit"
)

type ConfigRow struct {
	Field ConfigField
	Name  string
	Value string
}

const (
	DefaultRedirectsVal string = "no"

	ConfigModalTitle string = "Settings"
	ConfigModalCmds  string = "ctrl+n/p: nav · enter: toggle/cycle · ctrl+s: save · esc: cancel"
	ConfigModalWidth int    = 60

	ConfigSavedAlert string = "settings saved · restart for timeout/redirect changes"

	FollowRedirects string = "yes"
)

type ConfigModalKeyMap struct {
	Next   key.Binding
	Prev   key.Binding
	Toggle key.Binding
	Save   key.Binding
	Close  key.Binding
}

func defaultConfigModalKeys() ConfigModalKeyMap {
	return ConfigModalKeyMap{
		Next:   key.NewBinding(key.WithKeys("ctrl+n", "down"), key.WithHelp("ctrl+n", "next")),
		Prev:   key.NewBinding(key.WithKeys("ctrl+p", "up"), key.WithHelp("ctrl+p", "prev")),
		Toggle: key.NewBinding(key.WithKeys("enter", " "), key.WithHelp("enter", "toggle/cycle")),
		Save:   key.NewBinding(key.WithKeys("ctrl+s"), key.WithHelp("ctrl+s", "save")),
		Close:  key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
	}
}

type ConfigModalStyles struct {
	border   lipgloss.Style
	title    lipgloss.Style
	sel      lipgloss.Style
	normal   lipgloss.Style
	label    lipgloss.Style
	cmdsHelp lipgloss.Style
}

func NewConfigModalStyles() *ConfigModalStyles {
	return &ConfigModalStyles{
		border: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(1, 2).
			Width(ConfigModalWidth),
		title: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4")),
		sel: lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")),
		normal:   lipgloss.NewStyle().Foreground(lipgloss.Color("#AAAAAA")),
		label:    lipgloss.NewStyle().Foreground(lipgloss.Color("#F5A623")),
		cmdsHelp: lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")),
	}
}

type ConfigModal struct {
	open      bool
	width     int
	height    int
	cursor    ConfigField
	cfg       *valobj.Config
	methodId  int
	timeoutIn textinput.Model
	historyIn textinput.Model

	keys    ConfigModalKeyMap
	methods []valobj.HTTPMethod

	styles *ConfigModalStyles
}

func NewConfigModal() ConfigModal {
	timeoutIn := textinput.New()
	timeoutIn.Placeholder = "30"
	timeoutIn.Prompt = ""
	timeoutIn.CharLimit = 6

	historyIn := textinput.New()
	historyIn.Placeholder = "100"
	historyIn.Prompt = ""
	historyIn.CharLimit = 6

	return ConfigModal{
		timeoutIn: timeoutIn,
		historyIn: historyIn,
		keys:      defaultConfigModalKeys(),
		methods: []valobj.HTTPMethod{
			valobj.MethodGet,
			valobj.MethodPost,
			valobj.MethodPut,
			valobj.MethodDelete,
			valobj.MethodPatch,
			valobj.MethodHead,
			valobj.MethodOptions,
		},
		styles: NewConfigModalStyles(),
	}
}

type SaveConfigMsg struct {
	Cfg *valobj.Config
}

func (m *ConfigModal) Open(cfg *valobj.Config) tea.Cmd {
	m.open = true

	c := *cfg

	m.cfg = &c
	m.cursor = ConfigFieldMethod
	m.methodId = m.methodIndex(m.cfg.DefaultMethod)

	m.timeoutIn.SetValue(strconv.Itoa(m.cfg.TimeoutSeconds))
	m.historyIn.SetValue(strconv.Itoa(m.cfg.HistoryLimit))

	m.timeoutIn.Blur()
	m.historyIn.Blur()

	return nil
}

func (m *ConfigModal) Close() {
	m.open = false
	m.cfg = nil
	m.timeoutIn.Blur()
	m.historyIn.Blur()
}

func (m ConfigModal) IsOpen() bool {
	return m.open
}

func (m *ConfigModal) SetSize(w, h int) {
	m.width = w
	m.height = h
}

func (m ConfigModal) methodIndex(s string) int {
	for i := range m.methods {
		if string(m.methods[i]) == s {
			return i
		}
	}

	return 0
}

func (m ConfigModal) Update(msg tea.Msg) (ConfigModal, tea.Cmd) {
	if !m.open {
		return m, nil
	}

	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		var cmd tea.Cmd
		switch m.cursor {
		case ConfigFieldTimeout:
			if m.timeoutIn.Focused() {
				m.timeoutIn, cmd = m.timeoutIn.Update(msg)
			}
		case ConfigFieldHistoryLimit:
			if m.historyIn.Focused() {
				m.historyIn, cmd = m.historyIn.Update(msg)
			}
		}

		return m, cmd
	}

	switch {
	case key.Matches(km, m.keys.Close):
		m.Close()
		return m, nil
	case key.Matches(km, m.keys.Save):
		if err := m.commitInputs(); err != nil {
			return m, nil
		}

		cfg := m.cfg
		m.Close()

		return m, func() tea.Msg {
			return SaveConfigMsg{
				Cfg: cfg,
			}
		}
	case key.Matches(km, m.keys.Next):
		return m.moveCursor(1), m.refocus()
	case key.Matches(km, m.keys.Prev):
		return m.moveCursor(-1), m.refocus()
	case key.Matches(km, m.keys.Toggle):
		switch m.cursor {
		case ConfigFieldMethod:
			m.methodId = (m.methodId + 1) % len(m.methods)
			m.cfg.DefaultMethod = string(m.methods[m.methodId])
		case ConfigFieldRedirects:
			m.cfg.FollowRedirects = !m.cfg.FollowRedirects
		}

		return m, nil
	}

	var cmd tea.Cmd
	switch m.cursor {
	case ConfigFieldTimeout:
		m.timeoutIn, cmd = m.timeoutIn.Update(msg)
	case ConfigFieldHistoryLimit:
		m.historyIn, cmd = m.historyIn.Update(msg)
	}

	return m, cmd
}

func (m ConfigModal) moveCursor(dir int) ConfigModal {
	m.commitInputs()
	m.cursor = ConfigField((int(m.cursor) + dir + int(ConfigFieldCount)) % int(ConfigFieldCount))

	return m
}

func (m *ConfigModal) refocus() tea.Cmd {
	m.timeoutIn.Blur()
	m.historyIn.Blur()

	switch m.cursor {
	case ConfigFieldTimeout:
		return m.timeoutIn.Focus()
	case ConfigFieldHistoryLimit:
		return m.historyIn.Focus()
	}

	return nil
}

func (m *ConfigModal) commitInputs() error {
	timeout, err := strconv.Atoi(strings.TrimSpace(m.timeoutIn.Value()))
	if err != nil {
		return err
	}

	if timeout > 0 {
		m.cfg.TimeoutSeconds = timeout
	}

	historyLimit, err := strconv.Atoi(strings.TrimSpace(m.historyIn.Value()))
	if err != nil {
		return err
	}

	if historyLimit > 0 {
		m.cfg.HistoryLimit = historyLimit
	}

	return nil
}

func (m ConfigModal) View() string {
	if !m.open || m.cfg == nil {
		return ""
	}

	var redirectsVal string = DefaultRedirectsVal
	if m.cfg.FollowRedirects {
		redirectsVal = FollowRedirects
	}

	rows := m.buildRows(redirectsVal)

	var b strings.Builder
	b.WriteString(m.styles.title.Render(ConfigModalTitle))
	b.WriteString("\n\n")

	for i := range rows {
		line := fmt.Sprintf("%-20s %s", m.styles.label.Render(rows[i].Name), rows[i].Value)

		switch m.cursor == rows[i].Field {
		case true:
			b.WriteString(m.styles.sel.Render(fmt.Sprintf("▸ %-50s", line)))
		default:
			b.WriteString(m.styles.normal.Render("  " + line))
		}

		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(m.styles.cmdsHelp.Render(ConfigModalCmds))

	return m.styles.border.Render(b.String())
}

func (m ConfigModal) buildRows(redirects string) []ConfigRow {
	return []ConfigRow{
		{
			Field: ConfigFieldMethod,
			Name:  ConfigFieldMethodName,
			Value: m.cfg.DefaultMethod,
		},
		{
			Field: ConfigFieldTimeout,
			Name:  ConfigFieldTimeoutName,
			Value: m.timeoutIn.View(),
		},
		{
			Field: ConfigFieldRedirects,
			Name:  ConfigFieldRedirectsName,
			Value: redirects,
		},
		{
			Field: ConfigFieldHistoryLimit,
			Name:  ConfigFieldHistoryLimitName,
			Value: m.historyIn.View(),
		},
	}
}
