package tui

import (
	"fmt"
	"io"
	"limiu82214/lazyAppleMusic/internal/bridge"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"
	"limiu82214/lazyAppleMusic/internal/util"
	"strings"

	// "limiu82214/lazyAppleMusic/internal/bridge"
	"github.com/charmbracelet/bubbles/list"
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

	style lipgloss.Style
	list  list.Model
}

func newCurrentPlaylistTui(dump io.Writer, bridge bridge.PlayerBridge) CurrentPlaylistTui {
	list := list.New([]list.Item{
		model.Track{Name: "Loading...", Artist: "Loading..."},
	}, currentPlayListDelegate{}, 0, 0)
	list.SetShowTitle(false)
	list.SetShowTitle(false)
	list.SetShowHelp(false)
	list.SetShowStatusBar(false)
	list.SetShowPagination(true)

	obj := &currentPlaylistTui{
		dump:       dump,
		appleMusic: bridge,
		style: lipgloss.NewStyle().
			Align(lipgloss.Left).
			Border(lipgloss.RoundedBorder()),

		list: list,
	}

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
	return m.style.Render(m.list.View())
}

func (m *currentPlaylistTui) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if m.dump != nil {
		spew.Fprintln(m.dump, "currentplaylist: ", msg)
	}

	switch msg := msg.(type) {
	case constant.EventUpdateCurrentPlaylist:
		currentPlaylist := model.Playlist(msg)
		items := make([]list.Item, len(currentPlaylist.Tracks))
		for i := range currentPlaylist.Tracks {
			items[i] = currentPlaylist.Tracks[i]
		}
		m.list.SetItems(items)
	case tea.KeyMsg:
		switch msg.String() {
		case "k":
			m.list.CursorUp()
		case "j":
			m.list.CursorDown()
		case "h":
			m.list.PrevPage()
		case "l":
			m.list.NextPage()
		case "f":
			track := m.list.SelectedItem().(model.Track)
			return m, util.ToTeaCmdMsg(constant.ShouldFavoriteTrackId(track.Id))
		}
	case constant.EventFavoriteTrackId:
		items := m.list.Items()
		index := -1
		for i := range items {
			item := items[i].(model.Track)
			if item.Id == string(msg) {
				index = i
				break
			}
		}
		if index > 0 {
			tmp := m.list.Items()[index].(model.Track)
			tmp.Favorited = !tmp.Favorited
			m.list.SetItem(index, tmp)
		}
	}

	return m, nil
}

// ======= Other

func (m *currentPlaylistTui) Width(width int) CurrentPlaylistTui {
	m.list.SetWidth(width)
	m.style = m.style.Width(width)
	return m
}
func (m *currentPlaylistTui) Height(height int) CurrentPlaylistTui {
	m.list.SetHeight(height)
	m.style = m.style.Height(height)
	return m
}

type currentPlayListDelegate struct{}

func (d currentPlayListDelegate) Height() int                               { return 1 }
func (d currentPlayListDelegate) Spacing() int                              { return 0 }
func (d currentPlayListDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }
func (d currentPlayListDelegate) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	i, ok := listItem.(model.Track)
	if !ok {
		return
	}

	row := ""
	if i.Favorited {
		row = constant.Favorite + " " + i.Name + " - " + i.Artist
	} else {
		row = constant.Unfavorite + " " + i.Name + " - " + i.Artist
	}

	fn := lipgloss.NewStyle().PaddingLeft(4).Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return lipgloss.NewStyle().PaddingLeft(2).Bold(true).Foreground(lipgloss.Color("205")).Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(row))
}
