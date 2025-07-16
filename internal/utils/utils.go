package utils

import (
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
)

type Logger struct {
	File *os.File
}

func GetImage(url string) (string, error) {
	tempPath := filepath.Join("/tmp", "poster.png")

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	out, err := os.Create(tempPath)
	if err != nil {
		return "", err
	}

	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}

	return tempPath, nil
}

func SendNotification(title, msg, icon string) error {
	cmd := exec.Command("notify-send", "-i", icon, "-a", title, msg)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func NewLogger() (*Logger, error) {
	tmpFile, err := os.Create(filepath.Join("/tmp", "anitr-cli.log"))
	if err != nil {
		return nil, err
	}

	return &Logger{
		File: tmpFile,
	}, nil
}

func (l *Logger) LogError(err error) {
	if err == nil {
		return
	}

	log.Printf("[ERROR] %v\n", err)
}
