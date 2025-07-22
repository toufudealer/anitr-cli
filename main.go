package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
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

func updateWatchApi(
	source string,
	episodeData []models.Episode,
	index, id, seasonIndex, episodeIndex int,
	isMovie bool,
	slug *string,
) (map[string]interface{}, error) {
	var (
		captionData []map[string]string
		captionUrl  string
		err         error
	)

	switch source {
	case "animecix":
		if isMovie {
			// Movie için AnimeCix API çağrısı
			data, err := animecix.AnimeMovieWatchApiUrl(id)
			if err != nil {
				return nil, err
			}

			captionUrlIface := data["caption_url"]
			captionUrl, _ = captionUrlIface.(string)

			streamsIface, ok := data["video_streams"]
			if !ok {
				return nil, fmt.Errorf("video_streams not found")
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
				return nil, fmt.Errorf("index out of range")
			}
			urlData := episodeData[index].ID // models.Episode.ID string
			captionData, err = animecix.AnimeWatchApiUrl(urlData)
			if err != nil {
				return nil, err
			}

			// seasonEpisodeIndex hesaplama
			seasonEpisodeIndex := 0
			for i := 0; i < index; i++ {
				// episodeData[i].Extra["season_num"] interface{} tipinde, int e dönüştür
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
			return nil, fmt.Errorf("slug is required for openanime source")
		}

		if isMovie {
			// OpenAnime'da movie senaryosu yok ya da farklı olabilir,
			// burada bir hata veya farklı işlemi implement etmek gerekebilir.
			return nil, fmt.Errorf("movie not supported for openanime source")
		} else {
			// OpenAnime için episode bilgisi
			if index < 0 || index >= len(episodeData) {
				return nil, fmt.Errorf("index out of range")
			}

			ep := episodeData[index]

			// season_num ve episode_num almak
			seasonNum := 0
			episodeNum := 0

			if sn, ok := ep.Extra["season_num"].(int); ok {
				seasonNum = sn
			} else if snf, ok := ep.Extra["season_num"].(float64); ok {
				seasonNum = int(snf)
			} else {
				return nil, fmt.Errorf("season_num missing or invalid")
			}

			if en, ok := ep.Extra["episode_num"].(int); ok {
				episodeNum = en
			} else if enf, ok := ep.Extra["episode_num"].(float64); ok {
				episodeNum = int(enf)
			} else {
				// fallback episode number olarak ep.Number kullanabiliriz
				episodeNum = ep.Number
			}

			// OpenAnime'dan watch datayı almak için kendi GetWatchData kullanılabilir
			watchParams := models.WatchParams{
				Slug:    slug,
				Id:      &id,
				IsMovie: &isMovie,
				Extra:   &map[string]interface{}{"season_num": seasonNum, "episode_num": episodeNum},
			}

			watches, err := openanime.OpenAnime{}.GetWatchData(watchParams)
			if err != nil {
				return nil, err
			}

			if len(watches) < 1 {
				return nil, fmt.Errorf("no watch data found")
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
		}

	default:
		return nil, fmt.Errorf("unknown source: %s", source)
	}

	// captionData sıralama (label sayısal olarak büyükten küçüğe)
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
	}, nil
}

func main() {
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
		checkUpdate := f.CheckUpdate
		rofiMode := f.RofiMode
		rofiFlags := f.RofiFlags

		if printVersion {
			update.Version()
			return
		}

		if checkUpdate {
			err := update.RunUpdate()
			utils.FailIfErr(err, logger)
			return
		}

		if rofiMode {
			uiMode = "rofi"
		}

		update.CheckUpdates()

		ui.ClearScreen()
		sourceList := []string{"AnimeciX", "OpenAnime"}
		selectedSource, err := ui.SelectionList(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Kaynak seç ", List: &sourceList})
		utils.FailIfErr(err, logger)

		var source models.AnimeSource

		switch strings.ToLower(selectedSource) {
		case "animecix":
			source = animecix.AnimeCix{}
		case "openanime":
			source = openanime.OpenAnime{}
		}

		if strings.ToLower(selectedSource) == "" {
			log.Fatal("\033[31m[!] Kaynak seçilmedi\033[0m")
		}

		query, err := ui.InputFromUser(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, Label: "Anime ara "})
		utils.FailIfErr(err, logger)

		searchData, err := source.GetSearchData(query)
		utils.FailIfErr(err, logger)
		if searchData == nil {
			log.Fatal("\033[31m[!] Arama sonucu bulunamadı!\033[0m")
		}

		animeNames := []string{}
		animeTypes := []string{}
		var id string

		for _, item := range searchData {
			if item.ID != nil {
				id = strconv.Itoa(*item.ID)
			} else if item.Slug != nil {
				id = *item.Slug
			}

			animeNames = append(animeNames, fmt.Sprintf("%s (ID: %s)", item.Title, id))

			if item.TitleType != nil {
				ttype := item.TitleType
				if strings.ToLower(*ttype) == "movie" {
					animeTypes = append(animeTypes, "movie")
				} else {
					animeTypes = append(animeTypes, "tv")
				}
			}
		}

		ui.ClearScreen()
		selectedAnimeName, err := ui.SelectionList(internal.UiParams{Mode: uiMode, RofiFlags: &rofiFlags, List: &animeNames, Label: "Anime seç "})
		utils.FailIfErr(err, logger)
		if selectedAnimeName == "" {
			return
		}

		selectedIndex := slices.Index(animeNames, selectedAnimeName)
		selectedAnime := searchData[selectedIndex]

		var isMovie bool

		if len(animeTypes) > 0 {
			selectedAnimeType := animeTypes[selectedIndex]
			isMovie = selectedAnimeType == "movie"
		}

		posterUrl := selectedAnime.ImageURL
		if !utils.IsValidImage(posterUrl) {
			posterUrl = "anitrcli"
		}

		re := regexp.MustCompile(`^(.+?) \(ID: ([a-zA-Z0-9\-]+)\)$`)
		match := re.FindStringSubmatch(selectedAnimeName)
		if len(match) < 3 {
			log.Fatal("ID eşleşmedi")
		}
		selectedAnimeName = match[1]
		var (
			selectedAnimeID   int
			selectedAnimeSlug string
		)

		if source.Source() == "animecix" {
			selectedAnimeID, _ = strconv.Atoi(match[2])
		} else if source.Source() == "openanime" {
			selectedAnimeSlug = match[2]
		}

		var (
			episodes              []models.Episode
			episodeNames          []string
			selectedEpisodeIndex  int
			selectedResolution    string
			selectedResolutionIdx int
			selectedSeasonIndex   int
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
			selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1
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

		loggedIn, err := rpc.ClientLogin()
		if err != nil || !loggedIn {
			logger.LogError(err)
		}

		for {
			ui.ClearScreen()
			watchMenu := []string{"İzle", "Çözünürlük seç", "Çık"}
			if !isMovie {
				watchMenu = append([]string{"Sonraki bölüm", "Önceki bölüm", "Bölüm seç"}, watchMenu...)
			}

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

				// Sezonu her seferinde güncelle
				selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1

				data, err := updateWatchApi(
					strings.ToLower(selectedSource),
					episodes,
					selectedEpisodeIndex,
					selectedAnimeID,
					selectedSeasonIndex,
					selectedEpisodeIndex,
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
					return
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
					return
				}

				// 🎬 Rich Presence Güncelleme
				if !disableRpc {
					go func() {
						ticker := time.NewTicker(5 * time.Second) // Update interval: 5 saniye
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
								logger.LogError(errors.New("süre veya zaman konumu parse edilemedi"))
								continue
							}

							formatTime := func(seconds float64) string {
								total := int(seconds + 0.5)
								return fmt.Sprintf("%02d:%02d", total/60, total%60)
							}

							state := fmt.Sprintf("%s (%s / %s)",
								episodeNames[selectedEpisodeIndex],
								formatTime(timePos),
								formatTime(duration),
							)

							if isPaused {
								state = fmt.Sprintf("%s (%s / %s) (Paused)",
									episodeNames[selectedEpisodeIndex],
									formatTime(timePos),
									formatTime(duration),
								)
							}

							// Discord RPC güncelleme
							loggedIn, err = rpc.DiscordRPC(internal.RPCParams{
								Details:    selectedAnimeName,
								State:      state,
								SmallImage: strings.ToLower(selectedSource),
								SmallText:  selectedSource,
								LargeImage: posterUrl,
								LargeText:  selectedAnimeName,
							}, loggedIn)
							if err != nil {
								logger.LogError(fmt.Errorf("DiscordRPC hatası: %w", err))
								continue
							}
						}
					}()
				}

				err = cmd.Wait()
				if err != nil {
					fmt.Println("MPV çalışırken hata:", err)
				}

			case "Çözünürlük seç":
				data, err := updateWatchApi(
					strings.ToLower(selectedSource),
					episodes,
					selectedEpisodeIndex,
					selectedAnimeID,
					selectedSeasonIndex,
					selectedEpisodeIndex,
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

				if selected != "" {
					selectedEpisodeIndex = slices.Index(episodeNames, selected)

					if !isMovie && selectedEpisodeIndex >= 0 && selectedEpisodeIndex < len(episodes) {
						selectedSeasonIndex = int(episodes[selectedEpisodeIndex].Extra["season_num"].(float64)) - 1
					}
				}

			case "Çık":
				return
			default:
				return
			}
		}
	}

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
