package main

import (
	"errors"
	"fmt"
	"log"
	"os"
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

// --- API ve veri işleme fonksiyonları ---

func updateWatchApi(
	source string,
	episodeData []models.Episode,
	index, id, seasonIndex, selectedFansubIndex int,
	isMovie bool,
	slug *string,
) (map[string]interface{}, []models.Fansub, error) {
	var (
		captionData []map[string]string
		fansubData  []models.Fansub
		captionUrl  string
		err         error
	)

	switch source {
	case "animecix":
		if isMovie {
			data, err := animecix.AnimeMovieWatchApiUrl(id)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix watch api url alınamadı: %w", err)
			}
			captionUrlIface := data["caption_url"]
			captionUrl, _ = captionUrlIface.(string)
			streamsIface, ok := data["video_streams"]
			if !ok {
				return nil, nil, fmt.Errorf("video_streams bulunamadı")
			}
			rawStreams, _ := streamsIface.([]interface{})
			for _, streamIface := range rawStreams {
				stream, _ := streamIface.(map[string]interface{})
				label := internal.GetString(stream, "label")
				url := internal.GetString(stream, "url")
				captionData = append(captionData, map[string]string{"label": label, "url": url})
			}
		} else {
			if index < 0 || index >= len(episodeData) {
				return nil, nil, fmt.Errorf("index out of range")
			}
			urlData := episodeData[index].ID
			captionData, err = animecix.AnimeWatchApiUrl(urlData)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix watch api url alınamadı: %w", err)
			}
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
			captionUrl, err = animecix.FetchTRCaption(seasonIndex, seasonEpisodeIndex, id)
			if err != nil {
				captionUrl = ""
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
		if sn, ok := ep.Extra["season_num"].(int); ok {
			seasonNum = sn
		} else if snf, ok := ep.Extra["season_num"].(float64); ok {
			seasonNum = int(snf)
		} else {
			return nil, nil, fmt.Errorf("season_num geçersiz veya eksik")
		}
		if en, ok := ep.Extra["episode_num"].(int); ok {
			episodeNum = en
		} else if enf, ok := ep.Extra["episode_num"].(float64); ok {
			episodeNum = int(enf)
		} else {
			episodeNum = ep.Number
		}

		fansubParams := models.FansubParams{
			Slug:       slug,
			SeasonNum:  &seasonNum,
			EpisodeNum: &episodeNum,
		}

		fansubData, err = openanime.OpenAnime{}.GetFansubsData(fansubParams)
		if err != nil {
			return nil, nil, fmt.Errorf("openanime fansub data alınamadı: %w", err)
		}

		if selectedFansubIndex < 0 || selectedFansubIndex >= len(fansubData) {
			return nil, nil, fmt.Errorf("seçilen fansub indeksi geçersiz")
		}

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
			captionUrl = *w.TRCaption
		}
	default:
		return nil, nil, fmt.Errorf("geçersiz kaynak: %s", source)
	}
	sort.Slice(captionData, func(i, j int) bool {
		labelI := strings.TrimRight(captionData[i]["label"], "p")
		labelJ := strings.TrimRight(captionData[j]["label"], "p")
		intI, _ := strconv.Atoi(labelI)
		intJ, _ := strconv.Atoi(labelJ)
		return intI > intJ
	})
	labels := []string{}
	urls := []string{}
	for _, item := range captionData {
		labels = append(labels, item["label"])
		urls = append(urls, item["url"])
	}
	return map[string]interface{}{
		"labels":      labels,
		"urls":        urls,
		"caption_url": captionUrl,
	}, fansubData, nil
}

// --- UI ve kullanıcı etkileşimi fonksiyonları ---

func selectSource(uiMode string, rofiFlags string, logger *utils.Logger) (string, models.AnimeSource) {
	for {
		ui.ClearScreen()
		sourceList := []string{"OpenAnime", "AnimeciX"}
		selectedSource, err := ui.SelectionList(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Kaynak seç ", List: &sourceList})
		utils.FailIfErr(err, logger)
		var source models.AnimeSource
		switch strings.ToLower(selectedSource) {
		case "openanime":
			source = openanime.OpenAnime{}
		case "animecix":
			source = animecix.AnimeCix{}
		default:
			fmt.Printf("\033[31m[!] Geçersiz kaynak seçimi: %s\033[0m\n", selectedSource)
			time.Sleep(1500 * time.Millisecond)
			continue
		}
		return selectedSource, source
	}
}

func searchAnime(source models.AnimeSource, uiMode string, rofiFlags string, logger *utils.Logger) ([]models.Anime, []string, []string, map[string]models.Anime) {
	for {
		query, err := ui.InputFromUser(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Anime ara "})
		utils.FailIfErr(err, logger)
		searchData, err := source.GetSearchData(query)
		utils.FailIfErr(err, logger)
		if searchData == nil {
			fmt.Printf("\033[31m[!] Arama sonucu bulunamadı!\033[0m")
			time.Sleep(1500 * time.Millisecond)
			continue
		}
		animeNames := []string{}
		animeTypes := []string{}

		animeMap := make(map[string]models.Anime)
		for _, item := range searchData {
			animeNames = append(animeNames, item.Title)
			animeMap[item.Title] = item

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

func selectAnime(animeNames []string, searchData []models.Anime, uiMode string, isMovie bool, rofiFlags string, animeTypes []string, logger *utils.Logger) (models.Anime, bool, int) {
	for {
		ui.ClearScreen()
		selectedAnimeName, err := ui.SelectionList(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, List: &animeNames, Label: "Anime seç "})
		utils.FailIfErr(err, logger)
		if !slices.Contains(animeNames, selectedAnimeName) {
			continue
		}

		selectedIndex := slices.Index(animeNames, selectedAnimeName)
		selectedAnime := searchData[selectedIndex]

		if len(animeTypes) > 0 {
			selectedAnimeType := animeTypes[selectedIndex]
			isMovie = selectedAnimeType == "movie"
		}

		return selectedAnime, isMovie, selectedIndex
	}
}

func getAnimeIDs(source models.AnimeSource, selectedAnime models.Anime) (int, string) {
	var selectedAnimeID int
	var selectedAnimeSlug string
	if source.Source() == "animecix" {
		selectedId := selectedAnime.ID
		selectedAnimeID = *selectedId
	} else if source.Source() == "openanime" {
		selectedSlug := selectedAnime.Slug
		selectedAnimeSlug = *selectedSlug
	}
	return selectedAnimeID, selectedAnimeSlug
}

func getEpisodesAndNames(source models.AnimeSource, isMovie bool, selectedAnimeID int, selectedAnimeSlug string, selectedAnimeName string, logger *utils.Logger) ([]models.Episode, []string, bool, int) {
	var (
		episodes            []models.Episode
		episodeNames        []string
		selectedSeasonIndex int
		err                 error
	)

	if source.Source() == "openanime" {
		seasonData, err := source.GetSeasonsData(models.SeasonParams{Slug: &selectedAnimeSlug})
		utils.FailIfErr(err, logger)
		isMovie = *seasonData[0].IsMovie
	}
	if !isMovie {
		episodes, err = source.GetEpisodesData(models.EpisodeParams{SeasonID: &selectedAnimeID, Slug: &selectedAnimeSlug})
		utils.FailIfErr(err, logger)
		for _, e := range episodes {
			episodeNames = append(episodeNames, e.Title)
		}
		selectedSeasonIndex = int(episodes[0].Extra["season_num"].(float64)) - 1
	} else {
		episodeNames = []string{selectedAnimeName}
		episodes = []models.Episode{
			{
				Title: selectedAnimeName,
				Extra: map[string]interface{}{
					"season_num": float64(1),
				},
			},
		}
		selectedSeasonIndex = 0
	}
	return episodes, episodeNames, isMovie, selectedSeasonIndex
}

func playAnimeLoop(
	source models.AnimeSource,
	selectedSource string,
	episodes []models.Episode,
	episodeNames []string,
	selectedAnimeID int,
	selectedAnimeSlug string,
	selectedAnimeName string,
	isMovie bool,
	selectedSeasonIndex int,
	uiMode string,
	rofiFlags string,
	posterUrl string,
	disableRpc bool,
	logger *utils.Logger,
) (models.AnimeSource, string) {
	selectedEpisodeIndex := 0
	selectedFansubIdx := 0
	selectedResolution := ""
	selectedResolutionIdx := 0
	loggedIn, err := rpc.ClientLogin()
	if err != nil || !loggedIn {
		logger.LogError(err)
	}
	for {
		ui.ClearScreen()
		watchMenu := []string{}
		if !isMovie {
			watchMenu = append(watchMenu, "İzle", "Sonraki bölüm", "Önceki bölüm", "Bölüm seç", "Çözünürlük seç")
		} else {
			watchMenu = append(watchMenu, "İzle", "Çözünürlük seç")
		}

		if strings.ToLower(selectedSource) == "openanime" {
			watchMenu = append(watchMenu, "Fansub seç")
		}

		watchMenu = append(watchMenu, "Anime ara", "Çık")

		option, err := ui.SelectionList(internal.UiParams{
			Mode:      uiMode,
			RofiFlags: &rofiFlags,
			List:      &watchMenu,
			Label:     selectedAnimeName,
		})
		utils.FailIfErr(err, logger)
		switch option {
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
			selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1
			data, _, err := updateWatchApi(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if !utils.CheckErr(err, logger) {
				continue
			}
			labels := data["labels"].([]string)
			urls := data["urls"].([]string)
			subtitle := data["caption_url"].(string)
			if selectedResolution == "" {
				selectedResolutionIdx = 0
				if len(labels) > 0 {
					selectedResolution = labels[selectedResolutionIdx]
				}
			}
			if selectedResolutionIdx >= len(urls) {
				selectedResolutionIdx = len(urls) - 1
			}

			cmd, socketPath, err := player.Play(player.MPVParams{
				Url:         urls[selectedResolutionIdx],
				SubtitleUrl: &subtitle,
				Title:       fmt.Sprintf("%s - %s", selectedAnimeName, episodeNames[selectedEpisodeIndex]),
			})
			if !utils.CheckErr(err, logger) {
				return source, selectedSource
			}
			maxAttempts := 10
			mpvRunning := false
			for i := 0; i < maxAttempts; i++ {
				time.Sleep(300 * time.Millisecond)
				if player.IsMPVRunning(socketPath) {
					mpvRunning = true
					break
				}
			}
			if !mpvRunning {
				logger.LogError(errors.New("MPV başlatılamadı veya zamanında yanıt vermedi"))
				return source, selectedSource
			}
			if !disableRpc {
				go updateDiscordRPC(socketPath, episodeNames, selectedEpisodeIndex, selectedAnimeName, selectedSource, posterUrl, logger, &loggedIn)
			}
			err = cmd.Wait()
			if err != nil {
				fmt.Println("MPV çalışırken hata:", err)
			}
		case "Çözünürlük seç":
			data, _, err := updateWatchApi(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if !utils.CheckErr(err, logger) {
				continue
			}
			labels := data["labels"].([]string)
			selected, err := ui.SelectionList(internal.UiParams{
				Mode:      uiMode,
				RofiFlags: &rofiFlags,
				List:      &labels,
				Label:     "Çözünürlük seç ",
			})
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
		case "Bölüm seç":
			selected, err := ui.SelectionList(internal.UiParams{
				Mode:      uiMode,
				RofiFlags: &rofiFlags,
				List:      &episodeNames,
				Label:     "Bölüm seç ",
			})
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
		case "Fansub seç":
			fansubNames := []string{}

			if strings.ToLower(source.Source()) != "openanime" {
				fmt.Println("\033[31m[!] Bu seçenek sadece OpenAnime için geçerlidir.\033[0m")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			_, fansubData, err := updateWatchApi(
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
				logger.LogError(fmt.Errorf("updateWatchApi verisi alınamadı: %w", err))
				continue
			}

			for _, fansub := range fansubData {
				if fansub.Name != nil {
					fansubNames = append(fansubNames, *fansub.Name)
				}
			}

			selected, err := ui.SelectionList(internal.UiParams{
				Mode:      uiMode,
				RofiFlags: &rofiFlags,
				List:      &fansubNames,
				Label:     "Fansub seç ",
			})
			if !utils.CheckErr(err, logger) {
				continue
			}

			if !slices.Contains(fansubNames, selected) {
				fmt.Printf("\033[31m[!] Geçersiz fansub seçimi: %s\033[0m\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedFansubIdx = slices.Index(fansubNames, selected)

		case "Anime ara":
			for {
				choice, err := ui.SelectionList(internal.UiParams{
					Mode:      uiMode,
					RofiFlags: &rofiFlags,
					List:      &[]string{"Bu kaynakla devam et", "Kaynak değiştir", "Çık"},
					Label:     fmt.Sprintf("Arama kaynağı: %s", selectedSource),
				})

				if err != nil {
					logger.LogError(fmt.Errorf("seçim listesi oluşturulamadı: %w", err))
					continue
				}

				switch choice {
				case "Bu kaynakla devam et":
					// Devam et, hiçbir şey yapma
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
		case "Çık":
			os.Exit(0)
		default:
			return source, selectedSource
		}
	}
}

func updateDiscordRPC(socketPath string, episodeNames []string, selectedEpisodeIndex int, selectedAnimeName, selectedSource, posterUrl string, logger *utils.Logger, loggedIn *bool) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for range ticker.C {
		if !player.IsMPVRunning(socketPath) {
			break
		}
		isPaused, err := player.GetMPVPausedStatus(socketPath)
		if err != nil {
			logger.LogError(fmt.Errorf("pause durumu alınamadı: %w", err))
			continue
		}
		durationVal, err := player.MPVSendCommand(socketPath, []interface{}{"get_property", "duration"})
		if err != nil {
			logger.LogError(fmt.Errorf("süre alınamadı: %w", err))
			continue
		}
		timePosVal, err := player.MPVSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if err != nil {
			logger.LogError(fmt.Errorf("konum alınamadı: %w", err))
			continue
		}
		duration, ok1 := durationVal.(float64)
		timePos, ok2 := timePosVal.(float64)
		if !ok1 || !ok2 {
			logger.LogError(fmt.Errorf("süre veya zaman konumu parse edilemedi"))
			continue
		}
		formatTime := func(seconds float64) string {
			total := int(seconds + 0.5)
			hours := total / 3600
			minutes := (total % 3600) / 60
			secs := total % 60

			if hours > 0 {
				return fmt.Sprintf("%02d:%02d:%02d", hours, minutes, secs)
			}
			return fmt.Sprintf("%02d:%02d", minutes, secs)
		}

		state := fmt.Sprintf("%s (%s / %s)", episodeNames[selectedEpisodeIndex], formatTime(timePos), formatTime(duration))
		if isPaused {
			state = fmt.Sprintf("%s (%s / %s) (Paused)", episodeNames[selectedEpisodeIndex], formatTime(timePos), formatTime(duration))
		}
		var err2 error
		*loggedIn, err2 = rpc.DiscordRPC(internal.RPCParams{
			Details:    selectedAnimeName,
			State:      state,
			SmallImage: strings.ToLower(selectedSource),
			SmallText:  selectedSource,
			LargeImage: posterUrl,
			LargeText:  selectedAnimeName,
		}, *loggedIn)
		if err2 != nil {
			logger.LogError(fmt.Errorf("DiscordRPC hatası: %w", err2))
			continue
		}
	}
}

// --- Uygulama ana akışı ---

func runApp() {
	logger, err := utils.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	log.SetFlags(0)
	uiMode := "tui"
	rootCmd, f := flags.NewFlagsCmd()
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		disableRpc := f.DisableRPC
		printVersion := f.PrintVersion
		rofiMode := f.RofiMode
		rofiFlags := f.RofiFlags
		if printVersion {
			update.Version()
			return
		}
		if rofiMode {
			uiMode = "rofi"
		}
		update.CheckUpdates()
		selectedSource, source := selectSource(uiMode, rofiFlags, logger)

		for {
			searchData, animeNames, animeTypes, _ := searchAnime(source, uiMode, rofiFlags, logger)
			isMovie := false
			selectedAnime, isMovie, _ := selectAnime(animeNames, searchData, uiMode, isMovie, rofiFlags, animeTypes, logger)
			posterUrl := selectedAnime.ImageURL
			if !utils.IsValidImage(posterUrl) {
				posterUrl = "anitrcli"
			}
			selectedAnimeID, selectedAnimeSlug := getAnimeIDs(source, selectedAnime)
			episodes, episodeNames, isMovie, selectedSeasonIndex := getEpisodesAndNames(source, isMovie, selectedAnimeID, selectedAnimeSlug, selectedAnime.Title, logger)
			source, selectedSource = playAnimeLoop(source, selectedSource, episodes, episodeNames, selectedAnimeID, selectedAnimeSlug, selectedAnime.Title, isMovie, selectedSeasonIndex, uiMode, rofiFlags, posterUrl, disableRpc, logger)
		}
	}
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	runApp()
}
