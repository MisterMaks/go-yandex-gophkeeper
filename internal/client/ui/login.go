package ui

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	LoginIndex = iota
	RegisterIndex
	LoginKey    = "Login"
	RegisterKey = "Register"
)

type LoginModel struct {
	choices        []string
	cursor         int
	loginFormModel *LoginFormModel
}

func NewLoginModel(loginFormModel *LoginFormModel) *LoginModel {
	return &LoginModel{
		choices:        []string{LoginKey, RegisterKey},
		loginFormModel: loginFormModel,
	}
}

func (m LoginModel) Init() tea.Cmd {
	// Just return `nil`, which means "no I/O right now, please."
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			if m.choices[m.cursor] == RegisterKey {
				m.loginFormModel.SetRegister(true)
			}

			return m.loginFormModel, nil
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m LoginModel) View() string {
	// The header
	s := "Login or Register\n\n"

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
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}
