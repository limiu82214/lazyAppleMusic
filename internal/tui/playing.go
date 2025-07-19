package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"
	"limiu82214/lazyAppleMusic/internal/util"
	"time"

	// "limiu82214/lazyAppleMusic/internal/bridge"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var playingDebug = true

type PlayingTui interface {
	tea.Model
	Width(width int) PlayingTui
	Height(height int) PlayingTui
}

type playingTui struct {
	dump              io.Writer
	appleMusic        bridge.PlayerBridge
	playingTrackTimer timer.Model

	style    lipgloss.Style
	track    model.Track
	albumImg string
}

func newPlayingTui(dump io.Writer, bridge bridge.PlayerBridge) PlayingTui {
	obj := &playingTui{
		dump:       dump,
		appleMusic: bridge,
		style: lipgloss.NewStyle().
			Align(lipgloss.Center).
			Border(lipgloss.RoundedBorder()),

		playingTrackTimer: timer.NewWithInterval(0, time.Second),
		track:             model.Track{},
		albumImg:          "󰎃",
	}
	if !playingDebug {
		obj.dump = io.Discard
	}
	return obj

}

// ======= MAIN

func (m *playingTui) Init() tea.Cmd {
	return nil
}

func (m *playingTui) View() string {
	viewStr := m.track.Name + " - " + m.track.Artist
	if m.track.Favorited {
		viewStr += " (" + constant.Favorite + ") "
	} else {
		viewStr += " (" + constant.Unfavorite + ") "
	}
	viewStr += " " + m.playingTrackTimer.Timeout.Abs().String() + " / "
	viewStr += m.track.Time
	playPercentage := (m.track.Duration - m.playingTrackTimer.Timeout.Seconds()) * 100 / m.track.Duration
	viewStr = m.style.Render(m.albumImg + "\n" + util.ProgressBarUi(int(playPercentage), int(float64(m.style.GetWidth())*0.8)) + "\n" + viewStr)

	return viewStr
}

func (m *playingTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		// spew.Fprintln(m.dump, "playing", msg)
	}

	switch msg := msg.(type) {
	case constant.EventUpdateTrackData:
		m.track = model.Track(msg)
	case constant.EventUpdateCurrentAlbumImg:
		m.albumImg = string(msg)
	case constant.EventUpdatePlayerPosition:
		pos := int(msg)
		m.playingTrackTimer = timer.NewWithInterval(time.Duration(int(m.track.Duration)-pos)*time.Second, time.Second)
		return m, m.playingTrackTimer.Init()
	case constant.EventFavoriteTrackId:
		if m.track.Id == string(msg) {
			if m.track.Favorited {
				m.track.Favorited = false
			} else {
				m.track.Favorited = true
			}
			return m, nil
		}

	case timer.TickMsg:
		switch msg.ID {
		case m.playingTrackTimer.ID():
			var cmd tea.Cmd
			m.playingTrackTimer, cmd = m.playingTrackTimer.Update(msg)
			return m, cmd
		}
	case timer.TimeoutMsg:
		switch msg.ID {
		case m.playingTrackTimer.ID():
			return m, nil
		}
	case constant.StyleMsg:
		m.style = msg.Style
	}

	return m, nil
}

// ======= Other
func (m *playingTui) Width(width int) PlayingTui {
	m.style = m.style.Width(width)
	return m
}
func (m *playingTui) Height(height int) PlayingTui {
	m.style = m.style.Height(height)
	return m
}
