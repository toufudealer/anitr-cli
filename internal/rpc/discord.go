package rpc

import (
	"fmt"

	"github.com/hugolgst/rich-go/client"
	"github.com/xeyossr/anitr-cli/internal"
)

// ClientLogin, Discord RPC'ye giriş yapmaya çalışır ve başarı durumunu döner.
func ClientLogin() (bool, error) {
	// Discord RPC'ye giriş yapmayı dene
	err := client.Login("1383421771159572600")
	if err != nil {
		return false, fmt.Errorf("discord rpc login başarısız: %v", err) // Giriş hatası
	}

	return true, nil // Başarılı giriş
}

// DiscordRPC, Discord'a RPC (Remote Procedure Call) aktivitesi güncellemeleri gönderir.
func DiscordRPC(params internal.RPCParams, loggedIn bool) (bool, error) {
	// Eğer Discord'a giriş yapılmamışsa, giriş yap
	if !loggedIn {
	    ok, err := ClientLogin()
	   	if err != nil || !ok {
	    	return false, fmt.Errorf("discord rpc login başarısız: %v", err)
	    }
		loggedIn = true
	}

	// Discord aktivitesini ayarla
	activityDetails := params.AnimeTitle // Details will always be the AnimeTitle

	activityState := ""
	if params.CurrentEpisode > 0 && params.TotalEpisodes > 0 {
		activityState = fmt.Sprintf("Bölüm %d / %d - %s", params.CurrentEpisode, params.TotalEpisodes, params.EpisodeTitle)
	} else if params.CurrentEpisode > 0 {
		activityState = fmt.Sprintf("Bölüm %d - %s", params.CurrentEpisode, params.EpisodeTitle)
	} else if params.EpisodeTitle != "" {
		activityState = params.EpisodeTitle // Fallback to EpisodeTitle if no episode numbers
	}

	err := client.SetActivity(client.Activity{
		State:      activityState,     // Aktivite durumu
		Details:    activityDetails,   // Aktivite detayları
		LargeImage: params.LargeImage, // Büyük resim
		LargeText:  params.LargeText,  // Büyük resim açıklaması
		SmallImage: params.SmallImage, // Küçük resim
		SmallText:  params.SmallText,  // Küçük resim açıklaması
		Buttons: []*client.Button{ // Butonlar
			{
				Label: "MyAnimeList",
				Url:   params.MyAnimeListURL, // MyAnimeList bağlantısı
			},
			{
				Label: "GitHub",
				Url:   "https://github.com/toufudealer/anitr-cli", // GitHub bağlantısı
			},
		},
	})

	// Eğer aktivite güncelleme hatalıysa
	if err != nil {
		loggedIn = false
		ok, err := ClientLogin()
		if err != nil || !ok {
			return false, fmt.Errorf("discord rpc yeniden login başarısız: %v", err)
		}

		err = client.SetActivity(client.Activity{
			State:      activityState,
			Details:    activityDetails,
			LargeImage: params.LargeImage,
			LargeText:  params.LargeText,
			SmallImage: params.SmallImage,
			SmallText:  params.SmallText,
			Buttons: []*client.Button{
				{
					Label: "MyAnimeList",
					Url:   params.MyAnimeListURL,
				},
				{
					Label: "GitHub",
					Url:   "https://github.com/toufudealer/anitr-cli",
				},
			},
		})

		if err != nil {
			return false, fmt.Errorf("discord rpc retry setactivity başarısız: %v", err)
		}

		loggedIn = true
	}

	return loggedIn, nil // Başarılı RPC güncellemesi
}

// RPCDetails, Discord RPC için gerekli parametreleri hazırlar ve döner.
func RPCDetails(animeTitle, episodeTitle string, currentEpisode, totalEpisodes int, largeimg, largetext string) internal.RPCParams {
	// RPC parametrelerini yapılandır
	return internal.RPCParams{
		AnimeTitle:    animeTitle,
		EpisodeTitle:  episodeTitle,
		CurrentEpisode: currentEpisode,
		TotalEpisodes:  totalEpisodes,
		LargeImage:    largeimg,
		LargeText:     largetext,
	}
}
