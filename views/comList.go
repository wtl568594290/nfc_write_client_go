package views

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.bug.st/serial"
)

var (
	// listTitleStyle    = lipgloss.NewStyle().MarginLeft(2)
	itemDirStyle      = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	paginationStyle   = list.DefaultStyles().PaginationStyle.PaddingLeft(4)
	helpStyle         = list.DefaultStyles().HelpStyle.PaddingLeft(4).PaddingBottom(1)
)

type itemCom string
type keyMapCom struct {
	Select  key.Binding
	Refresh key.Binding
}

var _keysShelf = keyMapCom{
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Refresh: key.NewBinding(
		key.WithKeys("r"),
		key.WithHelp("r", "refresh"),
	),
}

func (i itemCom) FilterValue() string { return "" }

type itemComDelegate struct{}

func (d itemComDelegate) Height() int                             { return 1 }
func (d itemComDelegate) Spacing() int                            { return 0 }
func (d itemComDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d itemComDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(itemCom)
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

type modelComList struct {
	list     list.Model
	selected itemCom
}

func (m modelComList) Init() tea.Cmd {
	return nil
}

func (m modelComList) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 1)
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, _keysShelf.Select):
			// 打开
			i, ok := m.list.SelectedItem().(itemCom)
			if ok {
				m.selected = i
				var err error
				Port, err = serial.Open(string(i), &serial.Mode{BaudRate: 115200})
				if err != nil {
					return m, dialogCmd(dialogMsg{
						Type:    DialogAlert,
						Title:   "Open " + string(i) + " failed, " + err.Error(),
						Confirm: "OK",
						ConfirmFunc: func() tea.Cmd {
							return dialogCmd(dialogMsg{Type: DialogNone})
						},
					})
				}
			} else {
				return m, nil
			}
			return m, viewCmd(viewActionList)
		case key.Matches(msg, _keysShelf.Refresh):
			// 刷新
			ports, _ := serial.GetPortsList()
			itemComs := []list.Item{}
			for _, d := range ports {
				itemComs = append(itemComs, itemCom(d))
			}
			cmd := m.list.SetItems(itemComs)
			m.list.ResetSelected()
			return m, cmd
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m modelComList) View() string {
	return "\n" + m.list.View()
}

func NewComList() modelComList {
	ports, _ := serial.GetPortsList()
	itemComs := []list.Item{}
	for _, d := range ports {
		itemComs = append(itemComs, itemCom(d))
	}
	l := list.New(itemComs, itemComDelegate{}, 20, 20)
	l.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{_keysShelf.Select, _keysShelf.Refresh}
	}
	l.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{_keysShelf.Select, _keysShelf.Refresh}
	}
	l.Title = "Select a port"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	l.Styles.PaginationStyle = paginationStyle
	l.Styles.HelpStyle = helpStyle

	m := modelComList{list: l}

	return m
}
