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

var _ tea.Model = (*currentPlaylistTui)(nil)

type currentPlaylistTui struct {
	dump            io.Writer
	appleMusic      bridge.PlayerBridge
	width           int
	height          int
	currentPlaylist model.Playlist
}

func newCurrentPlaylistTui(dump io.Writer, bridge bridge.PlayerBridge) currentPlaylistTui {
	return currentPlaylistTui{
		dump:       dump,
		appleMusic: bridge,
	}
}

func (m *currentPlaylistTui) fetch() {
	currentPlaylist, err := m.appleMusic.GetCurrentPlaylist()
	if err != nil {
		spew.Fdump(m.dump, "Error fetching current playlist:", err)
	}
	m.currentPlaylist = currentPlaylist
}

// ===== MAIN

func (m currentPlaylistTui) Init() tea.Cmd {
	return nil
}

// ======= UPDATE
func (m currentPlaylistTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fprintln(m.dump, "currentplaylist: ", msg)
	}

	// switch msg := msg.(type) {
	// case tea.WindowSizeMsg:
	// 	spew.Fdump(m.dump, msg.Width)

	// 	m.width = msg.Width
	// 	m.height = msg.Height
	// }

	return m, nil
}
func (m *currentPlaylistTui) SetSize(width, height int) {
	m.width = width
	m.height = height
}

// ======= VIEW
func (m currentPlaylistTui) View() string {
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
