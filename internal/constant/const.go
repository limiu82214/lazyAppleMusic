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
type EventFavoriteTrackId string
type ShouldFavoriteTrackId string

const (
	Favorite   = "󰋑"
	Unfavorite = ""
)
