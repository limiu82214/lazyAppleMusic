package model

type Track struct {
	Id          string
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

func (t Track) FilterValue() string {
	return t.Name + " " + t.Artist
}

func (t Track) Description() string { return t.Name }
