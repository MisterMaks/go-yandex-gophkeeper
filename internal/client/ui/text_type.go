package ui

import (
	"context"
	"fmt"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

const (
	TextTypeNameInputIndex = iota
	TextTypeTextInputIndex
)

type TextTypeModel struct {
	nameInput     textinput.Model
	textArea      textarea.Model
	focused       int
	err           error
	usecase       UsecaseInterface
	homeModel     tea.Model
	typeListModel tea.Model
}

func NewTextTypeModel(usecase UsecaseInterface) *TextTypeModel {
	nameInput := textinput.New()
	nameInput.Focus()

	textArea := textarea.New()

	return &TextTypeModel{
		nameInput: nameInput,
		textArea:  textArea,
		usecase:   usecase,
	}
}

func (m TextTypeModel) Init() tea.Cmd {
	return nil
}

func (m TextTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.focused == TextTypeTextInputIndex {
				name := m.nameInput.Value()
				text := m.textArea.Value()

				err := m.usecase.CreateText(ctx, name, text)

				if err != nil {
					m.nameInput.Reset()
					m.nameInput.Focus()

					m.textArea.Reset()
					m.textArea.Blur()

					m.err = err
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

		switch m.focused {
		case TextTypeNameInputIndex:
			m.nameInput.Focus()
			m.textArea.Blur()
		case TextTypeTextInputIndex:
			m.textArea.Focus()
			m.nameInput.Blur()
		}

	case error:
		m.err = msg
		return m, nil
	}

	var cmd1, cmd2 tea.Cmd
	m.nameInput, cmd1 = m.nameInput.Update(msg)
	m.textArea, cmd2 = m.textArea.Update(msg)

	return m, tea.Batch(cmd1, cmd2)
}

func (m TextTypeModel) View() string {
	view := fmt.Sprintf(`Name: 
%s
Text:
%s
`,
		m.nameInput.View(),
		m.textArea.View(),
	) + "\n"

	if m.err != nil {
		view += m.err.Error() + "\n"
	}

	view += "\nesc back\n"

	return view
}

// nextInput focuses the next input field
func (m *TextTypeModel) nextInput() {
	m.focused = (m.focused + 1) % 2
}

// prevInput focuses the previous input field
func (m *TextTypeModel) prevInput() {
	m.focused--

	if m.focused < 0 {
		m.focused = 2 - 1
	}
}
