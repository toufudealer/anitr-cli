package openanime

import (
	"fmt"
	"strings"

	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/models"
	"github.com/xeyossr/anitr-cli/internal/utils"
)

type OpenAnime struct{}

var configOpenAnime = internal.Config{
	BaseUrl:      "https://api.openani.me",
	VideoPlayers: []string{"https://de2---vn-t9g4tsan-5qcl.yeshi.eu.org"},
	HttpHeaders:  map[string]string{"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/133.0.0.0 Safari/537.36", "Origin": "https://openani.me", "Referer": "https://openani.me", "Accept": "application/json"},
}

func (o OpenAnime) Source() string {
	return "openanime"
}

func (o OpenAnime) GetSearchData(query string) ([]models.Anime, error) {
	normalizedQuery := utils.NormalizeTurkishToASCII(query)
	normalizedQuery = strings.ReplaceAll(normalizedQuery, " ", "+")
	url := fmt.Sprintf("%s/anime/search?q=%s", configOpenAnime.BaseUrl, normalizedQuery)
	data, err := internal.GetJson(url, configOpenAnime.HttpHeaders)
	if err != nil {
		return nil, fmt.Errorf("arama verileri alınamadı: %w", err)
	}

	var returnData []models.Anime
	for _, item := range data.([]interface{}) {
		anime, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("geçersiz anime veri formatı")
		}

		name, ok := anime["english"].(string)
		if !ok {
			name = ""
		}

		slug, ok := anime["slug"].(string)
		if !ok {
			slug = ""
		}

		poster, ok := anime["pictures"].(map[string]interface{})["avatar"].(string)
		if !ok {
			poster = ""
		}

		returnData = append(returnData, models.Anime{
			Slug:     &slug,
			Title:    name,
			Source:   "openanime",
			ImageURL: poster,
		})
	}

	return returnData, nil
}

func (o OpenAnime) GetSeasonsData(params models.SeasonParams) ([]models.Season, error) {
	url := fmt.Sprintf("%s/anime/%s", configOpenAnime.BaseUrl, *params.Slug)
	data, err := internal.GetJson(url, configOpenAnime.HttpHeaders)
	if err != nil {
		return nil, fmt.Errorf("sezon verileri alınamadı: %w", err)
	}

	seasonData, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("sezon verisi beklenen formatta değil")
	}

	seasonCount, ok := seasonData["numberOfSeasons"].(float64)
	if !ok {
		seasonCount = 1
	}

	contentType := seasonData["type"].(string)
	isMovie := strings.ToLower(contentType) == "movie"

	return []models.Season{
		{
			Seasons: &[]int{int(seasonCount)},
			Type:    &contentType,
			IsMovie: &isMovie,
		},
	}, nil
}

func (o OpenAnime) GetEpisodesData(params models.EpisodeParams) ([]models.Episode, error) {
	seasonData, err := o.GetSeasonsData(models.SeasonParams{Slug: params.Slug})
	if err != nil {
		return nil, fmt.Errorf("sezon bilgisi alınamadı: %w", err)
	}

	var episodes []models.Episode
	seasondata := *seasonData[0].Seasons
	seasonCount := int(seasondata[0])

	for season := 1; season <= seasonCount; season++ {
		url := fmt.Sprintf("%s/anime/%s/season/%d", configOpenAnime.BaseUrl, *params.Slug, season)
		data, err := internal.GetJson(url, configOpenAnime.HttpHeaders)
		if err != nil {
			return nil, fmt.Errorf("sezon %d için bölüm verileri alınamadı: %w", season, err)
		}

		seasonInfo, ok := data.(map[string]interface{})["season"].(map[string]interface{})
		if !ok {
			continue
		}

		episodesRaw, ok := seasonInfo["episodes"].([]interface{})
		if !ok {
			continue
		}

		for _, episodeRaw := range episodesRaw {
			episode, ok := episodeRaw.(map[string]interface{})
			if !ok {
				continue
			}

			episodeNumber, _ := episode["episodeNumber"].(float64)
			seasonNumber, _ := seasonInfo["season_number"].(float64)
			name := fmt.Sprintf("%d. Sezon, %d. Bölüm", int(seasonNumber), int(episodeNumber))

			episodes = append(episodes, models.Episode{
				Title:  name,
				Number: int(episodeNumber),
				Extra: map[string]interface{}{
					"season_num": seasonNumber,
				},
			})
		}
	}

	return episodes, nil
}

func (o OpenAnime) GetFansubsData(params models.FansubParams) ([]models.Fansub, error) {
	if params.Slug == nil || params.SeasonNum == nil || params.EpisodeNum == nil {
		return nil, fmt.Errorf("slug, sezon numarası veya bölüm numarası eksik")
	}

	slug := *params.Slug
	seasonNum := *params.SeasonNum
	episodeNum := *params.EpisodeNum

	url := fmt.Sprintf("%s/anime/%s/season/%d/episode/%d", configOpenAnime.BaseUrl, slug, seasonNum, episodeNum)
	data, err := internal.GetJson(url, configOpenAnime.HttpHeaders)
	if err != nil {
		return nil, fmt.Errorf("fansub verileri alınamadı: %w", err)
	}

	rawFansubs, ok := data.(map[string]interface{})["fansubs"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("fansubs verisi eksik veya hatalı")
	}

	var fansubs []models.Fansub
	for _, f := range rawFansubs {
		fm, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		is4K, ok := fm["is4K"].(bool)
		if !ok || is4K {
			continue
		}

		id, idOK := fm["id"].(string)
		name, nameOK := fm["name"].(string)
		secureName, secureOK := fm["secureName"].(string)

		if !idOK || !nameOK || !secureOK {
			return nil, fmt.Errorf("fansub bilgisi eksik: %+v", fm)
		}

		fansubs = append(fansubs, models.Fansub{
			ID:         &id,
			Name:       &name,
			SecureName: &secureName,
		})
	}

	if len(fansubs) == 0 {
		return nil, fmt.Errorf("geçerli fansub bulunamadı")
	}

	return fansubs, nil
}

func (o OpenAnime) GetWatchData(req models.WatchParams) ([]models.Watch, error) {
	if req.Slug == nil || req.Extra == nil {
		return nil, fmt.Errorf("slug veya ekstra bilgiler eksik")
	}

	slug := *req.Slug
	extra := *req.Extra

	seasonNum, ok := extra["season_num"].(int)
	if !ok {
		return nil, fmt.Errorf("season_num geçersiz veya eksik")
	}

	episodeNum, ok := extra["episode_num"].(int)
	if !ok {
		return nil, fmt.Errorf("episode_num geçersiz veya eksik")
	}

	baseURL := fmt.Sprintf("%s/anime/%s/season/%d/episode/%d", configOpenAnime.BaseUrl, slug, int(seasonNum), int(episodeNum))

	fansubs := extra["fansubs"].([]models.Fansub)
	selectedFansubId := extra["selected_fansub_id"].(int)

	videoURL := fmt.Sprintf("%s?fansub=%s", baseURL, *fansubs[selectedFansubId].ID)
	data, err := internal.GetJson(videoURL, configOpenAnime.HttpHeaders)
	if err != nil {
		return nil, fmt.Errorf("video bağlantıları alınamadı: %w", err)
	}

	episodeData, ok := data.(map[string]interface{})["episodeData"].(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("episodeData eksik veya hatalı")
	}

	files, ok := episodeData["files"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("video dosyaları bulunamadı veya geçersiz")
	}

	var labels []string
	var urls []string

	for _, f := range files {
		fileData, ok := f.(map[string]interface{})
		if !ok {
			continue
		}

		urlRaw, urlOK := fileData["file"].(string)
		url := fmt.Sprintf("%s/animes/%s/%d/%s", configOpenAnime.VideoPlayers[0], slug, seasonNum, urlRaw)
		resolutionVal, resOK := fileData["resolution"].(float64)

		if !urlOK || !resOK {
			continue
		}

		labels = append(labels, fmt.Sprintf("%dp", int(resolutionVal)))
		urls = append(urls, url)
	}

	if len(urls) == 0 {
		return nil, fmt.Errorf("geçerli video bağlantısı bulunamadı")
	}

	return []models.Watch{
		{
			Labels: labels,
			Urls:   urls,
		},
	}, nil
}
