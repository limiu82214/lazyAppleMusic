package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"

	// "limiu82214/lazyAppleMusic/internal/bridge"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

var currentPlaylistDebug = false

type CurrentPlaylistTui interface {
	tea.Model
	Width(width int) CurrentPlaylistTui
	Height(height int) CurrentPlaylistTui
}

type currentPlaylistTui struct {
	dump       io.Writer
	appleMusic bridge.PlayerBridge
	vp         viewport.Model

	currentPlaylist model.Playlist
}

func newCurrentPlaylistTui(dump io.Writer, bridge bridge.PlayerBridge) CurrentPlaylistTui {
	obj := &currentPlaylistTui{
		dump:       dump,
		appleMusic: bridge,
		vp:         viewport.New(0, 0),

		currentPlaylist: model.Playlist{},
	}
	obj.vp.Style = lipgloss.NewStyle().
		Align(lipgloss.Center).
		Border(lipgloss.RoundedBorder())
	if !currentPlaylistDebug {
		obj.dump = io.Discard
	}
	return obj
}

// ======= MAIN

func (m *currentPlaylistTui) Init() tea.Cmd {
	return nil
}

func (m *currentPlaylistTui) View() string {
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

	m.vp.SetContent(viewStr)
	return m.vp.View()
}

func (m *currentPlaylistTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fprintln(m.dump, "currentplaylist: ", msg)
	}

	switch msg := msg.(type) {
	case constant.EventUpdateCurrentPlaylist:
		m.currentPlaylist = model.Playlist(msg)
	case tea.KeyMsg:
		switch msg.String() {
		case "k":
			m.vp.ScrollUp(1)
		case "j":
			m.vp.ScrollDown(1)
		}
	}

	return m, nil
}

// ======= Other

func (m *currentPlaylistTui) Width(width int) CurrentPlaylistTui {
	m.vp.Width = width
	return m
}
func (m *currentPlaylistTui) Height(height int) CurrentPlaylistTui {
	m.vp.Height = height
	return m
}
