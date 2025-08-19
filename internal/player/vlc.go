package player

import (
	"errors"
	"fmt"
	"os/exec"
	"runtime"
)

// VLCParams struct, VLC oynatıcı parametrelerini tutar.
type VLCParams struct {
	Url         string  // Oynatılacak video URL'si
	SubtitleUrl *string // Altyazı URL'si (isteğe bağlı)
	Title       string  // Video başlığı
	VLCPath     string
}

// isVLCInstalled fonksiyonu, sistemde VLC oynatıcısının yüklü olup olmadığını kontrol eder.
func isVLCInstalled(vlcPath string) error {
	vlcBinary := "vlc"
	if vlcPath != "" {
		vlcBinary = vlcPath
	} else if runtime.GOOS == "windows" {
		vlcBinary = "vlc.exe"
	}
	_, err := exec.LookPath(vlcBinary) // VLC'nin sistemde bulunup bulunmadığını kontrol eder
	return err
}

// getVLCBinary platforma göre vlc binary adını döner
func getVLCBinary(vlcPath string) string {
	if vlcPath != "" {
		return vlcPath
	}
	if runtime.GOOS == "windows" {
		return "vlc.exe"
	}
	return "vlc"
}

// PlayVLC fonksiyonu, verilen parametrelerle VLC oynatıcıyı başlatır.
func PlayVLC(params VLCParams) (*exec.Cmd, string, error) {
	// VLC'nin yüklü olup olmadığını kontrol et
	if err := isVLCInstalled(params.VLCPath); err != nil {
		return nil, "", errors.New("VLC sisteminizde yüklü değil veya belirtilen yolda bulunamadı") // Yükleme hatası
	}

	// VLC başlatma komutunu oluştur
	args := []string{
		"--fullscreen", // Tam ekran başlat
		"--play-and-exit", // Oynatma bitince VLC'yi kapat
		fmt.Sprintf("--meta-title=%s", params.Title), // Pencere başlığını ayarla
	}

	// Eğer altyazı URL'si varsa, altyazı dosyasını ekle
	if params.SubtitleUrl != nil && *params.SubtitleUrl != "" {
		args = append(args, fmt.Sprintf("--sub-file=%s", *params.SubtitleUrl))
	}

	// Video URL'sini ekle
	args = append(args, params.Url)

	vlcBinary := getVLCBinary(params.VLCPath)
	cmd := exec.Command(vlcBinary, args...)
	if err := cmd.Start(); err != nil {
		return cmd, "", err // Başlatma hatası
	}

	return cmd, "", nil
}