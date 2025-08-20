package downloader

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/schollz/progressbar/v3"
)

// DownloadFile downloads a file from the given URL to the specified filepath.
func DownloadFile(url string, filepath string) error {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("yeni istek oluşturulamadı: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("istek yapılamadı: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hatalı durum: %s", resp.Status)
	}

	f, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("dosya oluşturulamadı: %w", err)
	}
	defer f.Close()

	size, _ := strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)

	bar := progressbar.NewOptions64(size, 
		progressbar.OptionSetDescription(fmt.Sprintf("İndiriliyor: %s", filepath)),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(40),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() {
			fmt.Fprint(os.Stderr, "\n")
		}),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}),
	)
	bar.RenderBlank()

	_, err = io.Copy(io.MultiWriter(f, bar), resp.Body)
	if err != nil {
		return fmt.Errorf("dosyaya yazılamadı: %w", err)
	}

	return nil
}
