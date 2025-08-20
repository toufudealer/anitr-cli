package history

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

// WatchedEpisode represents the last watched episode for a given anime.
type WatchedEpisode struct {
	AnimeID       string `json:"anime_id"` // Using string for flexibility (ID or Slug)
	LastEpisode int    `json:"last_episode"`
}

// History stores a map of anime ID/slug to their last watched episode.
type History struct {
	Watched map[string]WatchedEpisode `json:"watched"`
	filePath string
}

// NewHistory creates a new History instance and loads data from the specified file.
func NewHistory(dataDir string) (*History, error) {
	history := &History{
		Watched:  make(map[string]WatchedEpisode),
		filePath: filepath.Join(dataDir, "watched_history.json"),
	}
	err := history.Load()
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return history, nil
}

// Load reads the watched history from the JSON file.
func (h *History) Load() error {
	data, err := ioutil.ReadFile(h.filePath)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &h.Watched)
}

// Save writes the watched history to the JSON file.
func (h *History) Save() error {
	data, err := json.MarshalIndent(h.Watched, "", "  ")
	if err != nil {
		return err
	}
	return ioutil.WriteFile(h.filePath, data, 0644)
}

// GetLastWatchedEpisode retrieves the last watched episode for a given anime.
func (h *History) GetLastWatchedEpisode(animeID string) (int, bool) {
	ep, ok := h.Watched[animeID]
	if !ok {
		return 0, false
	}
	return ep.LastEpisode, true
}

// SetLastWatchedEpisode sets the last watched episode for a given anime.
func (h *History) SetLastWatchedEpisode(animeID string, episodeNum int) {
	h.Watched[animeID] = WatchedEpisode{
		AnimeID:       animeID,
		LastEpisode: episodeNum,
	}
}
