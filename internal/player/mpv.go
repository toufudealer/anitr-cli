package player

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/xeyossr/anitr-cli/internal/ipc"
)

type MPVParams struct {
	Url         string
	SubtitleUrl *string
	Title       string
}

func isMPVInstalled() error {
	_, err := exec.LookPath("mpv")
	return err
}

func Play(params MPVParams) (*exec.Cmd, string, error) {
	mpvSocket := "anitr-cli-410.sock"
	mpvSocketPath := filepath.Join("/tmp", mpvSocket)

	if err := isMPVInstalled(); err != nil {
		return nil, "", errors.New("mpv sisteminizde yüklü değil")
	}

	args := []string{
		"--fullscreen",
		"--user-agent=Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/137.0.0.0 Safari/537.36",
		"--referrer=https://yeshi.eu.org/",
		"--save-position-on-quit",
		fmt.Sprintf("--title=%s", params.Title),
		fmt.Sprintf("--force-media-title=%s", params.Title),
		"--idle=yes", "--really-quiet", "--no-terminal",
		fmt.Sprintf("--input-ipc-server=%s", mpvSocketPath),
	}

	if params.SubtitleUrl != nil && *params.SubtitleUrl != "" {
		args = append(args, "--sub-file", *params.SubtitleUrl)
	}

	args = append(args, params.Url)

	cmd := exec.Command("mpv", args...)
	if err := cmd.Start(); err != nil {
		return cmd, "", err
	}

	maxRetries := 10
	retryDelay := 300 * time.Millisecond
	for i := 0; i < maxRetries; i++ {
		time.Sleep(retryDelay)
		conn, err := ipc.ConnectToPipe(mpvSocketPath)
		if err == nil {
			conn.Close()
			return cmd, mpvSocketPath, nil
		}
	}

	return cmd, "", errors.New("MPV socket hazır değil, başlatılamadı")
}

func MPVSendCommand(ipcSocketPath string, command []interface{}) (interface{}, error) {
	var lastErr error
	maxRetries := 3
	retryDelay := 100 * time.Millisecond

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(retryDelay)
		}

		conn, err := ipc.ConnectToPipe(ipcSocketPath)
		if err != nil {
			lastErr = err
			continue
		}
		defer func() { _ = conn.Close() }()

		commandStr, err := json.Marshal(map[string]interface{}{
			"command": command,
		})
		if err != nil {
			return nil, err
		}

		_, err = conn.Write(append(commandStr, '\n'))
		if err != nil {
			lastErr = err
			continue
		}

		buf := make([]byte, 4096)
		if deadline, ok := conn.(interface{ SetReadDeadline(time.Time) error }); ok {
			deadline.SetReadDeadline(time.Now().Add(1 * time.Second))
		}

		n, err := conn.Read(buf)
		if err != nil {
			lastErr = err
			continue
		}

		var response map[string]interface{}
		if err := json.Unmarshal(buf[:n], &response); err != nil {
			lastErr = err
			continue
		}

		if data, exists := response["data"]; exists {
			return data, nil
		}
		return nil, nil
	}

	return nil, fmt.Errorf("command failed after %d attempts: %w", maxRetries, lastErr)
}

func SeekMPV(ipcSocketPath string, time int) (interface{}, error) {
	command := []interface{}{"seek", time, "absolute"}
	return MPVSendCommand(ipcSocketPath, command)
}

func GetMPVPausedStatus(ipcSocketPath string) (bool, error) {
	status, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "pause"})
	if err != nil || status == nil {
		return false, err
	}

	paused, ok := status.(bool)
	if ok {
		return paused, nil
	}
	return false, nil
}

func GetMPVPlaybackSpeed(ipcSocketPath string) (float64, error) {
	speed, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "speed"})
	if err != nil || speed == nil {
		return 0, err
	}

	currentSpeed, ok := speed.(float64)
	if ok {
		return currentSpeed, nil
	}

	return 0, nil
}

func GetPercentageWatched(ipcSocketPath string) (float64, error) {
	currentTime, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "time-pos"})
	if err != nil || currentTime == nil {
		return 0, err
	}

	duration, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "duration"})
	if err != nil || duration == nil {
		return 0, err
	}

	currTime, ok1 := currentTime.(float64)
	dur, ok2 := duration.(float64)

	if ok1 && ok2 && dur > 0 {
		percentageWatched := (currTime / dur) * 100
		return percentageWatched, nil
	}

	return 0, nil
}

func PercentageWatched(playbackTime int, duration int) float64 {
	if duration > 0 {
		percentage := (float64(playbackTime) / float64(duration)) * 100
		return percentage
	}
	return float64(0)
}

func HasActivePlayback(ipcSocketPath string) (bool, error) {
	maxRetries := 3
	var lastErr error

	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(200 * time.Millisecond)
		}

		timePos, err := MPVSendCommand(ipcSocketPath, []interface{}{"get_property", "time-pos"})

		if err != nil {
			if strings.Contains(err.Error(), "property unavailable") {
				return false, nil
			}

			if strings.Contains(err.Error(), "connect: connection refused") ||
				strings.Contains(err.Error(), "connect: no such file or directory") {
				lastErr = err
				continue
			}

			return false, fmt.Errorf("error getting time-pos: %w", err)
		}

		if timePos != nil {
			return true, nil
		}

		return false, nil
	}

	return false, fmt.Errorf("failed to check playback status: %w", lastErr)
}

func IsMPVRunning(socketPath string) bool {
	if socketPath == "" {
		return false
	}

	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(200 * time.Millisecond)
		}

		conn, err := ipc.ConnectToPipe(socketPath)
		if err != nil {
			continue
		}
		defer conn.Close()

		_, err = MPVSendCommand(socketPath, []interface{}{"get_property", "pid"})
		if err == nil {
			return true
		}

	}

	return false
}
