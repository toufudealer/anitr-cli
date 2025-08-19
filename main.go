package main

import (
	"fmt"
	"log"
	"os"
	"runtime"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/flags"
	"github.com/xeyossr/anitr-cli/internal/models"
	"github.com/xeyossr/anitr-cli/internal/player"
	"github.com/xeyossr/anitr-cli/internal/rpc"
	"github.com/xeyossr/anitr-cli/internal/sources/animecix"
	"github.com/xeyossr/anitr-cli/internal/sources/openanime"
	"github.com/xeyossr/anitr-cli/internal/ui"
	"github.com/xeyossr/anitr-cli/internal/update"
	"github.com/xeyossr/anitr-cli/internal/utils"
)

// updateWatchAPI, seçilen kaynağa (animecix veya openanime) göre bir bölümün izlenebilir URL'lerini ve altyazı bilgilerini getirir.
// Ayrıca varsa TR altyazı URL'sini de döner.
// Params:
// - source: kaynak adı ("animecix", "openanime")
// - episodeData: bölüm listesi
// - index: seçilen bölümün dizindeki yeri
// - id: anime ID'si
// - seasonIndex: sezonun sıfırdan başlayan indeksi
// - selectedFansubIndex: openanime için seçilen fansub'un sırası
// - isMovie: film mi dizi mi
// - slug: openanime için gerekli olan tanımlayıcı
//
// Returns:
// - İzlenebilir kaynakları ve altyazı URL'sini içeren map[string]interface{}
// - Eğer openanime seçildiyse, fansub'ları içeren []models.Fansub
// - Hata (varsa)
func updateWatchAPI(
	source string,
	episodeData []models.Episode,
	index, id, seasonIndex, selectedFansubIndex int,
	isMovie bool,
	slug *string,
) (map[string]interface{}, []models.Fansub, error) {
	var (
		captionData []map[string]string // Video etiketleri ve URL'leri
		fansubData  []models.Fansub     // Fansub listesi (openanime için)
		captionURL  string              // Türkçe altyazı URL'si
		err         error
	)

	switch source {
	case "animecix":
		// Film ise farklı API kullan
		if isMovie {
			data, err := animecix.AnimeMovieWatchApiUrl(id)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix movie API çağrısı başarısız: %w", err)
			}
			// Caption URL ve video stream'leri al
			captionURLIface := data["caption_url"]
			captionURL, _ = captionURLIface.(string)
			streamsIface, ok := data["video_streams"]
			if !ok {
				return nil, nil, fmt.Errorf("video_streams beklenen formatta değil")
			}
			rawStreams, _ := streamsIface.([]interface{})
			for _, streamIface := range rawStreams {
				stream, _ := streamIface.(map[string]interface{})
				label := internal.GetString(stream, "label")
				url := internal.GetString(stream, "url")
				captionData = append(captionData, map[string]string{"label": label, "url": url})
			}
		} else {
			// Dizi bölümü için
			if index < 0 || index >= len(episodeData) {
				return nil, nil, fmt.Errorf("index out of range")
			}
			urlData := episodeData[index].ID
			captionData, err = animecix.AnimeWatchApiUrl(urlData)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix watch API çağrısı başarısız: %w", err)
			}
			// Sezon içerisindeki bölüm indeksini bul
			seasonEpisodeIndex := 0
			for i := 0; i < index; i++ {
				if sn, ok := episodeData[i].Extra["season_num"].(int); ok {
					if sn-1 == seasonIndex {
						seasonEpisodeIndex++
					}
				} else if snf, ok := episodeData[i].Extra["season_num"].(float64); ok {
					if int(snf)-1 == seasonIndex {
						seasonEpisodeIndex++
					}
				}
			}
			// TR altyazı URL'sini almaya çalış
			captionURL, err = animecix.FetchTRCaption(seasonIndex, seasonEpisodeIndex, id)
			if err != nil {
				captionURL = ""
			}
		}

	case "openanime":
		if slug == nil {
			return nil, nil, fmt.Errorf("slug gerekli")
		}
		if index < 0 || index >= len(episodeData) {
			return nil, nil, fmt.Errorf("index out of range")
		}
		ep := episodeData[index]
		seasonNum := 0
		episodeNum := 0

		// Sezon ve bölüm numaralarını al
		if sn, ok := ep.Extra["season_num"].(int); ok {
			seasonNum = sn
		} else if snf, ok := ep.Extra["season_num"].(float64); ok {
			seasonNum = int(snf)
		} else {
			return nil, nil, fmt.Errorf("season_num beklenen formatta değil")
		}
		if en, ok := ep.Extra["episode_num"].(int); ok {
			episodeNum = en
		} else if enf, ok := ep.Extra["episode_num"].(float64); ok {
			episodeNum = int(enf)
		} else {
			episodeNum = ep.Number
		}

		// Fansub listesini al
		fansubParams := models.FansubParams{
			Slug:       slug,
			SeasonNum:  &seasonNum,
			EpisodeNum: &episodeNum,
		}
		fansubData, err = openanime.OpenAnime{}.GetFansubsData(fansubParams)
		if err != nil {
			return nil, nil, fmt.Errorf("fansub data API çağrısı başarısız: %w", err)
		}
		if selectedFansubIndex < 0 || selectedFansubIndex >= len(fansubData) {
			return nil, nil, fmt.Errorf("seçilen fansub indeksi geçersiz")
		}

		// İzlenebilir veri isteği yap
		watchParams := models.WatchParams{
			Slug:    slug,
			Id:      &id,
			IsMovie: &isMovie,
			Extra: &map[string]interface{}{
				"season_num":         seasonNum,
				"episode_num":        episodeNum,
				"fansubs":            fansubData,
				"selected_fansub_id": selectedFansubIndex,
			},
		}
		watches, err := openanime.OpenAnime{}.GetWatchData(watchParams)
		if err != nil {
			return nil, nil, fmt.Errorf("openanime watch data alınamadı: %w", err)
		}
		if len(watches) < 1 {
			return nil, nil, fmt.Errorf("openanime watch data boş")
		}
		w := watches[0]
		captionData = make([]map[string]string, len(w.Labels))
		for i := range w.Labels {
			captionData[i] = map[string]string{
				"label": w.Labels[i],
				"url":   w.Urls[i],
			}
		}
		if w.TRCaption != nil {
			captionURL = *w.TRCaption
		}

	default:
		return nil, nil, fmt.Errorf("geçersiz kaynak: %s", source)
	}

	// Kaliteye göre (etiket sayısal değerine göre) sırala
	sort.Slice(captionData, func(i, j int) bool {
		labelI := strings.TrimRight(captionData[i]["label"], "p")
		labelJ := strings.TrimRight(captionData[j]["label"], "p")
		intI, _ := strconv.Atoi(labelI)
		intJ, _ := strconv.Atoi(labelJ)
		return intI > intJ
	})

	// Etiketleri ve URL'leri ayır
	labels := []string{}
	urls := []string{}
	for _, item := range captionData {
		labels = append(labels, item["label"])
		urls = append(urls, item["url"])
	}

	return map[string]interface{}{
		"labels":      labels,
		"urls":        urls,
		"caption_url": captionURL,
	}, fansubData, nil
}

// --- UI ve kullanıcı etkileşimi fonksiyonları ---

// Kullanıcıdan kaynak seçmesini isteyen fonksiyon
func selectSource(uiMode string, rofiFlags string, logger *utils.Logger) (string, models.AnimeSource) {
	for {
		// Ekranı temizle
		ui.ClearScreen()

		// Mevcut kaynaklar
		sourceList := []string{"OpenAnime", "AnimeciX"}

		// Kullanıcıdan seçim al
		selectedSource, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, sourceList, "Kaynak seç ")
		utils.FailIfErr(err, logger)

		// Seçilen kaynağa göre uygun AnimeSource objesini ayarla
		var source models.AnimeSource
		switch strings.ToLower(selectedSource) {
		case "openanime":
			source = openanime.OpenAnime{}
		case "animecix":
			source = animecix.AnimeCix{}
		default:
			// Geçersiz seçim yapılırsa kullanıcıyı uyar
			fmt.Printf("\033[31m[!] Geçersiz kaynak seçimi: %s\033[0m\n", selectedSource)
			time.Sleep(1500 * time.Millisecond)
			continue
		}
		return selectedSource, source
	}
}

// Kullanıcıdan arama girdisi alır ve API üzerinden sonuçları getirir
func searchAnime(source models.AnimeSource, uiMode string, rofiFlags string, logger *utils.Logger) ([]models.Anime, []string, []string, map[string]models.Anime) {
	for {
		// Kullanıcıdan arama kelimesi al
		query, err := ui.InputFromUser(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Anime ara "})
		utils.FailIfErr(err, logger)

		// API üzerinden arama yap
		searchData, err := source.GetSearchData(query)
		utils.FailIfErr(err, logger)

		// Hiç sonuç çıkmazsa kullanıcıyı bilgilendir
		if searchData == nil {
			fmt.Printf("\033[31m[!] Arama sonucu bulunamadı!\033[0m")
			time.Sleep(1500 * time.Millisecond)
			continue
		}

		// Arama sonuçlarını işleyip ilgili listeleri oluştur
		animeNames := make([]string, 0, len(searchData))
		animeTypes := make([]string, 0, len(searchData))
		animeMap := make(map[string]models.Anime)

		for _, item := range searchData {
			animeNames = append(animeNames, item.Title)
			animeMap[item.Title] = item

			// Anime türünü belirle (tv veya movie)
			if item.TitleType != nil {
				ttype := item.TitleType
				if strings.ToLower(*ttype) == "movie" {
					animeTypes = append(animeTypes, "movie")
				} else {
					animeTypes = append(animeTypes, "tv")
				}
			}
		}

		return searchData, animeNames, animeTypes, animeMap
	}
}

// Kullanıcının seçtiği animeyi belirler
func selectAnime(animeNames []string, searchData []models.Anime, uiMode string, isMovie bool, rofiFlags string, animeTypes []string, logger *utils.Logger) (models.Anime, bool, int) {
	for {
		ui.ClearScreen()

		// Kullanıcıdan anime seçimi al
		selectedAnimeName, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, animeNames, "Anime seç ")
		utils.FailIfErr(err, logger)

		// Geçerli bir anime ismi mi kontrol et
		if !slices.Contains(animeNames, selectedAnimeName) {
			continue
		}

		// Seçilen animeyi bul
		selectedIndex := slices.Index(animeNames, selectedAnimeName)
		selectedAnime := searchData[selectedIndex]

		// Anime türü (movie / tv) güncelleniyor
		if len(animeTypes) > 0 {
			selectedAnimeType := animeTypes[selectedIndex]
			isMovie = selectedAnimeType == "movie"
		}

		return selectedAnime, isMovie, selectedIndex
	}
}

// Seçilen animenin ID veya slug bilgisini döner
func getAnimeIDs(source models.AnimeSource, selectedAnime models.Anime) (int, string) {
	var selectedAnimeID int
	var selectedAnimeSlug string

	// Kaynağa göre ID veya slug alınır
	if source.Source() == "animecix" {
		selectedID := selectedAnime.ID
		selectedAnimeID = *selectedID
	} else if source.Source() == "openanime" {
		selectedSlug := selectedAnime.Slug
		selectedAnimeSlug = *selectedSlug
	}
	return selectedAnimeID, selectedAnimeSlug
}

// Seçilen animeye ait bölümleri getirir, isim listesi oluşturur ve movie olup olmadığını döner
func getEpisodesAndNames(source models.AnimeSource, isMovie bool, selectedAnimeID int, selectedAnimeSlug string, selectedAnimeName string, logger *utils.Logger) ([]models.Episode, []string, bool, int, error) {
	var (
		episodes            []models.Episode
		episodeNames        []string
		selectedSeasonIndex int
		err                 error
	)

	// OpenAnime ise sezon verisini alarak movie olup olmadığını kontrol et
	if source.Source() == "openanime" {
		seasonData, err := source.GetSeasonsData(models.SeasonParams{Slug: &selectedAnimeSlug})
		if err != nil {
			return nil, nil, false, 0, fmt.Errorf("sezon verisi alınamadı: %w", err)
		}
		isMovie = *seasonData[0].IsMovie
	}

	if !isMovie {
		// Dizi ise bölüm verilerini al
		episodes, err = source.GetEpisodesData(models.EpisodeParams{SeasonID: &selectedAnimeID, Slug: &selectedAnimeSlug})
		if err != nil {
			return nil, nil, false, 0, fmt.Errorf("bölüm verisi alınamadı: %w", err)
		}

		if len(episodes) == 0 {
			return nil, nil, false, 0, fmt.Errorf("hiçbir bölüm bulunamadı")
		}

		// Bölüm isimlerini listeye ekle
		episodeNames = make([]string, 0, len(episodes))
		for _, e := range episodes {
			episodeNames = append(episodeNames, e.Title)
		}

		// Sezon indeksini belirle
		selectedSeasonIndex = int(episodes[0].Extra["season_num"].(float64)) - 1
	} else {
		// Film ise sadece tek bir bölüm olarak ayarla
		episodeNames = []string{selectedAnimeName}
		episodes = []models.Episode{{
			Title: selectedAnimeName,
			Extra: map[string]interface{}{"season_num": float64(1)},
		}}
		selectedSeasonIndex = 0
	}

	return episodes, episodeNames, isMovie, selectedSeasonIndex, nil
}

// Seçilen animeyi oynatma döngüsünü yönetir.
// Kullanıcıdan izleme seçenekleri alır, çözünürlük/fansub seçtirir, animeyi oynatır ve Discord RPC'yi günceller.
func playAnimeLoop(
	source models.AnimeSource, // Seçilen anime kaynağı (OpenAnime, AnimeciX)
	selectedSource string, // Seçilen kaynak ismi
	episodes []models.Episode, // Tüm bölümler
	episodeNames []string, // Bölüm adları
	selectedAnimeID int, // Seçilen anime ID'si (AnimeciX için)
	selectedAnimeSlug string, // Seçilen anime slug'ı (OpenAnime için)
	selectedAnimeName string, // Seçilen anime ismi
	isMovie bool, // Film mi yoksa dizi mi olduğunu belirtir
	selectedSeasonIndex int, // Seçilen sezonun index'i
	uiMode string, // Arayüz tipi (örneğin terminal, rofi, vs.)
	rofiFlags string, // Rofi için özel bayraklar
	posterURL string, // Poster görseli URL'si (Discord RPC için)
	disableRPC bool, // Discord RPC devre dışı mı?
	logger *utils.Logger, // Logger
) (models.AnimeSource, string) { // Geriye güncel kaynak ve kaynak ismi döner

	selectedEpisodeIndex := 0
	selectedFansubIdx := 0
	selectedResolution := ""
	selectedResolutionIdx := 0

	// Discord RPC için giriş yap
	loggedIn, err := rpc.ClientLogin()
	if err != nil || !loggedIn {
		logger.LogError(err)
	}

	for {
		ui.ClearScreen()

		// Kullanıcıya sunulacak menü seçenekleri
		watchMenu := []string{}
		if !isMovie {
			watchMenu = append(watchMenu, "İzle", "Sonraki bölüm", "Önceki bölüm", "Bölüm seç", "Çözünürlük seç")
		} else {
			watchMenu = append(watchMenu, "İzle", "Çözünürlük seç")
		}

		// OpenAnime için fansub seçimi
		if strings.ToLower(selectedSource) == "openanime" {
			watchMenu = append(watchMenu, "Fansub seç")
		}

		// Genel seçenekler
		watchMenu = append(watchMenu, "Anime ara", "Çık")

		// Seçim arayüzünü göster
		option, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, watchMenu, selectedAnimeName)
		utils.FailIfErr(err, logger)

		switch option {

		// Oynatma ve bölüm gezme seçenekleri
		case "İzle", "Sonraki bölüm", "Önceki bölüm":
			ui.ClearScreen()

			if option == "Sonraki bölüm" {
				if selectedEpisodeIndex+1 >= len(episodes) {
					fmt.Println("Zaten son bölümdesiniz.")
					break
				}
				selectedEpisodeIndex++
			} else if option == "Önceki bölüm" {
				if selectedEpisodeIndex <= 0 {
					fmt.Println("Zaten ilk bölümdesiniz.")
					break
				}
				selectedEpisodeIndex--
			}

			// Güncel sezon bilgisi al
			selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1

			// API'den oynatma bilgilerini güncelle
			data, _, err := updateWatchAPI(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if err != nil {
				fmt.Printf("\033[31m[!] Bölüm oynatılamadı: %s\033[0m\n", err)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			labels := data["labels"].([]string)
			urls := data["urls"].([]string)
			subtitle := data["caption_url"].(string)

			// Varsayılan çözünürlük seçimi
			if selectedResolution == "" {
				selectedResolutionIdx = 0
				if len(labels) > 0 {
					selectedResolution = labels[selectedResolutionIdx]
				}
			}
			if selectedResolutionIdx >= len(urls) {
				selectedResolutionIdx = len(urls) - 1
			}

			// MPV başlığı ayarla
			mpvTitle := fmt.Sprintf("%s - %s", selectedAnimeName, episodeNames[selectedEpisodeIndex])
			if isMovie {
				mpvTitle = selectedAnimeName
			}

			// VLC ile oynat
			cmd, _, err := player.PlayVLC(player.VLCParams{
				Url:         urls[selectedResolutionIdx],
				SubtitleUrl: &subtitle,
				Title:       mpvTitle,
				VLCPath:     "", // Assuming VLC is in PATH, otherwise specify
			})
			if !utils.CheckErr(err, logger) {
				return source, selectedSource
			}

			// Discord RPC'yi başlat
			if !disableRPC {
				go updateDiscordRPC(episodeNames, selectedEpisodeIndex, selectedAnimeName, selectedSource, posterURL, logger, &loggedIn)
			}

			// Oynatma işlemi tamamlanana kadar bekle
			err = cmd.Wait()
			if err != nil {
				fmt.Println("VLC çalışırken hata:", err)
			}

		// Çözünürlük seçme ekranı
		case "Çözünürlük seç":
			data, _, err := updateWatchAPI(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if err != nil {
				fmt.Printf("\033[31m[!] Çözünürlükler yüklenemedi.\033[0m\n")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			labels := data["labels"].([]string)
			selected, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, labels, "Çözünürlük seç ")
			if !utils.CheckErr(err, logger) {
				continue
			}
			selectedResolution = selected
			if !slices.Contains(labels, selected) {
				fmt.Printf("\033[31m[!] Geçersiz çözünürlük seçimi: %s\033[0m\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedResolutionIdx = slices.Index(labels, selected)

		// Bölüm seçimi
		case "Bölüm seç":
			selected, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, episodeNames, "Bölüm seç ")
			if !utils.CheckErr(err, logger) {
				continue
			}
			if slices.Contains(episodeNames, selected) {
				selectedEpisodeIndex = slices.Index(episodeNames, selected)
				if !isMovie && selectedEpisodeIndex >= 0 && selectedEpisodeIndex < len(episodes) {
					selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1
				}
			} else {
				continue
			}

		// Fansub seçimi (yalnızca OpenAnime için)
		case "Fansub seç":
			fansubNames := []string{}

			if strings.ToLower(source.Source()) != "openanime" {
				fmt.Println("\033[31m[!] Bu seçenek sadece OpenAnime için geçerlidir.\033[0m")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			_, fansubData, err := updateWatchAPI(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)

			if err != nil {
				fmt.Printf("\033[31m[!] Fansublar yüklenemedi.\033[0m\n")
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			for _, fansub := range fansubData {
				if fansub.Name != nil {
					fansubNames = append(fansubNames, *fansub.Name)
				}
			}

			selected, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, fansubNames, "Fansub seç ")
			if !utils.CheckErr(err, logger) {
				continue
			}

			if !slices.Contains(fansubNames, selected) {
				fmt.Printf("\033[31m[!] Geçersiz fansub seçimi: %s\033[0m\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedFansubIdx = slices.Index(fansubNames, selected)

		// Yeni bir anime aramak için menü
		case "Anime ara":
			for {
				choice, err := showSelection(App{uiMode: &uiMode, rofiFlags: &rofiFlags}, []string{"Bu kaynakla devam et", "Kaynak değiştir", "Çık"}, fmt.Sprintf("Arama kaynağı: %s", selectedSource))
				if err != nil {
					logger.LogError(fmt.Errorf("seçim listesi oluşturulamadı: %w", err))
					continue
				}

				switch choice {
				case "Bu kaynakla devam et":
					// Hiçbir işlem yapma
				case "Kaynak değiştir":
					selectedSource, source = selectSource(uiMode, rofiFlags, logger)
				case "Çık":
					os.Exit(0)
				default:
					fmt.Printf("\033[31m[!] Geçersiz seçim: %s\033[0m\n", choice)
					time.Sleep(1500 * time.Millisecond)
					continue
				}

				return source, selectedSource
			}

		// Çıkış seçeneği
		case "Çık":
			os.Exit(0)

		default:
			return source, selectedSource
		}
	}
}

// Discord RPC'yi güncelleyerek anime oynatma durumunu Discord'a yansıtır
func updateDiscordRPC(episodeNames []string, selectedEpisodeIndex int, selectedAnimeName, selectedSource, posterURL string, logger *utils.Logger, loggedIn *bool) {
	// Discord RPC'yi başlat
	if !*loggedIn {
		var err error
		*loggedIn, err = rpc.ClientLogin()
		if err != nil {
			logger.LogError(err)
			return
		}
	}

	// Discord'da gösterilecek durum bilgisini oluştur
	state := fmt.Sprintf("Bölüm %d: %s", selectedEpisodeIndex+1, episodeNames[selectedEpisodeIndex])

	// Discord RPC için parametreleri ayarla ve RPC'yi güncelle
	var err2 error
	*loggedIn, err2 = rpc.DiscordRPC(internal.RPCParams{
		Details:    selectedAnimeName,
		State:      state,
		SmallImage: strings.ToLower(selectedSource),
		SmallText:  selectedSource,
		LargeImage: posterURL,
		LargeText:  selectedAnimeName,
	}, *loggedIn)

	// Discord RPC güncelleme hatası varsa logla
	if err2 != nil {
		logger.LogError(fmt.Errorf("DiscordRPC hatası: %w", err2))
	}
}

// Uygulama durumu ve ayarlarını saklayan struct
type App struct {
	source         *models.AnimeSource
	selectedSource *string
	uiMode         *string
	rofiFlags      *string
	disableRPC     *bool
	logger         *utils.Logger
}

// Kullanıcıdan bir seçim almak için kullanılan fonksiyon
func showSelection(cfx App, list []string, label string) (string, error) {
	return ui.SelectionList(internal.UiParams{
		Mode:      *cfx.uiMode,
		RofiFlags: cfx.rofiFlags,
		List:      &list,
		Label:     label,
	})
}

// Uygulamanın ana fonksiyonu, anime seçimi, oynatma ve hata yönetimini içerir
func app(cfx *App) error {
	for {
		// Anime arama işlemi yapılır
		searchData, animeNames, animeTypes, _ := searchAnime(*cfx.source, *cfx.uiMode, *cfx.rofiFlags, cfx.logger)
		isMovie := false
		// Kullanıcıdan anime seçimi yapılması istenir
		selectedAnime, isMovie, _ := selectAnime(animeNames, searchData, *cfx.uiMode, isMovie, *cfx.rofiFlags, animeTypes, cfx.logger)

		// Poster URL'si alınır ve geçersizse varsayılan bir URL kullanılır
		posterURL := selectedAnime.ImageURL
		if !utils.IsValidImage(posterURL) {
			posterURL = "anitrcli"
		}

		// Seçilen animeye ait ID ve slug alınır
		selectedAnimeID, selectedAnimeSlug := getAnimeIDs(*cfx.source, selectedAnime)

		// Anime bölümleri alınır
		episodes, episodeNames, isMovie, selectedSeasonIndex, err := getEpisodesAndNames(
			*cfx.source, isMovie, selectedAnimeID, selectedAnimeSlug, selectedAnime.Title, cfx.logger,
		)

		// Hata durumunda kullanıcıya seçenek sunulur
		if err != nil {
			cfx.logger.LogError(err)

			choice, err := showSelection(App{uiMode: cfx.uiMode, rofiFlags: cfx.rofiFlags}, []string{"Farklı Anime Ara", "Kaynak Değiştir", "Çık"}, fmt.Sprintf("Hata: %s", err.Error()))
			if err != nil {
				os.Exit(0)
			}

			// Kullanıcının seçimine göre işlem yapılır
			switch choice {
			case "Farklı Anime Ara":
				return nil // Üst döngüye geri dön
			case "Kaynak Değiştir":
				selectedSource, source := selectSource(*cfx.uiMode, *cfx.rofiFlags, cfx.logger)
				cfx.selectedSource = utils.Ptr(selectedSource)
				cfx.source = utils.Ptr(source)
				return nil
			default:
				os.Exit(0)
			}
		}

		// Oynatma döngüsüne girilir
		newSource, newSelectedSource := playAnimeLoop(
			*cfx.source, *cfx.selectedSource, episodes, episodeNames,
			selectedAnimeID, selectedAnimeSlug, selectedAnime.Title,
			isMovie, selectedSeasonIndex, *cfx.uiMode, *cfx.rofiFlags,
			posterURL, *cfx.disableRPC, cfx.logger,
		)

		// Kaynak değiştiyse güncellenir
		if newSource != *cfx.source || newSelectedSource != *cfx.selectedSource {
			cfx.source = &newSource
			cfx.selectedSource = &newSelectedSource
			return nil
		}
	}
}

// Ana uygulama döngüsünü yöneten fonksiyon
func runMain(f *flags.Flags, uiMode string, logger *utils.Logger) {
	// RPC'yi devre dışı bırakma bayrağı ayarlanır
	disableRPC := f.DisableRPC

	// Güncellemeleri kontrol et
	update.CheckUpdates()

	// Uygulama durumunu başlat
	currentApp := &App{
		source:         nil,
		selectedSource: utils.Ptr(""),
		uiMode:         &uiMode,
		rofiFlags:      &f.RofiFlags,
		disableRPC:     &disableRPC,
		logger:         logger,
	}

	for {
		// Eğer kaynak seçilmemişse, kaynak seçme işlemi yapılır
		if currentApp.source == nil {
			selectedSource, source := selectSource(uiMode, f.RofiFlags, logger)
			currentApp.selectedSource = utils.Ptr(selectedSource)
			currentApp.source = utils.Ptr(source)
		}

		// Uygulama fonksiyonu çağrılır
		if err := app(currentApp); err != nil {
			// Hata durumunda loglanır ve kaynak sıfırlanır
			logger.LogError(err)
			currentApp.source = nil
		}
	}
}

// Uygulama komutlarını çalıştıran giriş fonksiyonu
func runApp() {
	logger, err := utils.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	log.SetFlags(0)

	rootCmd, f := flags.NewFlagsCmd()

	commands := rootCmd.Commands()

	if runtime.GOOS != "linux" {
		// Windows ve Mac'te alt komut yok, doğrudan tui modunda çalıştır
		rootCmd.Run = func(cmd *cobra.Command, args []string) {
			f.RofiMode = false
			runMain(f, "tui", logger)
		}
	} else {
		// Linux için alt komutlar varsa ayarla
		var rofiCmd, tuiCmd *cobra.Command
		if len(commands) > 0 {
			rofiCmd = commands[0]
		}
		if len(commands) > 1 {
			tuiCmd = commands[1]
		}

		if rofiCmd != nil {
			rofiCmd.Run = func(cmd *cobra.Command, args []string) {
				f.RofiMode = true
				runMain(f, "rofi", logger)
			}
		}

		if tuiCmd != nil {
			tuiCmd.Run = func(cmd *cobra.Command, args []string) {
				f.RofiMode = false
				runMain(f, "tui", logger)
			}
		}

		rootCmd.Run = func(cmd *cobra.Command, args []string) {
			f.RofiMode = false
			runMain(f, "tui", logger)
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	// Uygulamayı başlat
	runApp()
}
