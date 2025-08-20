package tui

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"sort"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/reflow/truncate"
	"github.com/xeyossr/anitr-cli/internal"
	"github.com/xeyossr/anitr-cli/internal/utils"
)

// Renkler ve stil ayarları
var (
	highlightFgColor = "#e45cc0"
	normalFgColor    = "#aabbcc"
	highlightColor   = "#e45cc0"
	filterInputFg    = "#8bb27f"
	filterCursorFg   = "#c4b48b"
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
			Padding(0, 2)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color(normalFgColor)).
			Padding(0, 2)
)

// listItem, list elemanlarının türüdür
type listItem struct {
	title    string
	selected bool
}

// Title, listItem için başlık döndürür
func (i listItem) Title() string { return i.title }

// Description, listItem için açıklama döndürür (bu örnekte boş)
func (i listItem) Description() string { return "" }

// FilterValue, listItem için filtre değeri döndürür
func (i listItem) FilterValue() string { return i.title }

// slimDelegate, listDelegate'in bir özelleştirilmiş versiyonudur
type slimDelegate struct {
	list.DefaultDelegate
	multiSelect bool
}

// Height, item'in yüksekliğini döndürür
func (d slimDelegate) Height() int { return 1 }

// Spacing, item'ler arasındaki boşluğu döndürür
func (d slimDelegate) Spacing() int { return 0 }

// Render, item'in nasıl render edileceğini belirler
func (d slimDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	i, ok := item.(listItem)
	if !ok {
		return
	}

	prefix := ""
	if d.multiSelect {
		checkbox := "[ ]"
		if i.selected {
			checkbox = "[x]"
		}
		prefix = fmt.Sprintf("%s ", checkbox)
	}

	// Seçili olup olmadığını kontrol et
	isSelected := index == m.Index()

	if isSelected {
		prefix = selectionMark + prefix
	} else {
		prefix = "  " + prefix
	}

	// Alan genişliğini hesapla
	availableWidth := m.Width() - lipgloss.Width(prefix) - 4

	// Başlık, taşma durumuna göre kısaltılır
	displayTitle := truncate.StringWithTail(i.title, uint(availableWidth), "...")

	// Satırı oluştur
	line := prefix + displayTitle

	// Eğer seçiliyse, stili değiştir
	if isSelected {
		line = highlightStyle.Render(line)
	} else {
		line = normalStyle.Render(line)
	}

	// Satırı yazdır
	fmt.Fprint(w, line)
}

// SelectionListModel, seçim listesini tutan modeldir
type SelectionListModel struct {
	list         list.Model
	quitting     bool
	selected     []string
	selectedMap  map[string]struct{}
	multiSelect  bool // Add this field
	err          error
	width        int
}

// NewSelectionListModel, yeni bir SelectionListModel oluşturur
func NewSelectionListModel(params internal.UiParams) SelectionListModel {
	items := make([]list.Item, len(*params.List))
	for i, v := range *params.List {
		items[i] = listItem{title: v, selected: false}
	}

	const defaultWidth = 48
	const defaultHeight = 20

	multiSelect := params.Type == "multi-select"
	l := list.New(items, slimDelegate{multiSelect: multiSelect}, defaultWidth, defaultHeight)

	titleStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
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
		list:        l,
		selectedMap: make(map[string]struct{}),
		multiSelect: params.Type == "multi-select", // Set based on param
	}
}

// Init, başlangıçta yapılacak işlemi döndürür (boş)
func (m SelectionListModel) Init() tea.Cmd {
	return nil
}

// Update, kullanıcı etkileşimini günceller
func (m SelectionListModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "enter":
			if m.multiSelect {
				for k := range m.selectedMap {
					m.selected = append(m.selected, k)
				}
			} else {
				if i, ok := m.list.SelectedItem().(listItem); ok {
					m.selected = []string{i.title}
				}
			}
			sort.Strings(m.selected)
			m.quitting = true
			return m, tea.Quit

		case " ": // Handle spacebar for multi-selection
			if m.multiSelect {
				if i, ok := m.list.SelectedItem().(listItem); ok {
					if _, found := m.selectedMap[i.title]; found {
						delete(m.selectedMap, i.title)
						i.selected = false
					} else {
						m.selectedMap[i.title] = struct{}{} // Add to map
						i.selected = true
					}
					m.list.SetItem(m.list.Index(), i)
				}
			}

		case "ctrl+c", "esc", "q":
			m.err = utils.ErrQuit
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View, modelin görünümünü döndürür
func (m SelectionListModel) View() string {
	if m.quitting {
		return ""
	}
	return m.list.View()
}

// SelectionList, bir seçim listesi gösterir ve kullanıcının seçimini döner
func SelectionList(params internal.UiParams) ([]string, error) { // Changed return type to []string
	// Yeni bir program başlat ve seçimi al
	p := tea.NewProgram(NewSelectionListModel(params), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		if params.Logger != nil {
			params.Logger.LogError(fmt.Errorf("bubbletea p.Run() error in SelectionList: %w", err))
		}
		return nil, err // Return nil slice on error
	}
	model := m.(SelectionListModel)
	if model.err != nil {
		return nil, model.err // Return nil slice on error
	}
	return model.selected, nil // Return the slice of selected items
}

// InputFromUserModel, kullanıcıdan giriş almak için kullanılan modeldir
type InputFromUserModel struct {
	textInput textinput.Model
	err       error
	quitting  bool
}

// NewInputFromUserModel, yeni bir giriş modelini başlatır
func NewInputFromUserModel(params internal.UiParams) InputFromUserModel {
	ti := textinput.New()
	ti.Placeholder = ""
	ti.Prompt = "🔍 " + params.Label + ": "
	ti.CharLimit = 256
	ti.Focus()

	// Prompt ve metin stillerini ayarla
	ti.PromptStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(inputPromptFg)).Bold(true)
	ti.TextStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(inputTextFg))
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(inputCursorFg))

	return InputFromUserModel{
		textInput: ti,
	}
}

// Init, giriş modelini başlatır
func (m InputFromUserModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update, giriş modelini günceller
func (m InputFromUserModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Tuşlara göre işlem yap
		switch msg.String() {
		case "enter":
			if len(strings.TrimSpace(m.textInput.Value())) == 0 {
				m.err = errors.New("boş bırakılamaz")
				return m, nil
			}
			m.quitting = true
			return m, tea.Quit
		case "ctrl+c", "esc":
			m.err = utils.ErrQuit
			m.quitting = true
			return m, tea.Quit
		}
	}
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// View, giriş modelinin görünümünü döndürür
func (m InputFromUserModel) View() string {
	if m.quitting {
		return ""
	}
	return lipgloss.NewStyle().Padding(0, 2).Render(m.textInput.View())
}

// InputFromUser, kullanıcıdan giriş alır
func InputFromUser(params internal.UiParams) (string, error) {
	// Yeni bir program başlat ve kullanıcıdan giriş al
	p := tea.NewProgram(NewInputFromUserModel(params), tea.WithAltScreen())
	m, err := p.Run()

	if err != nil {
		if params.Logger != nil {
			params.Logger.LogError(fmt.Errorf("bubbletea p.Run() error in InputFromUser: %w", err))
		}
		return "", err
	}

	model := m.(InputFromUserModel)
	if model.err != nil {
		return "", model.err
	}

	return model.textInput.Value(), nil
}