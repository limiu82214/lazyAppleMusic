package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"

	// "limiu82214/lazyAppleMusic/internal/bridge"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

var _ tea.Model = (*currentPlaylistModel)(nil)

type currentPlaylistModel struct {
	dump            io.Writer
	appleMusic      bridge.PlayerBridge
	width           int
	height          int
	currentPlaylist model.Playlist
}

func newCurrentPlaylistModel(dump io.Writer, bridge bridge.PlayerBridge) currentPlaylistModel {
	return currentPlaylistModel{
		dump:       dump,
		appleMusic: bridge,
	}
}

func (m *currentPlaylistModel) fetch() {
	currentPlaylist, err := m.appleMusic.GetCurrentPlaylist()
	if err != nil {
		spew.Fdump(m.dump, "Error fetching current playlist:", err)
	}
	m.currentPlaylist = currentPlaylist
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
	viewStr := ""
	for _, track := range m.currentPlaylist.Tracks {
		if track.Favorited {
			viewStr += "(" + constant.Favorite + ") "
		} else {
			viewStr += "(" + constant.Unfavorite + ") "
		}
		viewStr += track.Name + " - " + track.Artist
		viewStr += "\n"
	}

	return viewStr
}
