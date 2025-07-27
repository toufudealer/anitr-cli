package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/xeyossr/anitr-cli/internal"
)

var (
	highlightFgColor = "#e45cc0"
	normalFgColor    = "#aabbcc"
	highlightColor   = "#e45cc0"
	filterInputFg    = "#8bb27f"
	filterCursorFg   = "#c4b48b"
	titleFg          = "#c4b48b"
	titleBg          = "#2a2c36"
	inputPromptFg    = "#c4b48b"
	inputTextFg      = "#aabbcc"
	inputCursorFg    = "#c4b48b"
	selectionMark    = "▸ "

	pinkHighlight = lipgloss.NewStyle().Foreground(lipgloss.Color(highlightColor))

	filterInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color(filterInputFg)).
				Bold(true)

	highlightStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(highlightFgColor)).
			Bold(true).
			Padding(0, 1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(normalFgColor)).
			Padding(0, 1)
)

type listItem string

func (i listItem) Title() string       { return string(i) }
func (i listItem) Description() string { return "" }
func (i listItem) FilterValue() string { return string(i) }

type slimDelegate struct {
	list.DefaultDelegate
}

func (d slimDelegate) Height() int  { return 1 }
func (d slimDelegate) Spacing() int { return 0 }

func (d slimDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title := ""
	if li, ok := item.(listItem); ok {
		title = li.Title()
	} else {
		title = "???"
	}

	isSelected := index == m.Index()

	prefix := "  "
	if isSelected {
		prefix = selectionMark
	}

	availableWidth := m.Width() - lipgloss.Width(prefix) - 4

	displayTitle := truncate.StringWithTail(title, uint(availableWidth), "...")

	line := prefix + displayTitle

	if isSelected {
		line = highlightStyle.Render(line)
	} else {
		line = normalStyle.Render(line)
	}

	fmt.Fprint(w, line)
}

type SelectionListModel struct {
	list     list.Model
	quitting bool
	selected string
	err      error
	width    int
}

func NewSelectionListModel(params internal.UiParams) SelectionListModel {
	items := make([]list.Item, len(*params.List))
	for i, v := range *params.List {
		items[i] = listItem(v)
	}

	const defaultWidth = 48
	const defaultHeight = 20

	l := list.New(items, slimDelegate{}, defaultWidth, defaultHeight)

	titleStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Padding(0, 2).
		Background(lipgloss.Color(titleBg)).
		Foreground(lipgloss.Color(titleFg)).
		Bold(true)

	l.Title = titleStyle.Render(params.Label)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	l.FilterInput.Prompt = pinkHighlight.Render("🔍 Search: ")
	l.FilterInput.Placeholder = "Ara..."
	l.FilterInput.TextStyle = filterInputStyle
	l.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(filterCursorFg))

	return SelectionListModel{
		list: l,
	}
}

func (m SelectionListModel) Init() tea.Cmd {
	return nil
}

func (m SelectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(listItem); ok {
				m.selected = string(i)
			}
			m.quitting = true
			return m, tea.Quit

		case "ctrl+c", "esc":
			m.err = errors.New("iptal edildi")
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m SelectionListModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

func SelectionList(params internal.UiParams) (string, error) {
	p := tea.NewProgram(NewSelectionListModel(params), tea.WithAltScreen())
	m, err := p.StartReturningModel()
	if err != nil {
		return "", err
	}
	model := m.(SelectionListModel)
	if model.err != nil {
		return "", model.err
	}
	return model.selected, nil
}

type InputFromUserModel struct {
	textInput textinput.Model
	err       error
	quitting  bool
}

func NewInputFromUserModel(params internal.UiParams) InputFromUserModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "🔍 " + params.Label + ": "
	ti.CharLimit = 256
	ti.Focus()

	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(inputPromptFg)).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(inputTextFg))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(inputCursorFg))

	return InputFromUserModel{
		textInput: ti,
	}
}

func (m InputFromUserModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m InputFromUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if len(strings.TrimSpace(m.textInput.Value())) == 0 {
				m.err = errors.New("boş bırakılamaz")
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.err = errors.New("iptal edildi")
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m InputFromUserModel) View() string {
	if m.quitting {
		return ""
	}
	return lipgloss.NewStyle().Padding(0, 2).Render(m.textInput.View())
}

func InputFromUser(params internal.UiParams) (string, error) {
	p := tea.NewProgram(NewInputFromUserModel(params))
	m, err := p.Run()
	if err != nil {
		return "", err
	}
	model := m.(InputFromUserModel)
	if model.err != nil {
		return "", model.err
	}
	return model.textInput.Value(), nil
}
