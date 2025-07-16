package player

import (
	"errors"
	"os/exec"
)

func isMPVInstalled() error {
	_, err := exec.LookPath("mpv")
	return err
}

func Play(url string, subtitleUrl *string) error {
	err := isMPVInstalled()
	if err != nil {
		return errors.New("mpv sisteminizde yüklü değil")
	}

	args := []string{"--fullscreen", "--user-agent=Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/137.0.0.0 Safari/537.36", "--referrer=https://yeshi.eu.org/", "--save-position-on-quit"}

	if subtitleUrl != nil && *subtitleUrl != "" {
		subtitle := []string{"--sub-file", *subtitleUrl}
		args = append(args, subtitle...)
	}

	args = append(args, url)

	cmd := exec.Command("mpv", args...)
	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}
