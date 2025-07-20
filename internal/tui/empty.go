package tui

import (
	tea "github.com/charmbracelet/bubbletea"
)

type emptyModel struct {
}

func (m emptyModel) Init() tea.Cmd {
	return nil
}

func (m emptyModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m emptyModel) View() string {
	return "This is an empty model."
}
