package model

import tea "github.com/charmbracelet/bubbletea"

type TabTuiData struct {
	Tabs       []string
	TabContent []tea.Model
	ActiveTab  int
}
