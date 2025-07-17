package bridge

import (
	"fmt"
	"io"
	"os/exec"
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
	PlayTrackByName(trackName string) tea.Cmd

	GetCurrentAlbum(width, height int) (string, error)
	GetCurrentTrack() (string, error)
	GetPlaylists() ([]string, error)
	GetCurrentPlaylist() ([]string, error)
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

func (a *appleMusicBridge) GetCurrentTrack() (string, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to set trackName to name of current track`, a.appName))
	output, err := cmd.Output()
	if err != nil {
		if err.Error() == "exit status 1" { // "Apple Music is not running"
			return "not playing", nil
		}

		a.log(fmt.Sprintf("Error getting current track: %v", err.Error()))
		return "", err
	}

	// Convert output to string and trim whitespace
	trackName := string(output)
	trackName = trackName[:len(trackName)-1] // Remove trailing newline

	return trackName, nil
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
		return nil
	}
}

func (a *appleMusicBridge) PreviousTrack() tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to previous track`, a.appName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error skipping to previous track: %v", err.Error()))
			return err
		}
		return nil
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

func (a *appleMusicBridge) PlayTrackByName(trackName string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command("osascript", "-e", fmt.Sprintf(`tell application "%s" to play track "%s"`, a.appName, trackName))
		if err := cmd.Run(); err != nil {
			a.log(fmt.Sprintf("Error playing track '%s': %v", trackName, err.Error()))
			return err
		}
		return nil
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

func (a *appleMusicBridge) GetCurrentPlaylist() ([]string, error) {
	cmd := exec.Command("osascript", "-e", fmt.Sprintf(`
	tell application "%s"
		set currentList to current playlist
		set output to {}
		repeat with t in tracks of currentList
			set end of output to (name of t & " - " & artist of t)
		end repeat
	end tell
	return output
	`, a.appName))
	output, err := cmd.Output()
	if err != nil {
		return []string{}, fmt.Errorf("error getting current playlist: %v", err)
	}

	// Convert output to string and split by comma
	playlist := strings.Split(string(output), ", ")
	playlistNames := []string{}
	for _, name := range playlist {
		if strings.TrimSpace(name) != "" {
			playlistNames = append(playlistNames, strings.TrimSpace(name))
		}
	}
	return playlistNames, nil
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
