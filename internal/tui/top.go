package tui

import (
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/model"
	"limiu82214/lazyAppleMusic/internal/util"
	"time"

	"limiu82214/lazyAppleMusic/internal/constant"

	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

type topData struct {
	CurrentTrack   model.Track
	PlayerPosition int
	CurrentAlbum   string
}
type topModel struct {
	dump            io.Writer
	appleMusic      bridge.PlayerBridge
	choices         []string         // items on the to-do list
	cursor          int              // which to-do list item our cursor is pointing at
	selected        map[int]struct{} // which to-do items are selected
	width           int
	height          int
	currentPlaylist currentPlaylistModel
	vpOfContent     viewport.Model

	playingModel PlayingModel
	data         topData
}

func InitialTopModel(dump io.Writer) topModel {
	appleMusic := bridge.NewAppleMusicBridge(dump)
	return topModel{
		dump:            dump,
		appleMusic:      appleMusic,
		choices:         []string{"Playing"},
		selected:        make(map[int]struct{}),
		currentPlaylist: newCurrentPlaylistModel(dump, appleMusic),
		vpOfContent:     viewport.New(0, 0),
		playingModel:    newPlayingModel(dump, appleMusic),
		data:            topData{},
	}
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return constant.TickMsg(t)
	})
}

// ======= MAIN

func (m topModel) Init() tea.Cmd {
	m.fetchData()
	m.reSize()
	return tea.Batch(
		doTick(),
	)
}

// ======= VIEW

func (m topModel) View() string {
	return m.reSize()
}

func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case constant.EventUpdateTrackData:
		spew.Fprintln(m.dump, "Top EventUpdateTrackData:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingModel.Update(msg)
		m.playingModel, _ = pm.(PlayingModel)

		m.reSize()
		return m, cmd
	case constant.EventUpdateCurrentAlbumImg:
		// spew.Fprintln(m.dump, "Top EventUpdateCurrentAlbumImg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingModel.Update(msg)
		m.playingModel, _ = pm.(PlayingModel)

		m.reSize()
		return m, cmd
	case constant.EventUpdatePlayerPosition:
		spew.Fprintln(m.dump, "Top EventUpdatePlayerPosition:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingModel.Update(msg)
		m.playingModel, _ = pm.(PlayingModel)
		m.reSize()
		return m, cmd

	case timer.TimeoutMsg:
		spew.Fprintln(m.dump, "Top TimeoutMsg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingModel.Update(msg)
		m.playingModel, _ = pm.(PlayingModel)
		m.reSize()
		return m, cmd

	case timer.TickMsg:
		spew.Fprintln(m.dump, "Top TickMsg:", util.JsonMarshalWhatever(msg))
		pm, cmd := m.playingModel.Update(msg)
		m.playingModel, _ = pm.(PlayingModel)
		m.reSize()
		return m, cmd

	case constant.TickMsg:
		spew.Fprintln(m.dump, "Top constant.TickMsg:", util.JsonMarshalWhatever(msg))
		cmds := m.fetchData()
		m.reSize()
		cmds = append(cmds, doTick())
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		spew.Fprintln(m.dump, "Top WindowSizeMsg:", util.JsonMarshalWhatever(msg))
		m.width = msg.Width
		m.height = msg.Height
		m.reSize()

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
			m.reSize()
			return m, tea.Batch(cmds...)
		case "k":
			m.vpOfContent.ScrollUp(1)
			return m, nil
		case "j":
			m.vpOfContent.ScrollDown(1)
			return m, nil

			// // old
			// case "up", "k":
			// 	if m.cursor > 0 {
			// 		m.cursor--
			// 	}
			// case "down", "j":
			// 	if m.cursor < len(m.choices)-1 {
			// 		m.cursor++
			// 	}
			// case "enter", " ":
			// 	_, ok := m.selected[m.cursor]
			// 	if ok {
			// 		delete(m.selected, m.cursor)
			// 	} else {
			// 		m.selected[m.cursor] = struct{}{}
			// 	}
		}
	default:
		spew.Fprintln(m.dump, "Top unknown case:", util.JsonMarshalWhatever(msg))
	}

	return m, nil
}

func (m *topModel) reSize() string {
	// 整理資料

	leftHeight := m.height
	borderSize := lipgloss.ASCIIBorder().GetLeftSize() + lipgloss.ASCIIBorder().GetRightSize()
	width := m.width - borderSize

	// header
	header := m.playingModel.Width(width).View()
	leftHeight -= lipgloss.Height(header)

	// footer
	footer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width).
		Render("p: play/pause, s: pause, n: next, b: previous, u: volume up, d: volume down, f: favorite, r: refresh, q: quit")
	leftHeight -= lipgloss.Height(footer)

	// content
	m.vpOfContent.Width = width
	m.vpOfContent.Height = leftHeight
	m.vpOfContent.Style = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(m.width)
	content := m.vpOfContent.View()

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

// ====== fetch

func (m *topModel) fetchData() []tea.Cmd {
	cmds := []tea.Cmd{}
	cmds = append(cmds, util.ToTeaCmd(m.fetchCurrentTrack))
	cmds = append(cmds, util.ToTeaCmd(m.fetchCurrentAlbumImg))
	cmds = append(cmds, util.ToTeaCmd(m.fetchPlayerPosition))

	// currentPlaylist
	m.currentPlaylist.fetch() // TODO: consider goroutine because it is slow, make sure using mutex prevent concurrent access
	m.vpOfContent.SetContent(m.currentPlaylist.View())

	return cmds
}

func (m topModel) fetchCurrentTrack() constant.EventUpdateTrackData {
	track, err := m.appleMusic.GetCurrentTrack()
	if err != nil {
		spew.Fprintln(m.dump, "Error fetching current track:", err)
		track.Name = err.Error()
	}
	return constant.EventUpdateTrackData(track)
}

func (m topModel) fetchCurrentAlbumImg() constant.EventUpdateCurrentAlbumImg {
	currentAlbumImg, err := m.appleMusic.GetCurrentAlbum(int(float64(m.height)/2.5), int(float64(m.height)/2.5))
	if err != nil {
		currentAlbumImg = "Error fetching current album: " + err.Error()
	}
	return constant.EventUpdateCurrentAlbumImg(currentAlbumImg)
}

func (m topModel) fetchPlayerPosition() constant.EventUpdatePlayerPosition {
	playerPosition, err := m.appleMusic.GetPlayerPosition()
	if err != nil {
		spew.Fprintln(m.dump, "Error fetching player position:", err)
		playerPosition = 0
	}
	return constant.EventUpdatePlayerPosition(playerPosition)
}
