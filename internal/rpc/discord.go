package rpc

import (
	"fmt"

	"github.com/hugolgst/rich-go/client"
	"github.com/xeyossr/anitr-cli/internal"
)

func ClientLogin() (bool, error) {
	err := client.Login("1383421771159572600")
	if err != nil {
		return false, fmt.Errorf("failed to log in to Discord RPC: %v", err)
	}

	return true, nil
}

func DiscordRPC(params internal.RPCParams, loggedIn bool) (bool, error) {
	if !loggedIn {
		ClientLogin()
		loggedIn = true
	}

	err := client.SetActivity(client.Activity{
		State:      params.State,
		Details:    params.Details,
		LargeImage: params.LargeImage,
		LargeText:  params.LargeText,
		SmallImage: params.SmallImage,
		SmallText:  params.SmallText,
		Buttons: []*client.Button{
			{
				Label: "GitHub",
				Url:   "https://github.com/xeyossr/anitr-cli",
			},
		},
	})

	if err != nil {
		loggedIn = false
		return loggedIn, fmt.Errorf("failed to set activity: %v", err)
	}

	return loggedIn, nil
}

func RPCDetails(details, state, largeimg, largetext string) internal.RPCParams {
	return internal.RPCParams{
		Details:    details,
		State:      state,
		LargeImage: largeimg,
		LargeText:  largetext,
	}
}
