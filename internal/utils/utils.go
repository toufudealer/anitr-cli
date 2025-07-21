package utils

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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

	log.SetOutput(l.File)
	log.Printf("[ERROR] %v\n", err)
}

func (l *Logger) Close() {
	l.File.Close()
}

func FailIfErr(err error, logger *Logger) {
	if err != nil {
		logger.LogError(err)
		log.Fatalf("\033[31mKritik hata: %v\033[0m", err)
	}
}

func CheckErr(err error, logger *Logger) bool {
	if err != nil {
		logger.LogError(err)
		fmt.Printf("\n\033[31mHata oluştu: %v\033[0m\nLog detayları: %s\nDevam etmek için bir tuşa basın...\n", err, logger.File.Name())
		fmt.Scanln()
		return false
	}
	return true
}

func IsValidImage(url string) bool {
	client := &http.Client{}
	req, err := http.NewRequest("HEAD", url, nil)
	if err != nil {
		return false
	}
	resp, err := client.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	return resp.StatusCode == 200 && strings.HasPrefix(contentType, "image/")
}
