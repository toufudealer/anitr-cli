package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Config struct {
	BaseUrl        string
	AlternativeUrl string
	VideoPlayers   []string
	HttpHeaders    map[string]string
}

type UiParams struct {
	Mode      string
	List      *[]string
	Label     string
	RofiFlags *string
}

type RPCParams struct {
	Details    string
	State      string
	LargeImage string
	LargeText  string
	SmallImage string
	SmallText  string
}

func GetStringPtr(m map[string]interface{}, key string) *string {
	if val, ok := m[key].(string); ok {
		return &val
	}
	return nil
}

func GetString(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func GetJson(url string, headers map[string]string) (interface{}, error) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği oluşturulamadı: %w", err)
	}

	for k, m := range headers {
		req.Header.Set(k, m)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("HTTP isteği başarısız: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("HTTP yanıtı okunamadı: %w", err)
	}

	var result interface{}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("JSON ayrıştırma başarısız: %w", err)
	}

	return result, nil
}
