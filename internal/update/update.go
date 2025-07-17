package update

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"

	"github.com/Masterminds/semver/v3"
)

const (
	ColorReset  = "\033[0m"
	ColorRed    = "\033[31m"
	ColorGreen  = "\033[32m"
	ColorYellow = "\033[33m"
	ColorCyan   = "\033[36m"
)

func fetchApi(url string) (interface{}, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result interface{}
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(respBody, &result)
	return result, err
}

func FetchUpdates() (msg, downloadUrl string, err error) {
	data, err := fetchApi(githubApi)
	if err != nil {
		return "", "", err
	}

	latestVerStr := data.(map[string]interface{})["tag_name"].(string)
	execFile := data.(map[string]interface{})["assets"].([]interface{})[0].(map[string]interface{})["browser_download_url"].(string)
	currentVer, err := semver.NewVersion(CurrentVersion)
	if err != nil {
		return "", "", err
	}

	latestVer, err := semver.NewVersion(latestVerStr)
	if err != nil {
		return "", "", err
	}

	if !currentVer.LessThan(latestVer) {
		return "Zaten en son sürümdesiniz.", "", nil
	} else {
		return fmt.Sprintf("Yeni sürüm bulundu: %s -> %s", CurrentVersion, latestVerStr), execFile, err
	}
}

func RunUpdate() error {
	fmt.Println(ColorYellow + "UYARI: Bu güncelleme yöntemi eski (deprecated) hâle gelmiştir." + ColorReset)
	fmt.Println(ColorYellow + "Lütfen anitr-cli'yi manuel olarak güncelleyiniz." + ColorReset)
	fmt.Println(ColorYellow + "Yeni sürümü edinmek için GitHub sayfasını ziyaret edin: https://github.com/kamisaki/anitr-cli" + ColorReset)
	return nil
}

func Version() {
	fmt.Printf("anitr-cli %s\nLisans: GPL 3.0 (Özgür Yazılım)\nDestek ver: %s\n\nGo sürümü: %s\n", CurrentVersion, repoLink, runtime.Version())
}

func CheckUpdates() {
	msg, _, err := FetchUpdates()

	if err != nil {
		fmt.Println(ColorRed + "Güncelleme kontrolü sırasında bir hata oluştu!" + ColorReset)
		return
	}

	if msg != "Zaten en son sürümdesiniz." {
		fmt.Println(ColorCyan + msg + ColorReset)
		fmt.Println(ColorYellow + "Yeni sürümü yüklemek için 'anitr-cli --update' komutunu kullanabilirsiniz." + ColorReset)
		os.Exit(0)
	}
}
