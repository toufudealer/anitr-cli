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
	"github.com/xeyossr/anitr-cli/internal/downloader"
	"github.com/xeyossr/anitr-cli/internal/flags"
	"github.com/xeyossr/anitr-cli/internal/models"
	"github.com/xeyossr/anitr-cli/internal/player"
	"github.com/xeyossr/anitr-cli/internal/rpc"
	"github.com/xeyossr/anitr-cli/internal/sources/animecix"
	"github.com/xeyossr/anitr-cli/internal/sources/openanime"
	"github.com/xeyossr/anitr-cli/internal/ui"
	"github.com/xeyossr/anitr-cli/internal/utils"
)

func updateWatchAPI(
	source string,
	episodeData []models.Episode,
	index, id, seasonIndex, selectedFansubIndex int,
	isMovie bool,
	slug *string,
) (map[string]interface{}, []models.Fansub, error) {
	var (
		captionData []map[string]string
		fansubData  []models.Fansub
		captionURL  string
		err         error
	)

	switch source {
	case "animecix":
		if isMovie {
			data, err := animecix.AnimeMovieWatchApiUrl(id)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix movie API çağrısı başarısız: %w", err)
			}
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
			if index < 0 || index >= len(episodeData) {
				return nil, nil, fmt.Errorf("index out of range")
			}
			urlData := episodeData[index].ID
			captionData, err = animecix.AnimeWatchApiUrl(urlData)
			if err != nil {
				return nil, nil, fmt.Errorf("animecix watch API çağrısı başarısız: %w", err)
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
		"caption_url": captionURL,
	}, fansubData, nil
}

func selectSource(uiMode string, rofiFlags string, logger *utils.Logger) (string, models.AnimeSource) {
	for {
		sourceList := []string{"OpenAnime", "AnimeciX"}

		appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
		selectedSourceSlice, err := showSelection(appCtx, sourceList, "Kaynak seç ", "generic", nil)
		if err != nil || len(selectedSourceSlice) == 0 {
			utils.FailIfErr(err, logger)
			continue
		}
		selectedSource := selectedSourceSlice[0]
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
		query, err := ui.InputFromUser(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Anime ara ", Logger: logger})
		utils.FailIfErr(err, logger)

		searchData, err := source.GetSearchData(query)
		utils.FailIfErr(err, logger)

		if searchData == nil {
			fmt.Printf("[!] Arama sonucu bulunamadı!\n")
			time.Sleep(1500 * time.Millisecond)
			continue
		}

	animeNames := make([]string, 0, len(searchData))
	animeTypes := make([]string, 0, len(searchData))
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

		appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
		selectedAnimeNameSlice, err := showSelection(appCtx, animeNames, "Anime seç ", "", nil)
		utils.FailIfErr(err, logger)

		if len(selectedAnimeNameSlice) == 0 {
			continue
		}
		selectedAnimeName := selectedAnimeNameSlice[0]

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

	if source.Source() == "animecix" && selectedAnime.ID != nil {
		selectedAnimeID = *selectedAnime.ID
	} else if source.Source() == "openanime" && selectedAnime.Slug != nil {
		selectedAnimeSlug = *selectedAnime.Slug
	}
	return selectedAnimeID, selectedAnimeSlug
}

func getEpisodesAndNames(source models.AnimeSource, isMovie bool, selectedAnimeID int, selectedAnimeSlug string, selectedAnimeName string, logger *utils.Logger) ([]models.Episode, []string, bool, int, error) {
	var (
		episodes            []models.Episode
		episodeNames        []string
		selectedSeasonIndex int
		err                 error
	)

	if source.Source() == "openanime" {
		seasonData, err := source.GetSeasonsData(models.SeasonParams{Slug: &selectedAnimeSlug})
		if err != nil {
			logger.LogError(err)
			return nil, nil, false, 0, fmt.Errorf("sezon verisi alınamadı: %w", err)
		}
		isMovie = *seasonData[0].IsMovie
	}

	if !isMovie {
		episodes, err = source.GetEpisodesData(models.EpisodeParams{SeasonID: &selectedAnimeID, Slug: &selectedAnimeSlug})
		if err != nil {
			return nil, nil, false, 0, fmt.Errorf("bölüm verisi alınamadı: %w", err)
		}

		if len(episodes) == 0 {
			return nil, nil, false, 0, fmt.Errorf("hiçbir bölüm bulunamadı")
		}

		episodeNames = make([]string, 0, len(episodes))
		for _, e := range episodes {
			episodeNames = append(episodeNames, e.Title)
		}

		selectedSeasonIndex = int(episodes[0].Extra["season_num"].(float64)) - 1
	} else {
		episodeNames = []string{selectedAnimeName}
		episodes = []models.Episode{{
			Title: selectedAnimeName,
			Extra: map[string]interface{}{"season_num": float64(1)},
		}}
		selectedSeasonIndex = 0
	}

	return episodes, episodeNames, isMovie, selectedSeasonIndex, nil
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
	posterURL string,
	disableRPC bool,
	logger *utils.Logger,
) (models.AnimeSource, string, bool) {

	selectedEpisodeIndex := 0
	selectedFansubIdx := 0
	selectedResolution := ""
	selectedResolutionIdx := 0

	loggedIn, err := rpc.ClientLogin()
	if err != nil || !loggedIn {
		logger.LogError(err)
	}

	for {
		watchMenu := []string{}
		if !isMovie {
			watchMenu = append(watchMenu, "İzle", "Sonraki bölüm", "Önceki bölüm", "Bölüm seç", "Çözünürlük seç", "İndir", "Toplu İndir")
		} else {
			watchMenu = append(watchMenu, "İzle", "Çözünürlük seç", "İndir")
		}

		if strings.ToLower(selectedSource) == "openanime" {
			watchMenu = append(watchMenu, "Fansub seç")
		}

		watchMenu = append(watchMenu, "Geri", "Anime ara", "Çık")

		appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
		optionSlice, err := showSelection(appCtx, watchMenu, selectedAnimeName, "", nil)
		utils.FailIfErr(err, logger)

		if len(optionSlice) == 0 {
			return source, selectedSource, true
		}
		option := optionSlice[0]

		switch option {
		case "Geri":
			return source, selectedSource, true
		case "İzle", "Sonraki bölüm", "Önceki bölüm":
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
				fmt.Printf("[!] Bölüm oynatılamadı: %s\n", err)
				time.Sleep(1500 * time.Millisecond)
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

			mpvTitle := fmt.Sprintf("%s - %s", selectedAnimeName, episodeNames[selectedEpisodeIndex])
			if isMovie {
				mpvTitle = selectedAnimeName
			}

			cmd, _, err := player.PlayVLC(player.VLCParams{
				Url:         urls[selectedResolutionIdx],
				SubtitleUrl: &subtitle,
				Title:       mpvTitle,
				VLCPath:     "",
			})
			if !utils.CheckErr(err, logger) {
				return source, selectedSource, false
			}

			if !disableRPC {
				go updateDiscordRPC(episodeNames, selectedEpisodeIndex, selectedAnimeName, selectedSource, posterURL, logger, &loggedIn)
			}

			err = cmd.Wait()
			if err != nil {
				fmt.Println("VLC çalışırken hata:", err)
			}

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
				fmt.Printf("[!] Çözünürlükler yüklenemedi.\n")
				time.Sleep(1000 * time.Millisecond)
				continue
			}
			labels := data["labels"].([]string)
			appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
			selectedSlice, err := showSelection(appCtx, labels, "Çözünürlük seç ", "", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}
			if len(selectedSlice) == 0 {
				continue
			}
			selected := selectedSlice[0]
			if !slices.Contains(labels, selected) {
				fmt.Printf("[!] Geçersiz çözünürlük seçimi: %s\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedResolutionIdx = slices.Index(labels, selected)

		case "Bölüm seç":
			appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
			selectedSlice, err := showSelection(appCtx, episodeNames, "Bölüm seç ", "", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}
			if len(selectedSlice) == 0 {
				continue
			}
			selected := selectedSlice[0]
			if !slices.Contains(episodeNames, selected) {
				fmt.Printf("[!] Geçersiz bölüm seçimi: %s\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedEpisodeIndex = slices.Index(episodeNames, selected)
			if !isMovie && selectedEpisodeIndex >= 0 && selectedEpisodeIndex < len(episodes) {
				selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1
			}

		case "Fansub seç":
			fansubNames := []string{}

			if strings.ToLower(source.Source()) != "openanime" {
				fmt.Println("[!] Bu seçenek sadece OpenAnime için geçerlidir.")
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
				fmt.Printf("[!] Fansublar yüklenemedi.\n")
				time.Sleep(1000 * time.Millisecond)
				continue
			}

			for _, fansub := range fansubData {
				if fansub.Name != nil {
					fansubNames = append(fansubNames, *fansub.Name)
				}
			}

			appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
			selectedSlice, err := showSelection(appCtx, fansubNames, "Fansub seç ", "", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}
			if len(selectedSlice) == 0 {
				continue
			}
			selected := selectedSlice[0]
			if !slices.Contains(fansubNames, selected) {
				fmt.Printf("[!] Geçersiz fansub seçimi: %s\n", selected)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			selectedFansubIdx = slices.Index(fansubNames, selected)

		case "İndir":
			downloadEpisodeIndex := selectedEpisodeIndex
			if !isMovie {
				appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
				selectedSlice, err := showSelection(appCtx, episodeNames, "İndirilecek bölümü seç ", "", nil)
				if !utils.CheckErr(err, logger) {
					continue
				}
				if len(selectedSlice) == 0 {
					continue
				}
				selected := selectedSlice[0]
				downloadEpisodeIndex = slices.Index(episodeNames, selected)
			} else {
				// Filmler için sadece 0. indeks kullan
				downloadEpisodeIndex = 0
			}

			data, _, err := updateWatchAPI(
				strings.ToLower(selectedSource),
				episodes,
				downloadEpisodeIndex,
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if err != nil {
				fmt.Printf("[!] İndirme bağlantıları yüklenemedi: %s\n", err)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			labels := data["labels"].([]string)
			urls := data["urls"].([]string)

			if len(urls) == 0 {
				fmt.Println("Bu bölüm için indirme bağlantısı bulunamadı.")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
			selectedResolutionLabelSlice, err := showSelection(appCtx, labels, "İndirilecek çözünürlüğü seç ", "", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}
			if len(selectedResolutionLabelSlice) == 0 {
				continue
			}
			selectedResolutionLabel := selectedResolutionLabelSlice[0]

			selectedDownloadIdx := -1
			for i, label := range labels {
				if label == selectedResolutionLabel {
					selectedDownloadIdx = i
					break
				}
			}

			if selectedDownloadIdx == -1 {
				fmt.Printf("[!] Geçersiz çözünürlük seçimi: %s\n", selectedResolutionLabel)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			downloadURL := urls[selectedDownloadIdx]
			fmt.Printf("İndirme URL'si bulundu: %s\n", downloadURL)

			filename := fmt.Sprintf("%s - E%02d.mp4", selectedAnimeName, episodes[downloadEpisodeIndex].Number)
			downloadPath := filename

			fmt.Printf("İndiriliyor: %s\n", downloadPath)
			err = downloader.DownloadFile(downloadURL, downloadPath)
			if err != nil {
				fmt.Printf("Dosya indirilirken hata: %v\n", err)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			fmt.Println("İndirme tamamlandı!")
			time.Sleep(1500 * time.Millisecond)

		case "Toplu İndir":
			if isMovie {
				fmt.Println("[!] Film için toplu indirme yapılamaz.")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
			selectedEpisodeTitles, err := showSelection(appCtx, episodeNames, "İndirilecek bölümleri seç (Space ile işaretle, Enter ile onayla)", "multi-select", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}

			if len(selectedEpisodeTitles) == 0 {
				fmt.Println("[!] İndirilecek bölüm seçilmedi.")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			// Convert selected episode titles back to indices
			epsToDownload := []int{}
			for _, title := range selectedEpisodeTitles {
				idx := slices.Index(episodeNames, title)
				if idx != -1 {
					epsToDownload = append(epsToDownload, idx)
				}
			}
			sort.Ints(epsToDownload) // Ensure sorted order

			// Ask for resolution once for all batch downloads
			data, _, err := updateWatchAPI(
				strings.ToLower(selectedSource),
				episodes,
				selectedEpisodeIndex, // Use current index to get resolutions, not specific episode
				selectedAnimeID,
				selectedSeasonIndex,
				selectedFansubIdx,
				isMovie,
				&selectedAnimeSlug,
			)
			if err != nil {
				fmt.Printf("[!] Çözünürlükler yüklenemedi: %s\n", err)
				time.Sleep(1500 * time.Millisecond)
				continue
			}
			labels := data["labels"].([]string)
			urls := data["urls"].([]string)

			if len(urls) == 0 {
				fmt.Println("Bu anime için indirme bağlantısı bulunamadı.")
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			selectedResolutionLabelsSlice, err := showSelection(appCtx, labels, "Tüm bölümler için çözünürlüğü seç ", "", nil)
			if !utils.CheckErr(err, logger) {
				continue
			}
			if len(selectedResolutionLabelsSlice) == 0 {
				continue
			}
			selectedResolutionLabel := selectedResolutionLabelsSlice[0]

			batchDownloadIdx := -1
			for i, label := range labels {
				if label == selectedResolutionLabel {
					batchDownloadIdx = i
					break
				}
			}

			if batchDownloadIdx == -1 {
				fmt.Printf("[!] Geçersiz çözünürlük seçimi: %s\n", selectedResolutionLabel)
				time.Sleep(1500 * time.Millisecond)
				continue
			}

			for _, epIdx := range epsToDownload {
				episode := episodes[epIdx]
				fmt.Printf("İndiriliyor: %s - E%02d...\n", selectedAnimeName, episode.Number)

				// Re-fetch watch data for each episode to get correct URLs for that episode
				currentEpisodeWatchData, _, err := updateWatchAPI(
					strings.ToLower(selectedSource),
					episodes,
					epIdx,
					selectedAnimeID,
					int(episode.Extra["season_num"].(float64)) - 1,
					selectedFansubIdx,
				isMovie,
					&selectedAnimeSlug,
				)
				if err != nil {
					fmt.Printf("[!] Bölüm %d için indirme bağlantıları yüklenemedi: %s\n", episode.Number, err)
					continue // Skip this episode and try next
				}

				currentEpisodeUrls := currentEpisodeWatchData["urls"].([]string)
				currentEpisodeLabels := currentEpisodeWatchData["labels"].([]string)

				// Find the URL corresponding to the selected resolution for the current episode
				currentDownloadURL := ""
				for i, label := range currentEpisodeLabels {
					if label == selectedResolutionLabel {
						currentDownloadURL = currentEpisodeUrls[i]
						break
					}
				}

				if currentDownloadURL == "" {
					fmt.Printf("[!] Bölüm %d için seçilen çözünürlükte indirme bağlantısı bulunamadı.\n", episode.Number)
					continue
				}

				filename := fmt.Sprintf("%s - E%02d.mp4", selectedAnimeName, episode.Number)
				downloadPath := filename

				err = downloader.DownloadFile(currentDownloadURL, downloadPath)
				if err != nil {
					fmt.Printf("[!] Dosya indirilirken hata (%s - E%02d): %v\n", selectedAnimeName, episode.Number, err)
				} else {
					fmt.Printf("İndirme tamamlandı: %s - E%02d\n", selectedAnimeName, episode.Number)
				}
				time.Sleep(500 * time.Millisecond) // Small delay between downloads
			}
			fmt.Println("Tüm toplu indirmeler tamamlandı!")
			time.Sleep(1500 * time.Millisecond)

		case "Anime ara":
			for {
				appCtx := App{uiMode: &uiMode, rofiFlags: &rofiFlags, logger: logger}
				choices, err := showSelection(appCtx, []string{"Bu kaynakla devam et", "Kaynak değiştir", "Çık"}, fmt.Sprintf("Arama kaynağı: %s", selectedSource), "", nil)
				if !utils.CheckErr(err, logger) {
					continue
				}
				if len(choices) == 0 { // User cancelled
					os.Exit(0)
				}
				choice := choices[0] // Get the single selected choice

				switch choice {
				case "Bu kaynakla devam et":
				case "Kaynak değiştir":
					selectedSource, source = selectSource(uiMode, rofiFlags, logger)
				case "Çık":
					os.Exit(0)
				default:
					fmt.Printf("[!] Geçersiz seçim: %s\n", choice)
					time.Sleep(1500 * time.Millisecond)
					continue
				}

				return source, selectedSource, false
			}

		case "Çık":
			os.Exit(0)

		default:
			return source, selectedSource, false
		}
	}
}

func updateDiscordRPC(episodeNames []string, selectedEpisodeIndex int, selectedAnimeName, selectedSource, posterURL string, logger *utils.Logger, loggedIn *bool) {
	if !*loggedIn {
		var err error
		*loggedIn, err = rpc.ClientLogin()
		if err != nil {
			logger.LogError(err)
			return
		}
	}

	state := fmt.Sprintf("Bölüm %d: %s", selectedEpisodeIndex+1, episodeNames[selectedEpisodeIndex])

	var err2 error
	*loggedIn, err2 = rpc.DiscordRPC(internal.RPCParams{
		Details:    selectedAnimeName,
		State:      state,
		SmallImage: strings.ToLower(selectedSource),
		SmallText:  selectedSource,
		LargeImage: posterURL,
		LargeText:  selectedAnimeName,
	}, *loggedIn)

	if err2 != nil {
		logger.LogError(fmt.Errorf("DiscordRPC hatası: %w", err2))
	}
}

type App struct {
	source         *models.AnimeSource
	selectedSource *string
	uiMode         *string
	rofiFlags      *string
	disableRPC     *bool
	logger         *utils.Logger
}

func showSelection(cfx App, list []string, label string, promptType string, data interface{}) ([]string, error) {
	response, err := ui.SelectionList(internal.UiParams{
		Mode:      *cfx.uiMode,
		RofiFlags: cfx.rofiFlags,
		List:      &list,
		Label:     label,
		Type:      promptType,
		Data:      data,
		Logger:    cfx.logger,
	})
	if err != nil {
		return nil, err
	}
	return response, nil
}

func app(cfx *App) error {
	for {
		searchData, animeNames, animeTypes, _ := searchAnime(*cfx.source, *cfx.uiMode, *cfx.rofiFlags, cfx.logger)
		isMovie := false
		selectedAnime, isMovie, _ := selectAnime(animeNames, searchData, *cfx.uiMode, isMovie, *cfx.rofiFlags, animeTypes, cfx.logger)

		stayInActionMenu := true
		for stayInActionMenu {
			actionMenu := []string{"Bölümleri Listele", "Anime Ara", "Kaynak Değiştir", "Çık"}
			selectedActionSlice, err := showSelection(*cfx, actionMenu, selectedAnime.Title, "", nil)
			if err != nil {
				cfx.logger.LogError(fmt.Errorf("aksiyon menüsü hatası: %w", err))
				stayInActionMenu = false
				continue
			}

			if len(selectedActionSlice) == 0 {
				stayInActionMenu = false
				continue
			}
			selectedAction := selectedActionSlice[0]

			switch selectedAction {
			case "Bölümleri Listele":
				posterURL := selectedAnime.ImageURL
				if !utils.IsValidImage(posterURL) {
					posterURL = "anitrcli"
				}

				selectedAnimeID, selectedAnimeSlug := getAnimeIDs(*cfx.source, selectedAnime)

				episodes, episodeNames, isMovie, selectedSeasonIndex, err := getEpisodesAndNames(
					*cfx.source, isMovie, selectedAnimeID, selectedAnimeSlug, selectedAnime.Title, cfx.logger,
				)

				if err != nil {
					cfx.logger.LogError(err)

					choices, err := showSelection(*cfx, []string{"Farklı Anime Ara", "Kaynak Değiştir", "Çık"}, fmt.Sprintf("Hata: %s", err.Error()), "", nil)
					if !utils.CheckErr(err, cfx.logger) {
						return err
					}
					if len(choices) == 0 {
						os.Exit(0)
					}
					choice := choices[0]

					switch choice {
					case "Farklı Anime Ara":
						stayInActionMenu = false
					case "Kaynak Değiştir":
						selectedSource, source := selectSource(*cfx.uiMode, *cfx.rofiFlags, cfx.logger)
						cfx.selectedSource = utils.Ptr(selectedSource)
						cfx.source = utils.Ptr(source)
						stayInActionMenu = false
					default:
						os.Exit(0)
					}
					continue
				}

				newSource, newSelectedSource, backPressed := playAnimeLoop(
					*cfx.source, *cfx.selectedSource, episodes, episodeNames,
					selectedAnimeID, selectedAnimeSlug, selectedAnime.Title,
					isMovie, selectedSeasonIndex, *cfx.uiMode, *cfx.rofiFlags,
					posterURL, *cfx.disableRPC, cfx.logger,
				)

				if newSource != *cfx.source || newSelectedSource != *cfx.selectedSource {
					cfx.source = &newSource
					cfx.selectedSource = &newSelectedSource
				}

				if !backPressed {
					stayInActionMenu = false
				}

			case "Anime Ara":
				stayInActionMenu = false

			case "Kaynak Değiştir":
				selectedSource, source := selectSource(*cfx.uiMode, *cfx.rofiFlags, cfx.logger)
				cfx.selectedSource = utils.Ptr(selectedSource)
				cfx.source = utils.Ptr(source)
				stayInActionMenu = false

			case "Çık":
				os.Exit(0)
			}
		}
	}
}

func runMain(f *flags.Flags, uiMode string, logger *utils.Logger) {
	disableRPC := f.DisableRPC

	currentApp := &App{
		source:         nil,
		selectedSource: utils.Ptr(""),
		uiMode:         &uiMode,
		rofiFlags:      &f.RofiFlags,
		disableRPC:     &disableRPC,
		logger:         logger,
	}

	for {
		if currentApp.source == nil {
			selectedSource, source := selectSource(uiMode, f.RofiFlags, logger)
			currentApp.selectedSource = utils.Ptr(selectedSource)
			currentApp.source = utils.Ptr(source)
		}

		if err := app(currentApp); err != nil {
			logger.LogError(err)
			currentApp.source = nil
		}
	}
}

var downloadCmd = &cobra.Command{
	Use:   "download [anime_title] [episode_number]",
	Short: "Downloads an anime episode",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		animeTitle := args[0]
		episodeNumberStr := args[1]
		episodeNumber, err := strconv.Atoi(episodeNumberStr)
		if err != nil {
			fmt.Printf("Error: Invalid episode number '%s'\n", episodeNumberStr)
			os.Exit(1)
		}

		logger, err := utils.NewLogger()
		if err != nil {
			panic(err)
		}
		defer logger.Close()

		fmt.Printf("Attempting to download %s episode %d...\n", animeTitle, episodeNumber)

		animeSource := animecix.AnimeCix{}
		searchData, err := animeSource.GetSearchData(animeTitle)
		if err != nil {
			fmt.Printf("Error searching for anime: %v\n", err)
			os.Exit(1)
		}
		if len(searchData) == 0 {
			fmt.Printf("No anime found for '%s'\n", animeTitle)
			os.Exit(1)
		}

		selectedAnime := searchData[0]
		selectedAnimeID := *selectedAnime.ID

		episodes, _, _, _, err := getEpisodesAndNames(
			animeSource,
			false,
			selectedAnimeID,
			"",
			selectedAnime.Title,
			logger,
		)
		if err != nil {
			fmt.Printf("Error getting episodes: %v\n", err)
			os.Exit(1)
		}

		var targetEpisode models.Episode
		foundEpisode := false
		for _, ep := range episodes {
			if ep.Number == episodeNumber {
				targetEpisode = ep
				foundEpisode = true
				break
			}
		}

		if !foundEpisode {
			fmt.Printf("Episode %d not found for %s\n", episodeNumber, animeTitle)
			os.Exit(1)
		}

		episodeIndex := -1
		for i, ep := range episodes {
			if ep.ID == targetEpisode.ID {
				episodeIndex = i
				break
			}
		}

		if episodeIndex == -1 {
			fmt.Printf("Could not find internal index for episode %d\n", episodeNumber)
			os.Exit(1)
		}

		seasonNum, ok := targetEpisode.Extra["season_num"].(float64)
		if !ok {
			fmt.Printf("Error: Could not determine season number for episode %d\n", episodeNumber)
			os.Exit(1)
		}
		seasonIndex := int(seasonNum) - 1

		watchData, _, err := updateWatchAPI(
			"animecix",
			episodes,
			episodeIndex,
			selectedAnimeID,
			seasonIndex,
			0,
			false,
			nil,
		)
		if err != nil {
			fmt.Printf("Error getting watch data: %v\n", err)
			os.Exit(1)
		}

		urls := watchData["urls"].([]string)
		if len(urls) == 0 {
			fmt.Println("No video URLs found for this episode.")
			os.Exit(1)
		}

		downloadURL := urls[0]
		fmt.Printf("Found download URL: %s\n", downloadURL)

		filename := fmt.Sprintf("%s - E%02d.mp4", selectedAnime.Title, episodeNumber)
		downloadPath := filename

		fmt.Printf("İndiriliyor: %s\n", downloadPath)
		err = downloader.DownloadFile(downloadURL, downloadPath)
		if err != nil {
			fmt.Printf("Error downloading file: %v\n", err)
			os.Exit(1)
		}

		fmt.Println("Download complete!")
	},
}

func runApp() {
	logger, err := utils.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Close()
	log.SetFlags(0)

	rootCmd, f := flags.NewFlagsCmd()

	rootCmd.AddCommand(downloadCmd)

	commands := rootCmd.Commands()

	if runtime.GOOS != "linux" {
		rootCmd.Run = func(cmd *cobra.Command, args []string) {
			f.RofiMode = false
			runMain(f, "tui", logger)
		}
	} else {
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
	runApp()
}