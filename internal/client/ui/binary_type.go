package ui

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	tea "github.com/charmbracelet/bubbletea"
)

type BinaryTypeModel struct {
	filepicker    filepicker.Model
	quitting      bool
	err           error
	usecase       UsecaseInterface
	homeModel     tea.Model
	typeListModel tea.Model
	isInitialized bool
}

func NewBinaryTypeModel(usecase UsecaseInterface) *BinaryTypeModel {
	fp := filepicker.New()
	fp.Height = 33
	fp.CurrentDirectory, _ = os.UserHomeDir()

	return &BinaryTypeModel{
		filepicker: fp,
		usecase:    usecase,
	}
}

func (m *BinaryTypeModel) Init() tea.Cmd {
	return m.filepicker.Init()
}

func (m *BinaryTypeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	ctx := context.Background()

	if !m.isInitialized {
		cmd := m.Init()
		m.isInitialized = true
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		case "esc":
			m.isInitialized = false
			return m.typeListModel, nil
		}
	case clearErrorMsg:
		m.err = nil
	}

	var cmd tea.Cmd
	m.filepicker, cmd = m.filepicker.Update(msg)

	if didSelect, path := m.filepicker.DidSelectFile(msg); didSelect {
		err := m.usecase.CreateBinary(ctx, path)
		if err != nil {
			m.err = err
			return m, tea.Batch(cmd, clearErrorAfter(2*time.Second))
		}

		return m.homeModel, nil
	}

	return m, cmd
}

func (m *BinaryTypeModel) View() string {
	if m.quitting {
		return ""
	}

	var s strings.Builder

	s.WriteString(m.filepicker.CurrentDirectory)

	s.WriteString("\n  ")

	if m.err != nil {
		s.WriteString(m.filepicker.Styles.DisabledFile.Render(m.err.Error()))
	}

	s.WriteString("\n\n" + m.filepicker.View() + "\n")

	return s.String()
}

type clearErrorMsg struct{}

func clearErrorAfter(t time.Duration) tea.Cmd {
	return tea.Tick(t, func(_ time.Time) tea.Msg {
		return clearErrorMsg{}
	})
}
