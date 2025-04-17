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
	inputs        []*textinput.Model
	focused       int
	err           error
	usecase       UsecaseInterface
	homeModel     tea.Model
	typeListModel tea.Model

	nameInput           *textinput.Model
	numberInput         *textinput.Model
	expirationDateInput *textinput.Model
	securityCodeInput   *textinput.Model
}

func createTextInput(placeholder string, charLimit int, width int) *textinput.Model {
	inp := textinput.New()
	inp.Placeholder = placeholder
	inp.CharLimit = charLimit
	inp.Width = width
	return &inp
}

func NewBankCardTypeModel(usecase UsecaseInterface) *BankCardTypeModel {
	inputs := make([]*textinput.Model, 4)

	nameInput := textinput.New()
	nameInput.Focus()

	numberInput := createTextInput("1234********1234", 16, 16)
	expirationDateInput := createTextInput("10/11", 5, 5)
	securityCodeInput := createTextInput("123", 3, 3)

	inputs[BankCardTypeNameInputIndex] = &nameInput
	inputs[BankCardTypeNumberInputIndex] = numberInput
	inputs[BankCardTypeExpirationDateInputIndex] = expirationDateInput
	inputs[BankCardTypeSecurityCodeInputIndex] = securityCodeInput

	return &BankCardTypeModel{
		inputs:              inputs,
		usecase:             usecase,
		nameInput:           &nameInput,
		numberInput:         numberInput,
		expirationDateInput: expirationDateInput,
		securityCodeInput:   securityCodeInput,
	}
}

func (m BankCardTypeModel) Init() tea.Cmd {
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
				name := m.nameInput.Value()
				number := m.numberInput.Value()
				expirationDate := m.expirationDateInput.Value()
				securityCode := m.securityCodeInput.Value()

				err := m.usecase.CreateBankCard(ctx, name, number, expirationDate, securityCode)

				if err != nil {
					m.err = err
					m.resetInput()

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

	case error:
		m.err = msg
		return m, nil
	}

	var newModel textinput.Model
	for i := range m.inputs {
		newModel, cmds[i] = m.inputs[i].Update(msg)
		*m.inputs[i] = newModel
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
		m.nameInput.View(),
		m.numberInput.View(),
		m.expirationDateInput.View(),
		m.securityCodeInput.View(),
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

	if m.focused < 0 {
		m.focused = len(m.inputs) - 1
	}
}

func (m *BankCardTypeModel) resetInput() {
	for _, inp := range m.inputs {
		inp.Reset()
		inp.Blur()
	}

	m.nameInput.Focus()

	m.focused = 0
}
