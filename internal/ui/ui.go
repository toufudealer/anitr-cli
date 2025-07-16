package ui

import (
	"os"
	"os/exec"

	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/ui/rofi"
	"github.com/xeyossr/anitr-cli/internal/ui/tui"
)

func ClearScreen() {
	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func SelectionList(params internal.UiParams) (string, error) {
	if params.Mode == "rofi" {
		response, err := rofi.SelectionList(params)
		if err != nil {
			return "", nil
		}

		return response, nil
	}

	response, err := tui.SelectionList(params)
	if err != nil {
		return "", nil
	}

	return response, nil
}

func InputFromUser(params internal.UiParams) (string, error) {
	if params.Mode == "rofi" {
		response, err := rofi.InputFromUser(params)
		if err != nil {
			return "", nil
		}

		return response, nil
	}

	response, err := tui.InputFromUser(params)
	if err != nil {
		return "", nil
	}

	return response, nil
}
