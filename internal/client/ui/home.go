package ui

import (
	"context"
	"crypto/md5"
	"fmt"
	"os"

	"github.com/MisterMaks/go-yandex-gophkeeper/internal/client/domain"
	tea "github.com/charmbracelet/bubbletea"
)

type HomeModel struct {
	usecase        UsecaseInterface
	dataBatch      []domain.Data
	err            error
	cursor         int
	typeListModel  tea.Model
	isUpdatedModel bool
}

func NewHomeModel(usecase UsecaseInterface, typeListModel tea.Model) *HomeModel {
	return &HomeModel{
		usecase:       usecase,
		dataBatch:     []domain.Data{},
		typeListModel: typeListModel,
	}
}

func (m *HomeModel) Init() tea.Cmd {
	return nil
}

func (m *HomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
			if m.cursor < len(m.dataBatch)-1 {
				m.cursor++
			}

		case "n":
			m.isUpdatedModel = false
			return m.typeListModel, nil
		case "r":
			m.isUpdatedModel = false
		case "d":
			if len(m.dataBatch) == 0 {
				m.err = fmt.Errorf("No data for delete")
				m.isUpdatedModel = false
				return m, nil
			}

			data := m.dataBatch[m.cursor]
			id := data.ID
			err := m.usecase.DeleteData(context.Background(), id)
			if err != nil {
				m.err = err
				return m, nil
			}

			m.isUpdatedModel = false
			m.cursor = 0
		case "enter":
			data := m.dataBatch[m.cursor]
			if data.Type == domain.BinaryDataType {
				file, err := os.Create("./" + data.Name)
				if err != nil {
					m.err = err
					break
				}

				_, err = file.Write(data.Data)
				if err != nil {
					m.err = err
					return m, nil
				}
			}
		}
	}

	return m, nil
}

func (m *HomeModel) View() string {
	if !m.isUpdatedModel {
		m.GetDataBatch()
		m.isUpdatedModel = true
	}

	// The header
	s := "Data\n\n"

	// Iterate over our choices
	for i, data := range m.dataBatch {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		printData := string(data.Data)
		if data.Type == domain.BinaryDataType {
			hashSum := md5.Sum(data.Data)
			s += fmt.Sprintf("%s %s %s %s %x\n", cursor, data.ID, data.Name, data.Type, hashSum)
			continue
		}

		// Render the row
		s += fmt.Sprintf("%s %s %s %s %s\n", cursor, data.ID, data.Name, data.Type, printData)
	}

	if m.err != nil {
		s += m.err.Error() + "\n"
	}

	// The footer
	s += "\nr refresh | n new data | d delete | q quit\n"

	// Send the UI for rendering
	return s
}

func (m *HomeModel) GetDataBatch() {
	dataBatch, err := m.usecase.GetDataBatch(context.Background())
	if err != nil {
		m.err = err
	}

	m.dataBatch = dataBatch
}
