package tui

import (
	"io"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpDebug = false

type HelpTui interface {
	tea.Model
	Width(width int) HelpTui
	Height(height int) HelpTui
}
type helpTui struct {
	dump io.Writer

	style   lipgloss.Style
	content string
}

func newHelpTui(dump io.Writer) HelpTui {
	obj := &helpTui{
		dump:  dump,
		style: lipgloss.NewStyle().Align(lipgloss.Center),
		content: "p: play/pause, " +
			"n: next, " +
			"b: previous, " +
			"u: volume up, " +
			"d: volume down, " +
			"h: list prev page," +
			"l: list next page, " +
			"j: list cursor down, " +
			"k: list cursor up, " +
			"s: select current track, " +
			"f: favorite selected track, " +
			"F: favorite current track, " +
			"g: play selected track, " +
			"<: prev list, " +
			">: next list, " +
			"r: refresh, " +
			"q: quit",
	}
	if !helpDebug {
		obj.dump = io.Discard
	}
	return obj
}

// ======= MAIN

func (m *helpTui) Init() tea.Cmd {
	return nil
}

func (m *helpTui) View() string {
	return m.style.Render(m.content)
}

func (m *helpTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

// ======= Other

func (m *helpTui) Width(width int) HelpTui {
	m.style = m.style.Width(width)
	return m
}

func (m *helpTui) Height(height int) HelpTui {
	m.style = m.style.Height(height)
	return m
}
