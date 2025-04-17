package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type TypeListModel struct {
	cursor     int
	choices    []string
	nextModels []tea.Model
	homeModel  tea.Model
}

func NewTypeListModel(
	loginPasswordTypeModel tea.Model,
	bankCardTypeModel tea.Model,
	textTypeModel tea.Model,
	binaryTypeModel tea.Model,
) *TypeListModel {
	return &TypeListModel{
		choices: []string{
			"Login & Password",
			"Bank card",
			"Text",
			"Binary",
		},
		nextModels: []tea.Model{
			loginPasswordTypeModel,
			bankCardTypeModel,
			textTypeModel,
			binaryTypeModel,
		},
	}
}

func (m TypeListModel) Init() tea.Cmd {
	return nil
}

func (m TypeListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		case "enter", " ":
			if m.cursor >= len(m.choices) {
				return m, nil
			}

			return m.nextModels[m.cursor].Update(nil)
		case "esc":
			return m.homeModel, nil
		}
	}

	return m, nil
}

func (m TypeListModel) View() string {
	s := "Choose data type:\n\n"

	for i, choice := range m.choices {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nesc back | q quit.\n"

	return s
}
