package constant

import (
	"limiu82214/lazyAppleMusic/internal/model"
	"time"

	"github.com/charmbracelet/lipgloss"
)

type TickMsg time.Time

type StyleMsg struct {
	Style lipgloss.Style
}
type EventTrackChanged struct{}
type EventUpdateTrackData model.Track
type EventUpdateCurrentAlbumImg string
type EventUpdatePlayerPosition int
type EventUpdateCurrentPlaylist model.Playlist

const (
	Favorite   = "󰋑"
	Unfavorite = ""
)
