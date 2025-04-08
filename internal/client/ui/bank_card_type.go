package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	BankCardTypeNameInputIndex = iota
	BankCardTypeNumberInputIndex
	BankCardTypeExpirationDateInputIndex
	BankCardTypeSecurityCodeInputIndex
)

type BankCardTypeModel struct {
	inputs        []textinput.Model
	focused       int
	err           error
	usecase       UsecaseInterface
	homeModel     tea.Model
	typeListModel tea.Model
}

func NewBankCardTypeModel(usecase UsecaseInterface) *BankCardTypeModel {
	inputs := make([]textinput.Model, 4)

	inputs[BankCardTypeNameInputIndex] = textinput.New()
	inputs[BankCardTypeNameInputIndex].Focus()

	inputs[BankCardTypeNumberInputIndex] = textinput.New()
	inputs[BankCardTypeNumberInputIndex].Placeholder = "4505 **** **** 1234"
	inputs[BankCardTypeNumberInputIndex].CharLimit = 16
	inputs[BankCardTypeNumberInputIndex].Width = 16

	inputs[BankCardTypeExpirationDateInputIndex] = textinput.New()
	inputs[BankCardTypeExpirationDateInputIndex].Placeholder = "10/11"
	inputs[BankCardTypeExpirationDateInputIndex].CharLimit = 5
	inputs[BankCardTypeExpirationDateInputIndex].Width = 5

	inputs[BankCardTypeSecurityCodeInputIndex] = textinput.New()
	inputs[BankCardTypeSecurityCodeInputIndex].Placeholder = "123"
	inputs[BankCardTypeSecurityCodeInputIndex].CharLimit = 3
	inputs[BankCardTypeSecurityCodeInputIndex].Width = 3

	return &BankCardTypeModel{
		inputs:  inputs,
		usecase: usecase,
	}
}

func (m BankCardTypeModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m BankCardTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()
	cmds := make([]tea.Cmd, len(m.inputs))

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == len(m.inputs)-1 {
				name := m.inputs[BankCardTypeNameInputIndex].Value()
				number := m.inputs[BankCardTypeNumberInputIndex].Value()
				expirationDate := m.inputs[BankCardTypeExpirationDateInputIndex].Value()
				securityCode := m.inputs[BankCardTypeSecurityCodeInputIndex].Value()

				err := m.usecase.CreateBankCard(ctx, name, number, expirationDate, securityCode)

				if err != nil {
					m.inputs[BankCardTypeNameInputIndex].Reset()
					m.inputs[BankCardTypeNameInputIndex].Focus()

					m.inputs[BankCardTypeNumberInputIndex].Reset()
					m.inputs[BankCardTypeNumberInputIndex].Blur()

					m.inputs[BankCardTypeExpirationDateInputIndex].Reset()
					m.inputs[BankCardTypeExpirationDateInputIndex].Blur()

					m.inputs[BankCardTypeSecurityCodeInputIndex].Reset()
					m.inputs[BankCardTypeSecurityCodeInputIndex].Blur()

					m.err = err
					m.prevInput()
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

func (m BankCardTypeModel) View() string {
	view := fmt.Sprintf(`Name: 
%s
Number:
%s
Expiration date:
%s
Security code:
%s
`,
		m.inputs[BankCardTypeNameInputIndex].View(),
		m.inputs[BankCardTypeNumberInputIndex].View(),
		m.inputs[BankCardTypeExpirationDateInputIndex].View(),
		m.inputs[BankCardTypeSecurityCodeInputIndex].View(),
	) + "\n"

	if m.err != nil {
		view += m.err.Error() + "\n"
	}

	view += "\nesc back\n"

	return view
}

// nextInput focuses the next input field
func (m *BankCardTypeModel) nextInput() {
	m.focused = (m.focused + 1) % len(m.inputs)
}

// prevInput focuses the previous input field
func (m *BankCardTypeModel) prevInput() {
	m.focused--
	// Wrap around
	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}
