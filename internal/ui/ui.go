package ui

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/ui/rofi"
	"github.com/xeyossr/anitr-cli/internal/ui/tui"
)

// Ekranı temizler
func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// Kullanıcıya seçim listesi gösterir
// Mode rofi ise rofi arayüzü, değilse tui kullanılır
func SelectionList(params internal.UiParams) ([]string, error) {
	if params.Mode == "rofi" {
		response, err := rofi.SelectionList(params)
		if err != nil {
			return nil, fmt.Errorf("rofi seçim listesi oluşturulamadı: %w", err)
		}
		return []string{response}, nil // Wrap single string in a slice
	}

	response, err := tui.SelectionList(params)
	if err != nil {
		return nil, fmt.Errorf("tui seçim listesi oluşturulamadı: %w", err)
	}
	return response, nil
}

// Kullanıcıdan input almak için
// rofi ya da tui üzerinden alınır
func InputFromUser(params internal.UiParams) (string, error) {
	if params.Mode == "rofi" {
		response, err := rofi.InputFromUser(params)
		if err != nil {
			return "", fmt.Errorf("rofi kullanıcı girişi alınamadı: %w", err)
		}
		return response, nil
	}

	response, err := tui.InputFromUser(params)
	if err != nil {
		return "", fmt.Errorf("tui kullanıcı girişi alınamadı: %w", err)
	}
	return response, nil
}
