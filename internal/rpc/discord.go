package rpc

import (
	"fmt"
	"time"

	"github.com/hugolgst/rich-go/client"
	"github.com/xeyossr/anitr-cli/internal"
)

var loggedIn bool

func DiscordRPC(params internal.RPCParams) error {
	if !loggedIn {
		err := client.Login("1383421771159572600")
		if err != nil {
			return fmt.Errorf("failed to log in to Discord RPC: %v", err)
		}
		loggedIn = true
	}

	now := time.Now()
	err := client.SetActivity(client.Activity{
		State:      params.State,
		Details:    fmt.Sprintf("Watching %s", params.Details),
		LargeImage: params.LargeImage,
		LargeText:  params.LargeText,
		//SmallImage: params.SmallImage,
		//SmallText:  params.SmallText,
		Timestamps: &client.Timestamps{
			Start: &now,
		},
		Buttons: []*client.Button{
			{
				Label: "GitHub",
				Url:   "https://github.com/xeyossr/anitr-cli",
			},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to set activity: %v", err)
	}

	return nil
}

func RPCDetails(details, state, largeimg, largetext string) internal.RPCParams {
	return internal.RPCParams{
		Details:    details,
		State:      state,
		LargeImage: largeimg,
		LargeText:  largetext,
	}
}
