package views

import (
	"nfc-write-client/components"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
)

type dialog struct {
	Type        int
	Title       string
	Confirm     string
	Cancel      string
	ConfirmFunc func() tea.Cmd
}
type dialogMsg dialog

func dialogCmd(dialog dialogMsg) tea.Cmd {
	return func() tea.Msg {
		return dialog
	}
}

const (
	DialogNone = iota
	DialogDefault
	DialogAlert
	DialogWait
)

var dialogW, dialogH int

func (d dialog) Init() tea.Cmd {
	return nil
}

func (d dialog) Update(msg tea.Msg) (dialog, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if d.Type != DialogNone {
			if d.Type == DialogDefault && key.Matches(msg, key.NewBinding(key.WithKeys("enter"))) {
				cmds = append(cmds, d.ConfirmFunc())
			} else {
				cmds = append(cmds, dialogCmd(dialogMsg{Type: DialogNone}))
			}
			return d, tea.Batch(cmds...)
		}

	case tea.WindowSizeMsg:
		dialogW, dialogH = msg.Width, msg.Height
	case dialogMsg:
		d.Type = msg.Type
		d.Title = msg.Title
		d.Confirm = msg.Confirm
		d.Cancel = msg.Cancel
		d.ConfirmFunc = msg.ConfirmFunc
	}
	return d, nil
}

func (d dialog) View() string {
	switch d.Type {
	case DialogDefault:
		return components.DialogBox(d.Title, d.Confirm+"(Enter)", d.Cancel+"(Any)", dialogW)
	case DialogAlert:
		return components.Alert(d.Title, d.Confirm+"(Any)", dialogW)
	case DialogWait:
		return components.Wait(d.Title, dialogW)
	}
	return ""
}

func (d dialog) AppendDialog(cxt string) string {
	if d.Type == DialogNone {
		return cxt
	}
	return components.AppendDialog(cxt, d.View(), dialogH)
}

func NewDialog() dialog {
	return dialog{
		Type: DialogNone,
	}
}
