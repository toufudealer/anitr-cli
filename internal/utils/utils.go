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
	Log  *log.Logger
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
	logPath := filepath.Join("/tmp", "anitr-cli.log")
	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	multiWriter := io.MultiWriter(os.Stdout, file)

	logger := log.New(multiWriter, "", log.LstdFlags|log.Lmsgprefix)

	return &Logger{
		File: file,
		Log:  logger,
	}, nil
}

// LogError hata objesini loglar, nil ise atlar
func (l *Logger) LogError(err error) {
	if err == nil {
		return
	}
	l.Log.Printf("[ERROR] %v\n", err)
}

// LogMsg formatlı string loglamak için
func (l *Logger) LogMsg(format string, a ...interface{}) {
	l.Log.Printf(format, a...)
}

func (l *Logger) Close() error {
	return l.File.Close()
}

// FailIfErr kritik hata durumunda loglar ve kapanır
func FailIfErr(err error, logger *Logger) {
	if err != nil {
		logger.LogError(err)
		logger.LogMsg("\033[31mKritik hata: %v\033[0m\n", err)
		logger.Close()
		os.Exit(1)
	}
}

// CheckErr hata varsa loglar, ekranda gösterir ve devam etmeyi kullanıcıya bırakır
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

func NormalizeTurkishToASCII(input string) string {
	replacer := strings.NewReplacer(
		"ö", "o", "ü", "u", "ı", "i", "ç", "c", "ş", "s", "ğ", "g",
		"Ö", "O", "Ü", "U", "İ", "I", "Ç", "C", "Ş", "S", "Ğ", "G",
	)
	return replacer.Replace(input)
}
