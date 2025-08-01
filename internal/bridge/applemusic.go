package bridge

import (
	"fmt"
	"io"
	"limiu82214/lazyAppleMusic/internal/constant"
	"limiu82214/lazyAppleMusic/internal/model"

	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/BigJk/imeji"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/davecgh/go-spew/spew"
)

type PlayerBridge interface {
	PlayPause() tea.Cmd
	Play() tea.Cmd
	Pause() tea.Cmd
	NextTrack() tea.Cmd
	PreviousTrack() tea.Cmd
	SetVolume(volume int) tea.Cmd
	IncreaseVolume() tea.Cmd
	DecreaseVolume() tea.Cmd
	PlayPlaylist(playlistName string) tea.Cmd
	PlayTrackById(id string) tea.Cmd
	FavoriteCurrentTrack() tea.Cmd
	FavoriteTrackByTrackId(id string) tea.Cmd

	GetPlayerPosition() (int, error)
	GetCurrentAlbum(width, height int) (string, error)
	GetCurrentTrack() (model.Track, error)
	GetPlaylists() ([]string, error)
	GetCurrentPlaylist() (model.Playlist, error)
}
type appleMusicBridge struct {
	appName string
	dump    io.Writer
}

func NewAppleMusicBridge(dump io.Writer) PlayerBridge {

	return &appleMusicBridge{
		appName: "Music",
		dump:    dump,
	}
}

func (a *appleMusicBridge) log(msg interface{}) {
	if a.dump != nil {

		spew.Fdump(a.dump, msg)
	}
}

func (a *appleMusicBridge) parseAppleRecord(input string) map[string]string {
	re := regexp.MustCompile(`(\w[\w ]*?):\s*([^,]*?)\s*(?:,|$)`)
	matches := re.FindAllStringSubmatch(input, -1)

	result := make(map[string]string)
	for _, m := range matches {
		key := strings.TrimSpace(m[1])
		val := strings.TrimSpace(m[2])
		result[key] = val
	}
	return result
}
func (a *appleMusicBridge) appleTrackRecordMap2Track(m map[string]string) model.Track {
	duration, _ := strconv.ParseFloat(m["duration"], 64)
	playedCount, _ := strconv.Atoi(m["played count"])

	track := model.Track{
		Id:          m["persistent ID"],
		Name:        m["name"],
		Time:        m["time"],
		Duration:    duration,
		PlayedCount: playedCount,
		Favorited:   m["favorited"] == "true",
		Album:       m["album"],
		AlbumArtist: m["album artist"],
		Artist:      m["artist"],
		Lyrics:      m["lyrics"],
	}
	return track
}

func (a *appleMusicBridge) GetCurrentTrack() (model.Track, error) {
	nullTrack := model.Track{Name: "No Track Playing"}
	script := fmt.Sprintf(`
		tell application "%s"
			set output to {}
			set end of output to properties of current track
		end tell
		return output
	`, a.appName)
	cmd := exec.Command("osascript", "-e", script)
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exit status 1" { // "Apple Music is not running"
			return nullTrack, nil
		}
		return nullTrack, fmt.Errorf("error getting current track: %v", err)
	}

	track := a.appleTrackRecordMap2Track(a.parseAppleRecord(string(output)))

	return track, nil
}

func (a *appleMusicBridge) PlayPause() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to playpause`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error toggling play/pause: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) Play() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to play`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error playing track: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) Pause() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to pause`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error pausing track: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) NextTrack() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to next track`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error skipping to next track: %v", err.Error()))
			return err
		}

		return constant.EventTrackChanged{}
	}
}

func (a *appleMusicBridge) PreviousTrack() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to previous track`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error skipping to previous track: %v", err.Error()))
			return err
		}

		return constant.EventTrackChanged{}
	}
}

func (a *appleMusicBridge) SetVolume(volume int) tea.Cmd {
	return func() tea.Msg {
		if volume < 0 || volume > 100 {
			return fmt.Errorf("volume must be between 0 and 100")
		}

		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to set sound volume to %d`, a.appName, volume))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error setting volume: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) IncreaseVolume() tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(`
		tell application "%s"
			set currentVolume to sound volume
			set sound volume to (currentVolume + 10)
		end tell
		`, a.appName)
		cmd := exec.Command("osascript", "-e", script)

		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error increasing volume: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) DecreaseVolume() tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(`
		tell application "%s"
			set currentVolume to sound volume
			set sound volume to (currentVolume - 10)
		end tell
		`, a.appName)
		cmd := exec.Command("osascript", "-e", script)

		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error decreasing volume: %v", err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) PlayPlaylist(playlistName string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to play playlist "%s"`, a.appName, playlistName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error playing playlist '%s': %v", playlistName, err.Error()))
			return err
		}
		return nil
	}
}

func (a *appleMusicBridge) PlayTrackById(id string) tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(`set targetID to "%s"
			set foundTrack to missing value
			tell application "%s"
				repeat with p in every playlist
					try
						set foundTrack to (first track of p whose persistent ID is targetID)
						if foundTrack is not missing value then
							exit repeat
						end if
					end try
				end repeat

				if foundTrack is not missing value then
					play foundTrack
					return "播放「" & (get name of foundTrack) & "」！"
				else
					return "錯誤：找不到 persistent ID 為 " & targetID & " 的歌曲。"
				end if
			end tell`, id, a.appName)

		cmd := exec.Command("osascript", "-e", script)

		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error play track byid: %v", err))
			return err
		}

		return constant.EventTrackChanged{}
	}
}

func (a *appleMusicBridge) FavoriteCurrentTrack() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s"
			set aTrack to current track
			set persistentId to persistent ID of aTrack
			if favorited of aTrack then
				set favorited of aTrack to false
			else
				set favorited of aTrack to true
			end if
		end tell
		return persistentId
		`, a.appName))
		output, err := cmd.Output()
		if err != nil {
			a.log(fmt.Sprintf("Error favoriting track: %v", err.Error()))
			return err
		}

		return constant.EventFavoriteTrackId(strings.TrimSpace(string(output)))
	}
}

func (a *appleMusicBridge) FavoriteTrackByTrackId(id string) tea.Cmd {
	return func() tea.Msg {
		script := fmt.Sprintf(`set targetID to "%s"
			set foundTrack to missing value
			tell application "%s"
				repeat with p in every playlist
					try
						set foundTrack to (first track of p whose persistent ID is targetID)
						if foundTrack is not missing value then
							exit repeat
						end if
					end try
				end repeat

				if foundTrack is not missing value then
					if favorited of foundTrack then
						set favorited of foundTrack to false
					else
						set favorited of foundTrack to true
					end if
					return "成功將歌曲「" & (get name of foundTrack) & "」加入收藏！"
				else
					return "錯誤：找不到 persistent ID 為 " & targetID & " 的歌曲。"
				end if
			end tell`, id, a.appName)

		cmd := exec.Command("osascript", "-e", script)

		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error favoriting track byid: %v", err))
			return err
		}

		return constant.EventFavoriteTrackId(id)
	}
}

func (a *appleMusicBridge) GetPlaylists() ([]string, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to set playlistNames to name of every playlist`, a.appName))
	output, err := cmd.Output()
	if err != nil {
		a.log(fmt.Sprintf("Error getting playlists: %v", err.Error()))
		return nil, err
	}

	// Convert output to string and split by comma
	playlists := strings.Split(string(output), ", ")
	playlistNames := []string{}
	for _, name := range playlists {
		if strings.TrimSpace(name) != "" {
			playlistNames = append(playlistNames, strings.TrimSpace(name))
		}
	}
	return playlistNames, nil
}

// FIXME: if is big list, it will be slow
func (a *appleMusicBridge) GetCurrentPlaylist() (model.Playlist, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`
	tell application "%s"
		set currentList to current playlist
		set output to {}
		repeat with t in tracks of currentList
			set end of output to properties of t
			set end of output to "######"
		end repeat
	end tell
	return output
	`, a.appName))

	// set end of output to (name of t & " - " & artist of t)
	output, err := cmd.Output()
	if err != nil {
		return model.Playlist{}, fmt.Errorf("error getting current playlist: %v", err)
	}

	trackStrList := strings.Split(string(output), "######")
	if len(trackStrList) > 0 {
		trackStrList = trackStrList[:len(trackStrList)-1] // Remove the last empty string if it exists
	}
	playlist := model.Playlist{}
	for _, trackStr := range trackStrList {
		playlist.Tracks = append(playlist.Tracks,
			a.appleTrackRecordMap2Track(a.parseAppleRecord(trackStr)),
		)
	}

	return playlist, nil
}

func (a *appleMusicBridge) GetPlayerPosition() (int, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`
		tell application "%s"
			set playerPosition to player position
		end tell
		return playerPosition
	`, a.appName))

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("error getting player position: %v", err)
	}

	position, err := strconv.ParseFloat(strings.TrimSpace(string(output)), 64)
	if err != nil {
		return 0, fmt.Errorf("error parsing player position: %v", err)
	}

	return int(position), nil
}

// TODO: cache album img
// TODO: check img exist
func (a *appleMusicBridge) GetCurrentAlbum(width, height int) (string, error) {
	width *= 2
	filePath := "/tmp/cover.jpg"
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`
		set tmpPath to POSIX file "%s"
		tell application "%s"
			set aTrack to current track
			set ac to count of artworks of aTrack
			if ac = 0 then return "No Artwork"
			set artData to data of artwork 1 of aTrack
		end tell
		set outFile to open for access tmpPath with write permission
		try
			set eof outFile to 0
			write artData to outFile
		end try
		close access outFile
		return POSIX path of tmpPath
		EOF
	`, filePath, a.appName))

	err := cmd.Run()
	if err != nil {
		spew.Fdump(a.dump, "Error getting current album artwork:", err)
		return "", fmt.Errorf("error getting current album artwork: %v", err)
	}

	sizeOpt := imeji.WithResize(width, height)
	text, err := imeji.FileString(
		filePath,
		sizeOpt,
		imeji.WithTrueColor(), // 24-bit
	)
	if err != nil {
		return "", fmt.Errorf("imeji: %w", err)
	}
	return text, nil
}
