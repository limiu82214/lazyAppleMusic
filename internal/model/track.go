package model

type Track struct {
	Name        string
	Time        string
	Duration    float64
	PlayedCount int
	Favorited   bool
	Artist      string
	Album       string
	AlbumArtist string
	Lyrics      string
}
