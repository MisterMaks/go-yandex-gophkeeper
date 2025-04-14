package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	LoginFormLoginInputIndex = iota
	LoginFormPasswordInputIndex
)

type LoginFormModel struct {
	inputs     []textinput.Model
	focused    int
	isRegister bool
	err        error
	usecase    UsecaseInterface
	homeModel  tea.Model
}

func NewLoginFormModel(usecase UsecaseInterface, homeModel tea.Model) *LoginFormModel {
	inputs := make([]textinput.Model, 2)

	inputs[LoginFormLoginInputIndex] = textinput.New()
	inputs[LoginFormLoginInputIndex].Focus()

	inputs[LoginFormPasswordInputIndex] = textinput.New()
	inputs[LoginFormPasswordInputIndex].EchoMode = textinput.EchoPassword
	inputs[LoginFormPasswordInputIndex].EchoCharacter = '*'

	return &LoginFormModel{
		inputs:    inputs,
		usecase:   usecase,
		homeModel: homeModel,
	}
}

func (m *LoginFormModel) SetRegister(isRegister bool) {
	m.isRegister = isRegister
}

func (m LoginFormModel) Init() tea.Cmd {
	return nil
}

func (m LoginFormModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				login := m.inputs[LoginFormLoginInputIndex].Value()
				password := m.inputs[LoginFormPasswordInputIndex].Value()

				var err error
				if m.isRegister {
					err = m.usecase.Register(ctx, login, password)
				} else {
					err = m.usecase.Login(ctx, login, password)
				}

				if err != nil {
					m.inputs[LoginFormLoginInputIndex].Reset()
					m.inputs[LoginFormLoginInputIndex].Focus()
					m.inputs[LoginFormPasswordInputIndex].Reset()
					m.inputs[LoginFormPasswordInputIndex].Blur()

					m.err = err
					m.prevInput()

					return m, nil
				}

				return m.homeModel, nil
			}
			m.nextInput()
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyShiftTab, tea.KeyCtrlP:
			m.prevInput()
		case tea.KeyTab, tea.KeyCtrlN:
			m.nextInput()
		}
		for i := range m.inputs {
			m.inputs[i].Blur()
		}
		m.inputs[m.focused].Focus()

	case error:
		m.err = msg
		return m, nil
	}

	for i := range m.inputs {
		m.inputs[i], cmds[i] = m.inputs[i].Update(msg)
	}
	return m, tea.Batch(cmds...)
}

func (m LoginFormModel) View() string {
	view := fmt.Sprintf(`Login:
%s
Password:
%s
`,
		m.inputs[LoginFormLoginInputIndex].View(),
		m.inputs[LoginFormPasswordInputIndex].View(),
	) + "\n"

	if m.err != nil {
		view += m.err.Error() + "\n"
	}

	return view
}

// nextInput focuses the next input field
func (m *LoginFormModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *LoginFormModel) prevInput() {
	m.focused--

	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
