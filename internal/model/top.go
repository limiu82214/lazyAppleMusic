package model

import (
	"encoding/json"
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"time"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/davecgh/go-spew/spew"
)

// 新增錯誤和計時訊息的類型定義
type errorMsg struct{ err error }
type tickMsg time.Time

type topModel struct {
	dump            io.Writer
	appleMusic      bridge.PlayerBridge
	choices         []string         // items on the to-do list
	cursor          int              // which to-do list item our cursor is pointing at
	selected        map[int]struct{} // which to-do items are selected
	width           int
	height          int
	currentPlaylist currentPlaylistModel
}

// ======= INITIAL
func InitialTopModel(dump io.Writer) topModel {
	appleMusic := bridge.NewAppleMusicBridge(dump)
	return topModel{
		dump:            dump,
		appleMusic:      appleMusic,
		choices:         []string{"Playing"},
		selected:        make(map[int]struct{}),
		currentPlaylist: newCurrentPlaylistModel(dump, appleMusic),
	}
}

func (m topModel) Init() tea.Cmd {
	// TODO: auto refresh cover and playing info
	return nil
}

// ======= UPDATE
func (m topModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		b, _ := json.Marshal(msg)
		spew.Fdump(m.dump, "top"+string(b))
	}

	// var cmds []tea.Cmd

	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

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
			return m, m.appleMusic.Pause()
		case "u":
			return m, m.appleMusic.IncreaseVolume()
		case "d":
			return m, m.appleMusic.DecreaseVolume()

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
	trackName, err := m.appleMusic.GetCurrentTrack()
	if err != nil {
		trackName = err.Error()
	}
	// TODO: don't generate cover every time
	currentAlbum, err := m.appleMusic.GetCurrentAlbum(int(float64(m.height)/2.5), int(float64(m.height)/2.5))
	if err != nil {
		currentAlbum = "Error fetching current album: " + err.Error()
	}

	leftHeight := m.height
	borderSize := lipgloss.ASCIIBorder().GetLeftSize() + lipgloss.ASCIIBorder().GetRightSize()
	width := m.width - borderSize
	header := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width).
		Border(lipgloss.RoundedBorder()).
		Render(currentAlbum + "\nPlaying: " + trackName)
	leftHeight -= lipgloss.Height(header)

	footer := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Width(width).
		Render("p: play/pause, s: pause, n: next, b: previous, u: volume up, d: volume down, q: quit")
	leftHeight -= lipgloss.Height(footer)

	vp := viewport.New(width, leftHeight-lipgloss.ASCIIBorder().GetTopSize()-lipgloss.ASCIIBorder().GetBottomSize())
	vp.SetContent(m.currentPlaylist.View())
	content := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		Width(width).
		Render(vp.View())

	leftHeight -= lipgloss.Height(content) + lipgloss.ASCIIBorder().GetTopSize() + lipgloss.ASCIIBorder().GetBottomSize()
	// spew.Fprintln(m.dump, "height:", m.height, "header:", lipgloss.Height(header), "content:", lipgloss.Height(content), "footer:", lipgloss.Height(footer))

	view := lipgloss.JoinVertical(
		lipgloss.Top,
		header,
		content,
		footer,
	)

	return view

}
