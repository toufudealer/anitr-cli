package models

type AnimeSource interface {
	GetSearchData(query string) ([]Anime, error)
	GetSeasonsData(params SeasonParams) ([]Season, error)
	GetEpisodesData(params EpisodeParams) ([]Episode, error)
	GetWatchData(params WatchParams) ([]Watch, error)
	Source() string
}

type Anime struct {
	Title     string
	ID        *int
	Slug      *string
	Type      *string
	TitleType *string
	ImageURL  string
	Source    string
	Extra     map[string]interface{}
}

type Season struct {
	Seasons *[]int
	Count   *int
	Type    *string
	IsMovie *bool
}

type Episode struct {
	ID     string
	Title  string
	Number int
	Extra  map[string]interface{}
}

type WatchParams struct {
	Slug    *string
	Url     *string
	Id      *int
	IsMovie *bool
	Extra   *map[string]interface{}
}

type SeasonParams struct {
	Slug *string
	Id   *int
}

type EpisodeParams struct {
	Slug     *string
	SeasonID *int
}

type Watch struct {
	Labels    []string
	Urls      []string
	TRCaption *string
}
