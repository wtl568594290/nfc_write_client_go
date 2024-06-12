package views

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type itemAction string
type keyMapAction struct {
	Quit   key.Binding
	Select key.Binding
}

const (
	ACTION_INIT     = "Init"
	ACTION_RECHARGE = "Recharge"
	ACTION_BALANCE  = "Balance"
	ACTION_FORMAT   = "Format"
	ACTION_GETWIFI  = "Getwifi"
	ACTION_GETIP    = "Getip"
	ACTION_SETWIFI  = "Setwifi"
)

var _keysAction = keyMapAction{
	Quit: key.NewBinding(
		key.WithKeys("q"),
		key.WithHelp("q", "quit"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
}
var _inputStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#00CC00"))

func (i itemAction) FilterValue() string { return "" }

type itemActionDelegate struct{}

func (d itemActionDelegate) Height() int                             { return 1 }
func (d itemActionDelegate) Spacing() int                            { return 0 }
func (d itemActionDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemActionDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(itemAction)
	if !ok {
		return
	}

	str := string(i)

	fn := itemDirStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("◉ " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}

type modelActionList struct {
	list     list.Model
	input    textinput.Model
	selected itemAction
}

func (m modelActionList) Init() tea.Cmd {
	return nil
}

func (m modelActionList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	isListUpdate := true
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 1)
		return m, nil

	case tea.KeyMsg:
		// 当input focus时,除了up down enter外,m.list不响应其他按键
		if m.input.Focused() {
			isListUpdate = false
			switch msg.String() {
			case "up", "down", "enter":
				isListUpdate = true
			case "esc":
				m.input.SetValue("")
			}
		}
		switch {
		case key.Matches(msg, _keysAction.Select):
			// 打开
			i, ok := m.list.SelectedItem().(itemAction)
			if ok {
				m.selected = i
				action := strings.ToLower(string(i))
				switch string(i) {
				case ACTION_RECHARGE:
					// 判断 input 是否为数字且在1-9999之间
					amountStr := m.input.Value()
					if amountStr == "" {
						return m, dialogCmd(dialogMsg{
							Type:    DialogAlert,
							Title:   "Please input recharge amount",
							Confirm: "OK",
						})
					}
					amount, err := strconv.Atoi(amountStr)
					if err != nil {
						return m, dialogCmd(dialogMsg{
							Type:    DialogAlert,
							Title:   "Recharge amount must be a number",
							Confirm: "OK",
						})
					}
					if amount < 1 || amount > 9999 {
						return m, dialogCmd(dialogMsg{
							Type:    DialogAlert,
							Title:   "Recharge amount must be between 1-9999",
							Confirm: "OK",
						})
					}
					action += ":" + m.input.Value()
				case ACTION_SETWIFI:
					// 判断 input 是否为 ssid,passwd 模式
					ssidPasswd := strings.Split(m.input.Value(), ",")
					if len(ssidPasswd) != 2 {
						return m, dialogCmd(dialogMsg{
							Type:    DialogAlert,
							Title:   "Please input ssid,passwd",
							Confirm: "OK",
						})
					}
					action += ":" + m.input.Value()
				}
				res, err := PortWriteAndRead(action, 5000)
				if err != nil {
					return m, dialogCmd(dialogMsg{
						Type:    DialogAlert,
						Title:   "Execute " + string(i) + " failed, " + err.Error(),
						Confirm: "OK",
					})
				}
				return m, dialogCmd(dialogMsg{
					Type:    DialogAlert,
					Title:   res,
					Confirm: "OK",
				})
			}
			return m, nil
		case key.Matches(msg, _keysAction.Quit):
			if !m.input.Focused() {
				Port.Close()
				Port = nil
				return m, viewCmd(viewComList)
			}
		}

	}

	if isListUpdate {
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}
	m.input, cmd = m.input.Update(msg)
	cmds = append(cmds, cmd)
	if i, ok := m.list.SelectedItem().(itemAction); ok && (string(i) == ACTION_RECHARGE || string(i) == ACTION_SETWIFI) {
		switch string(i) {
		case ACTION_RECHARGE:
			m.input.Placeholder = "Recharge amount"
		case ACTION_SETWIFI:
			m.input.Placeholder = "SSID,Passwd"
		}
		m.input.Focus()
		m.list.SetShowHelp(false)
	} else {
		m.input.Blur()
		m.list.SetShowHelp(true)
	}
	return m, tea.Batch(cmds...)
}

func (m modelActionList) View() string {
	in := ""
	if m.input.Focused() {
		in = lipgloss.NewStyle().PaddingLeft(2).Render(m.input.View())
	}
	return in + "\n" + m.list.View()
}

func NewActionList() modelActionList {
	actions := []string{ACTION_INIT, ACTION_RECHARGE, ACTION_BALANCE, ACTION_FORMAT, ACTION_GETIP, ACTION_GETWIFI, ACTION_SETWIFI}
	itemActions := []list.Item{}
	for _, d := range actions {
		itemActions = append(itemActions, itemAction(d))
	}
	l := list.New(itemActions, itemActionDelegate{}, 20, 20)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{_keysAction.Select}
	}
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{_keysAction.Select}
	}
	l.Title = "Choose an action"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	input := textinput.New()
	// input.Placeholder = "Recharge amount"
	input.TextStyle = _inputStyle
	input.Cursor.Style = _inputStyle
	input.PromptStyle = _inputStyle
	m := modelActionList{list: l, input: input}

	return m
}
