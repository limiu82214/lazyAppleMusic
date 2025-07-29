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

// Event for ation already been

type EventTrackChanged struct{}
type EventUpdateTrackData model.Track
type EventUpdateCurrentAlbumImg string
type EventUpdatePlayerPosition int
type EventUpdateCurrentPlaylist model.Playlist
type EventFavoriteTrackId string

// Should for need to be some action

type ShouldFavoriteTrackId string
type ShouldUpdateTabs model.TabTuiData
type ShouldPlayTrackId string
type ShouldSelectTrackId string
type ShouldClearFilter struct{}


const (
	Favorite   = "󰋑"
	Unfavorite = ""
)
