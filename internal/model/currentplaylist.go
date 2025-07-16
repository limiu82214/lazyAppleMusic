package model

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"

	// "limiu82214/lazyAppleMusic/internal/bridge"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

var _ tea.Model = (*currentPlaylistModel)(nil)

type currentPlaylistModel struct {
	dump       io.Writer
	appleMusic bridge.PlayerBridge
	width      int
	height     int
}

func newCurrentPlaylistModel(dump io.Writer, bridge bridge.PlayerBridge) currentPlaylistModel {
	return currentPlaylistModel{
		dump:       dump,
		appleMusic: bridge,
	}
}

// ===== MAIN

func (m currentPlaylistModel) Init() tea.Cmd {
	return nil
}

// ======= UPDATE
func (m currentPlaylistModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fdump(m.dump, msg)
	}

	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	spew.Fdump(m.dump, msg.Width)

	// 	m.width = msg.Width
	// 	m.height = msg.Height
	// }

	return m, nil
}
func (m *currentPlaylistModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ======= VIEW
func (m currentPlaylistModel) View() string {
	currentPlaylist, err := m.appleMusic.GetCurrentPlaylist()
	if err != nil {
		currentPlaylist = []string{"Error fetching current playlist: " + err.Error()}
	}

	view := lipgloss.NewStyle().
		MaxHeight(m.height).
		MaxWidth(m.width).
		Render("Current Playlist:\n" + lipgloss.JoinVertical(lipgloss.Top, currentPlaylist...))

	return view
}
