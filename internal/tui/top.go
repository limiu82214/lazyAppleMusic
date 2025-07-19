package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/util"
	"time"

	"limiu82214/lazyAppleMusic/internal/constant"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

type topTui struct {
	dump       io.Writer
	appleMusic bridge.PlayerBridge
	choices    []string         // items on the to-do list
	cursor     int              // which to-do list item our cursor is pointing at
	selected   map[int]struct{} // which to-do items are selected
	width      int
	height     int

	playingTui         PlayingTui
	currentPlaylistTui CurrentPlaylistTui
	helpTui            HelpTui
}

func InitialTopTui(dump io.Writer) topTui {
	appleMusic := bridge.NewAppleMusicBridge(dump)
	return topTui{
		dump:       dump,
		appleMusic: appleMusic,
		choices:    []string{"Playing"},
		selected:   make(map[int]struct{}),

		playingTui:         newPlayingTui(dump, appleMusic),
		currentPlaylistTui: newCurrentPlaylistTui(dump, appleMusic),
		helpTui:            newHelpTui(dump),
	}
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return constant.TickMsg(t)
	})
}

// ======= MAIN

func (m topTui) Init() tea.Cmd {
	m.fetchData()
	return tea.Batch(
		doTick(),
	)
}

// ======= VIEW

func (m topTui) View() string {
	leftHeight := m.height
	borderSize := lipgloss.ASCIIBorder().GetLeftSize() + lipgloss.ASCIIBorder().GetRightSize()
	width := m.width - borderSize

	// header
	header := m.playingTui.Width(width).View()
	leftHeight -= lipgloss.Height(header)

	// footer
	footer := m.helpTui.Width(width).View()
	leftHeight -= lipgloss.Height(footer)

	// content
	content := m.currentPlaylistTui.Width(width).Height(leftHeight).View()

	// leftHeight -= lipgloss.Height(content) + lipgloss.ASCIIBorder().GetTopSize() + lipgloss.ASCIIBorder().GetBottomSize()
	// spew.Fprintln(m.dump, "height:", m.height, "header:", lipgloss.Height(header), "content:", lipgloss.Height(content), "footer:", lipgloss.Height(footer))

	view := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)

	return view
}

func (m topTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case constant.EventUpdateTrackData:
		spew.Fprintln(m.dump, "Top EventUpdateTrackData:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingTui.Update(msg)
		m.playingTui, _ = pm.(PlayingTui)

		return m, cmd
	case constant.EventUpdateCurrentAlbumImg:
		// spew.Fprintln(m.dump, "Top EventUpdateCurrentAlbumImg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingTui.Update(msg)
		m.playingTui, _ = pm.(PlayingTui)

		return m, cmd
	case constant.EventUpdatePlayerPosition:
		spew.Fprintln(m.dump, "Top EventUpdatePlayerPosition:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingTui.Update(msg)
		m.playingTui, _ = pm.(PlayingTui)

		return m, cmd
	case constant.EventUpdateCurrentPlaylist:
		// spew.Fprintln(m.dump, "Top EventUpdateCurrentPlaylist:", util.JsonMarshalWhatever(msg))
		cpt, cmd := m.currentPlaylistTui.Update(msg)
		m.currentPlaylistTui, _ = cpt.(CurrentPlaylistTui)

		return m, cmd

	case timer.TimeoutMsg:
		spew.Fprintln(m.dump, "Top TimeoutMsg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingTui.Update(msg)
		m.playingTui, _ = pm.(PlayingTui)

		return m, cmd

	case timer.TickMsg:
		spew.Fprintln(m.dump, "Top TickMsg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingTui.Update(msg)
		m.playingTui, _ = pm.(PlayingTui)

		return m, cmd

	case constant.TickMsg:
		spew.Fprintln(m.dump, "Top constant.TickMsg:", util.JsonMarshalWhatever(msg))
		cmds := m.fetchData()

		cmds = append(cmds, doTick())
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		spew.Fprintln(m.dump, "Top WindowSizeMsg:", util.JsonMarshalWhatever(msg))
		m.width = msg.Width
		m.height = msg.Height

	case constant.EventTrackChanged:
		spew.Fprintln(m.dump, "Top EventTrackChanged:", util.JsonMarshalWhatever(msg))
		cmds := m.fetchData()
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
		spew.Fprintln(m.dump, "Top KeyMsg:", util.JsonMarshalWhatever(msg))
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "p":
			return m, m.appleMusic.PlayPause()
		case "n":
			return m, m.appleMusic.NextTrack()
		case "b":
			return m, m.appleMusic.PreviousTrack()
		case "s":
			return m, tea.Batch(m.appleMusic.Pause())
		case "u":
			return m, m.appleMusic.IncreaseVolume()
		case "d":
			return m, m.appleMusic.DecreaseVolume()
		case "f":
			return m, m.appleMusic.FavoriteTrack()
		case "r":
			cmds := m.fetchData()

			return m, tea.Batch(cmds...)
		case "k":
			cpt, cmd := m.currentPlaylistTui.Update(msg)
			m.currentPlaylistTui, _ = cpt.(CurrentPlaylistTui)
			return m, cmd
		case "j":
			cpt, cmd := m.currentPlaylistTui.Update(msg)
			m.currentPlaylistTui, _ = cpt.(CurrentPlaylistTui)
			return m, cmd
		}
	default:
		spew.Fprintln(m.dump, "Top unknown case:", util.JsonMarshalWhatever(msg))
	}

	return m, nil
}

// ====== fetch

func (m *topTui) fetchData() []tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, util.ToTeaCmd(m.fetchCurrentTrack))
	cmds = append(cmds, util.ToTeaCmd(m.fetchCurrentAlbumImg))
	cmds = append(cmds, util.ToTeaCmd(m.fetchPlayerPosition))
	cmds = append(cmds, util.ToTeaCmd(m.fetchCurrentPlaylist)) // TODO: consider goroutine because it is slow, make sure using mutex prevent concurrent access
	return cmds
}

func (m topTui) fetchCurrentTrack() constant.EventUpdateTrackData {
	track, err := m.appleMusic.GetCurrentTrack()
	if err != nil {
		spew.Fprintln(m.dump, "Error fetching current track:", err)
		track.Name = err.Error()
	}
	return constant.EventUpdateTrackData(track)
}

func (m topTui) fetchCurrentAlbumImg() constant.EventUpdateCurrentAlbumImg {
	currentAlbumImg, err := m.appleMusic.GetCurrentAlbum(int(float64(m.height)/2.5), int(float64(m.height)/2.5))
	if err != nil {
		currentAlbumImg = "Error fetching current album: " + err.Error()
	}
	return constant.EventUpdateCurrentAlbumImg(currentAlbumImg)
}

func (m topTui) fetchPlayerPosition() constant.EventUpdatePlayerPosition {
	playerPosition, err := m.appleMusic.GetPlayerPosition()
	if err != nil {
		spew.Fprintln(m.dump, "Error fetching player position:", err)
		playerPosition = 0
	}
	return constant.EventUpdatePlayerPosition(playerPosition)
}

func (m topTui) fetchCurrentPlaylist() constant.EventUpdateCurrentPlaylist {
	currentPlaylist, err := m.appleMusic.GetCurrentPlaylist()
	if err != nil {
		spew.Fdump(m.dump, "Error fetching current playlist:", err)
	}
	return constant.EventUpdateCurrentPlaylist(currentPlaylist)
}
