package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
)

const (
	ColorReset = "\033[0m"
	ColorRed   = "\033[31m"
	ColorCyan  = "\033[36m"
)

// GitHub API'den JSON verisi çeker
func fetchAPI(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("API'ye erişim başarısız: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("API'den veri okunamadı: %w", err)
	}

	var result interface{}
	if err = json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("API'den veri ayrıştırma başarısız: %w", err)
	}
	return result, nil
}

// Sürüm kontrolü yapar, yeni bir güncelleme olup olmadığını döner
func FetchUpdates() (string, error) {
	data, err := fetchAPI(githubAPI)
	if err != nil {
		return "", fmt.Errorf("güncelleme verileri alınamadı: %w", err)
	}

	latestVerStr := data.(map[string]interface{})["tag_name"].(string)

	currentVer, err := semver.NewVersion(version)
	if err != nil {
		return "", fmt.Errorf("geçerli sürüm numarası geçersiz: %v", err)
	}

	latestVer, err := semver.NewVersion(latestVerStr)
	if err != nil {
		return "", fmt.Errorf("en son sürüm numarası geçersiz: %v", err)
	}

	if !currentVer.LessThan(latestVer) {
		return "Zaten en son sürümdesiniz.", nil
	}
	return fmt.Sprintf("Yeni sürüm bulundu: %s -> %s", version, latestVerStr), nil
}

// Sürüm bilgisini döner
func Version() string {
	return fmt.Sprintf(
		"anitr-cli %s\nLisans: GPL 3.0 (Özgür Yazılım)\nDestek ver: %s\n\nGo sürümü: %s\n",
		version, repoLink, buildEnv,
	)
}

// Güncellemeleri kontrol eder ve varsa kullanıcıya bildirir
func CheckUpdates() {
	msg, err := FetchUpdates()

	if err != nil {
		fmt.Println(ColorRed + "Güncelleme kontrolü sırasında bir hata oluştu!" + ColorReset)
		time.Sleep(2 * time.Second)
		return
	}

	if msg != "Zaten en son sürümdesiniz." {
		fmt.Println(ColorCyan + msg + ColorReset)
		time.Sleep(2 * time.Second)
	}
}
