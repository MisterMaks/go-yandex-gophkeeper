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
	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":
			if m.cursor >= len(m.choices) {
				return m, nil
			}

			return m.nextModels[m.cursor].Update(nil)
		case "esc":
			return m.homeModel, nil
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m TypeListModel) View() string {
	// The header
	s := "Choose data type:\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	// The footer
	s += "\nesc back | q quit.\n"

	// Send the UI for rendering
	return s
}
