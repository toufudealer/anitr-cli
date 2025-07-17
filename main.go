package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"

	"github.com/xeyossr/anitr-cli/api/animecix"
	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/player"
	"github.com/xeyossr/anitr-cli/internal/rpc"
	"github.com/xeyossr/anitr-cli/internal/ui"
	"github.com/xeyossr/anitr-cli/internal/update"
	"github.com/xeyossr/anitr-cli/internal/utils"
)

func FailIfErr(err error, logger *utils.Logger) {
	if err != nil {
		logger.LogError(err)
		log.Fatalf("\033[31mKritik hata: %v\033[0m", err)
	}
}

func checkErr(err error, logger *utils.Logger) bool {
	if err != nil {
		logger.LogError(err)
		fmt.Printf("\n\033[31mHata oluştu: %v\033[0m\nLog detayları: %s\nDevam etmek için bir tuşa basın...\n", err, logger.File.Name())
		fmt.Scanln()
		return false
	}
	return true
}

func isValidImage(url string) bool {
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

func updateWatchApi(episodeData []map[string]interface{}, index, id, seasonIndex, episodeIndex int, isMovie bool) (map[string]interface{}, error) {
	var (
		captionData []map[string]string
		captionUrl  string
		err         error
	)

	if isMovie {
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
		indexData := episodeData[index]
		urlData, _ := indexData["url"].(string)
		captionData, err = animecix.AnimeWatchApiUrl(urlData)
		if err != nil {
			return nil, err
		}

		seasonEpisodeIndex := 0
		for i := 0; i < index; i++ {
			if int(episodeData[i]["season_num"].(float64))-1 == seasonIndex {
				seasonEpisodeIndex++
			}
		}
		captionUrl, _ = animecix.FetchTRCaption(seasonIndex, seasonEpisodeIndex, id)

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
	}, nil
}

func main() {
	logger, err := utils.NewLogger()
	if err != nil {
		panic(err)
	}
	defer logger.Close()

	update.CheckUpdates()

	log.SetFlags(0)
	uiMode := "tui"

	disableRpc := flag.Bool("disable-rpc", false, "Discord Rich Presence özelliğini devre dışı bırakır.")
	checkUpdate := flag.Bool("update", false, "anitr-cli aracını en son sürüme günceller.")
	printVersion := flag.Bool("version", false, "versiyon")
	rofiMode := flag.Bool("rofi", false, "Rofi arayüzü ile başlatır.")
	rofiFlags := flag.String("rofi-flags", "", "Rofi için flag'ler")
	flag.Parse()

	if *printVersion {
		update.Version()
		return
	}

	if *checkUpdate {
		err := update.RunUpdate()
		FailIfErr(err, logger)
		return
	}

	if *rofiMode {
		uiMode = "rofi"
	}

	ui.ClearScreen()
	query, err := ui.InputFromUser(internal.UiParams{Mode: uiMode, RofiFlags: rofiFlags, Label: "Anime ara "})
	FailIfErr(err, logger)

	searchData, err := animecix.FetchAnimeSearchData(query)
	FailIfErr(err, logger)
	if searchData == nil {
		log.Fatal("\033[31m[!] Arama sonucu bulunamadı!\033[0m")
	}

	animeNames := []string{}
	animeTypes := []string{}
	for _, item := range searchData {
		id := int(item["id"].(float64))
		animeNames = append(animeNames, fmt.Sprintf("%s (ID: %d)", item["name"], id))

		ttype := internal.GetString(item, "title_type")
		if strings.ToLower(ttype) == "movie" {
			animeTypes = append(animeTypes, "movie")
		} else {
			animeTypes = append(animeTypes, "tv")
		}
	}

	selectedAnimeName, err := ui.SelectionList(internal.UiParams{Mode: uiMode, RofiFlags: rofiFlags, List: &animeNames, Label: "Anime seç "})
	FailIfErr(err, logger)
	if selectedAnimeName == "" {
		return
	}

	selectedIndex := slices.Index(animeNames, selectedAnimeName)
	selectedAnime := searchData[selectedIndex]
	selectedAnimeType := animeTypes[selectedIndex]
	isMovie := selectedAnimeType == "movie"

	posterUrl := internal.GetString(selectedAnime, "poster")
	if !isValidImage(posterUrl) {
		posterUrl = "anitrcli"
	}

	re := regexp.MustCompile(`^(.+?) \(ID: (\d+)\)$`)
	match := re.FindStringSubmatch(selectedAnimeName)
	if len(match) < 3 {
		log.Fatal("ID eşleşmedi")
	}
	selectedAnimeName = match[1]
	selectedAnimeID, _ := strconv.Atoi(match[2])

	var (
		episodes              []map[string]interface{}
		episodeNames          []string
		selectedEpisodeIndex  int
		selectedResolution    string
		selectedResolutionIdx int
		selectedSeasonIndex   int
	)

	if !isMovie {
		episodes, err = animecix.FetchAnimeEpisodesData(selectedAnimeID)
		FailIfErr(err, logger)
		for _, e := range episodes {
			episodeNames = append(episodeNames, internal.GetString(e, "name"))
		}
		selectedSeasonIndex = int(episodes[selectedEpisodeIndex]["season_num"].(float64)) - 1
	} else {
		episodeNames = []string{selectedAnimeName}
		episodes = []map[string]interface{}{
			{
				"name":       selectedAnimeName,
				"season_num": float64(1),
			},
		}
		selectedSeasonIndex = 0
	}

	for {
		ui.ClearScreen()
		watchMenu := []string{"İzle", "Çözünürlük seç", "Çık"}
		if !isMovie {
			watchMenu = append([]string{"Sonraki bölüm", "Önceki bölüm", "Bölüm seç"}, watchMenu...)
		}

		option, err := ui.SelectionList(internal.UiParams{
			Mode:      uiMode,
			RofiFlags: rofiFlags,
			List:      &watchMenu,
			Label:     selectedAnimeName,
		})
		FailIfErr(err, logger)

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
			selectedSeasonIndex = int(episodes[selectedEpisodeIndex]["season_num"].(float64)) - 1

			data, err := updateWatchApi(episodes, selectedEpisodeIndex, selectedAnimeID, selectedSeasonIndex, selectedEpisodeIndex, isMovie)
			if !checkErr(err, logger) {
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

			if !*disableRpc {
				state := selectedAnimeName
				if !isMovie {
					state = fmt.Sprintf("%s (%d/%d)", episodeNames[selectedEpisodeIndex], selectedEpisodeIndex+1, len(episodes))
				}

				if err := rpc.DiscordRPC(internal.RPCParams{
					Details:    selectedAnimeName,
					State:      state,
					LargeImage: posterUrl,
					LargeText:  selectedAnimeName,
				}); err != nil {
					logger.LogError(err)
				}
			}

			err = player.Play(urls[selectedResolutionIdx], &subtitle)
			if !checkErr(err, logger) {
				continue
			}

		case "Çözünürlük seç":
			data, err := updateWatchApi(episodes, selectedEpisodeIndex, selectedAnimeID, selectedSeasonIndex, selectedEpisodeIndex, isMovie)
			if !checkErr(err, logger) {
				continue
			}

			labels := data["labels"].([]string)
			selected, err := ui.SelectionList(internal.UiParams{
				Mode:      uiMode,
				RofiFlags: rofiFlags,
				List:      &labels,
				Label:     "Çözünürlük seç ",
			})
			if !checkErr(err, logger) {
				continue
			}

			selectedResolution = selected
			selectedResolutionIdx = slices.Index(labels, selected)

		case "Bölüm seç":
			selected, err := ui.SelectionList(internal.UiParams{
				Mode:      uiMode,
				RofiFlags: rofiFlags,
				List:      &episodeNames,
				Label:     "Bölüm seç ",
			})
			if !checkErr(err, logger) {
				continue
			}

			if selected != "" {
				selectedEpisodeIndex = slices.Index(episodeNames, selected)

				if !isMovie && selectedEpisodeIndex >= 0 && selectedEpisodeIndex < len(episodes) {
					selectedSeasonIndex = int(episodes[selectedEpisodeIndex]["season_num"].(float64)) - 1
				}
			}

			if !*disableRpc {
				totalEpisodes := len(episodes)
				state := fmt.Sprintf("%s (%d/%d)", episodeNames[selectedEpisodeIndex], selectedEpisodeIndex+1, totalEpisodes)
				if err := rpc.DiscordRPC(internal.RPCParams{
					Details:    selectedAnimeName,
					State:      state,
					LargeImage: posterUrl,
					LargeText:  selectedAnimeName,
				}); err != nil {
					logger.LogError(err)
				}
			}

		case "Çık":
			return
		default:
			return
		}
	}
}
