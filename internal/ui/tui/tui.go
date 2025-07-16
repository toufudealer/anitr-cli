package tui

import (
	"errors"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/xeyossr/anitr-cli/internal"
)

func SelectionList(params internal.UiParams) (string, error) {
	searcher := func(input string, index int) bool {
		list := *params.List
		item := strings.ToLower(list[index])
		input = strings.ToLower(input)

		for _, word := range strings.Fields(input) {
			if !strings.Contains(item, word) {
				return false
			}
		}
		return true
	}

	prompt := promptui.Select{
		Label:    params.Label,
		Items:    *params.List,
		Size:     13,
		Searcher: searcher,
	}

	_, result, err := prompt.Run()

	if err != nil {
		return "", err
	}

	return result, nil
}

func InputFromUser(params internal.UiParams) (string, error) {
	validate := func(input string) error {
		if len(input) < 1 {
			return errors.New("boş bırakılamaz")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:    params.Label,
		Validate: validate,
	}

	result, err := prompt.Run()
	if err != nil {
		return "", err
	}

	return result, nil
}
