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
	return nil
}

func (m LoginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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

			if m.choices[m.cursor] == RegisterKey {
				m.loginFormModel.SetRegister(true)
			}

			return m.loginFormModel, nil
		}
	}

	return m, nil
}

func (m LoginModel) View() string {
	s := "Login or Register\n\n"

	for i, choice := range m.choices {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	s += "\nPress q to quit.\n"

	return s
}
