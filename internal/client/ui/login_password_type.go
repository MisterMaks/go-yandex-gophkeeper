package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	LoginPasswordTypeNameInputIndex = iota
	LoginPasswordTypeLoginInputIndex
	LoginPasswordTypePasswordInputIndex
)

type LoginPasswordTypeModel struct {
	inputs        []textinput.Model
	focused       int
	err           error
	usecase       UsecaseInterface
	homeModel     tea.Model
	typeListModel tea.Model
}

func NewLoginPasswordTypeModel(usecase UsecaseInterface) *LoginPasswordTypeModel {
	inputs := make([]textinput.Model, 3)

	inputs[LoginPasswordTypeNameInputIndex] = textinput.New()
	inputs[LoginPasswordTypeNameInputIndex].Focus()

	inputs[LoginPasswordTypeLoginInputIndex] = textinput.New()

	inputs[LoginPasswordTypePasswordInputIndex] = textinput.New()

	return &LoginPasswordTypeModel{
		inputs:  inputs,
		usecase: usecase,
	}
}

func (m LoginPasswordTypeModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m LoginPasswordTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				name := m.inputs[LoginPasswordTypeNameInputIndex].Value()
				login := m.inputs[LoginPasswordTypeLoginInputIndex].Value()
				password := m.inputs[LoginPasswordTypePasswordInputIndex].Value()

				err := m.usecase.CreateLoginPassword(ctx, name, login, password)

				if err != nil {
					m.inputs[LoginPasswordTypeNameInputIndex].Reset()
					m.inputs[LoginPasswordTypeNameInputIndex].Focus()

					m.inputs[LoginPasswordTypeLoginInputIndex].Reset()
					m.inputs[LoginPasswordTypeLoginInputIndex].Blur()

					m.inputs[LoginPasswordTypePasswordInputIndex].Reset()
					m.inputs[LoginPasswordTypePasswordInputIndex].Blur()

					m.err = err
					m.prevInput()
					m.prevInput()

					return m, nil
				}

				return m.homeModel, nil
			}
			m.nextInput()
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		case tea.KeyEsc:
			return m.typeListModel, nil
		}

		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	// We handle errors just like any other message
	case error:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m LoginPasswordTypeModel) View() string {
	view := fmt.Sprintf(`Name: 
%s
Login:
%s
Password:
%s
`,
		m.inputs[LoginPasswordTypeNameInputIndex].View(),
		m.inputs[LoginPasswordTypeLoginInputIndex].View(),
		m.inputs[LoginPasswordTypePasswordInputIndex].View(),
	) + "\n"

	if m.err != nil {
		view += m.err.Error() + "\n"
	}

	view += "\nesc back\n"

	return view
}

// nextInput focuses the next input field
func (m *LoginPasswordTypeModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *LoginPasswordTypeModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
