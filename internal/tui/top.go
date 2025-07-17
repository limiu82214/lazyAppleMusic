package tui

import (
	"encoding/json"
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/model"
	"time"

	"limiu82214/lazyAppleMusic/internal/constant"

	"github.com/charmbracelet/bubbles/timer"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

// 新增錯誤和計時訊息的類型定義
type errorMsg struct{ err error }

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

	playingTrackTimer timer.Model
	data              topData
}

// ======= INITIAL
func InitialTopModel(dump io.Writer) topModel {
	appleMusic := bridge.NewAppleMusicBridge(dump)
	return topModel{
		dump:              dump,
		appleMusic:        appleMusic,
		choices:           []string{"Playing"},
		selected:          make(map[int]struct{}),
		currentPlaylist:   newCurrentPlaylistModel(dump, appleMusic),
		vpOfContent:       viewport.New(0, 0),
		playingTrackTimer: timer.NewWithInterval(0, time.Second),
		data:              topData{},
	}
}

func doTick() tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return constant.TickMsg(t)
	})
}

func (m topModel) Init() tea.Cmd {
	m.fetchData()
	m.reSize()
	return tea.Batch(
		doTick(),
	)
}

// ======= UPDATE
func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		b, _ := json.Marshal(msg)
		spew.Fdump(m.dump, "top"+string(b))
	}

	// var cmds []tea.Cmd

	switch msg := msg.(type) {

	case timer.TimeoutMsg:
		spew.Fdump(m.dump, "TimeoutMsg:", msg)
		switch msg.ID {
		case m.playingTrackTimer.ID():
			return m, nil
		}
	case timer.TickMsg:
		spew.Fdump(m.dump, "TickMsg:", msg)
		switch msg.ID {
		case m.playingTrackTimer.ID():
			var cmd tea.Cmd
			m.playingTrackTimer, cmd = m.playingTrackTimer.Update(msg)
			return m, cmd
		}
	case constant.TickMsg:
		cmds := m.fetchData()
		m.reSize()
		cmds = append(cmds, doTick())
		return m, tea.Batch(cmds...)

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.reSize()

	case constant.EventTrackChanged:
		cmds := m.fetchData()
		return m, tea.Batch(cmds...)

	case tea.KeyMsg:
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
			return m, tea.Batch(m.playingTrackTimer.Stop(), m.appleMusic.Pause())
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
	}

	return m, nil
}

// ======= VIEW

func (m topModel) View() string {
	return m.reSize()
}

func (m *topModel) fetchData() []tea.Cmd {
	cmds := []tea.Cmd{}
	// track
	oldTrckId := m.data.CurrentTrack.Id
	track, err := m.appleMusic.GetCurrentTrack()
	if err != nil {
		track.Name = err.Error()
	}
	m.data.CurrentTrack = track

	// other
	if oldTrckId != m.data.CurrentTrack.Id {
		// playingTrackTimer
		m.data.PlayerPosition, err = m.appleMusic.GetPlayerPosition()
		if err != nil {
			spew.Fdump(m.dump, "Error fetching player position:", err)
			m.data.PlayerPosition = 0
		}
		m.playingTrackTimer = timer.NewWithInterval(time.Duration(int(m.data.CurrentTrack.Duration)-m.data.PlayerPosition)*time.Second, time.Second)
		cmds = append(cmds, m.playingTrackTimer.Init())

		// currentAlbum
		currentAlbum, err := m.appleMusic.GetCurrentAlbum(int(float64(m.height)/2.5), int(float64(m.height)/2.5))
		if err != nil {
			currentAlbum = "Error fetching current album: " + err.Error()
		}
		m.data.CurrentAlbum = currentAlbum

		// currentPlaylist
		m.currentPlaylist.fetch() // TODO: consider goroutine because it is slow, make sure using mutex prevent concurrent access
		m.vpOfContent.SetContent(m.currentPlaylist.View())
	}

	return cmds
}
func (m *topModel) reSize() string {
	// 整理資料
	trackString := m.data.CurrentTrack.Name + " - " + m.data.CurrentTrack.Artist
	if m.data.CurrentTrack.Favorited {
		trackString += " (" + constant.Favorite + ") "
	} else {
		trackString += " (" + constant.Unfavorite + ") "
	}
	trackString += " " + m.playingTrackTimer.Timeout.Abs().String() + " / "
	trackString += m.data.CurrentTrack.Time

	leftHeight := m.height
	borderSize := lipgloss.ASCIIBorder().GetLeftSize() + lipgloss.ASCIIBorder().GetRightSize()
	width := m.width - borderSize

	// header
	header := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width).
		Border(lipgloss.RoundedBorder()).
		Render(m.data.CurrentAlbum + "\n" + trackString)
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
