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
	case tea.KeyMsg:
		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

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
			switch {
			case data.IsChunked:
				err := m.usecase.GetChunkedData(context.Background(), data.ID, "./"+data.Name)
				if err != nil {
					m.err = err
					break
				}
			case data.Type == domain.BinaryDataType:
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

	s := "Data\n\n"

	for i, data := range m.dataBatch {

		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}

		printData := string(data.Data)
		if data.Type == domain.BinaryDataType {
			hashSum := md5.Sum(data.Data)
			s += fmt.Sprintf("%s %s %s %s %x\n", cursor, data.ID, data.Name, data.Type, hashSum)
			continue
		}

		s += fmt.Sprintf("%s %s %s %s %s\n", cursor, data.ID, data.Name, data.Type, printData)
	}

	if m.err != nil {
		s += m.err.Error() + "\n"
	}

	s += "\nr refresh | n new data | d delete | q quit\n"

	return s
}

func (m *HomeModel) GetDataBatch() {
	dataBatch, err := m.usecase.GetDataBatch(context.Background())
	if err != nil {
		m.err = err
	}

	m.dataBatch = dataBatch
}
