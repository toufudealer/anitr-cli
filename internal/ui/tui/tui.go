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
type listItem string

// Title, listItem için başlık döndürür
func (i listItem) Title() string { return string(i) }

// Description, listItem için açıklama döndürür (bu örnekte boş)
func (i listItem) Description() string { return "" }

// FilterValue, listItem için filtre değeri döndürür
func (i listItem) FilterValue() string { return string(i) }

// slimDelegate, listDelegate'in bir özelleştirilmiş versiyonudur
type slimDelegate struct {
	list.DefaultDelegate
}

// Height, item'in yüksekliğini döndürür
func (d slimDelegate) Height() int { return 1 }

// Spacing, item'ler arasındaki boşluğu döndürür
func (d slimDelegate) Spacing() int { return 0 }

// Render, item'in nasıl render edileceğini belirler
func (d slimDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	title := ""
	if li, ok := item.(listItem); ok {
		title = li.Title()
	} else {
		title = "???"
	}

	// Seçili olup olmadığını kontrol et
	isSelected := index == m.Index()

	prefix := "  "
	if isSelected {
		prefix = selectionMark
	}

	// Alan genişliğini hesapla
	availableWidth := m.Width() - lipgloss.Width(prefix) - 4

	// Başlık, taşma durumuna göre kısaltılır
	displayTitle := truncate.StringWithTail(title, uint(availableWidth), "...")

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
	list     list.Model
	quitting bool
	selected string
	err      error
	width    int
}

// NewSelectionListModel, yeni bir SelectionListModel oluşturur
func NewSelectionListModel(params internal.UiParams) SelectionListModel {
	// Seçenekleri listeye ekle
	items := make([]list.Item, len(*params.List))
	for i, v := range *params.List {
		items[i] = listItem(v)
	}

	// Listeyi başlat
	const defaultWidth = 48
	const defaultHeight = 20

	l := list.New(items, slimDelegate{}, defaultWidth, defaultHeight)

	// Başlık stilini ayarla
	titleStyle := lipgloss.NewStyle().
		Align(lipgloss.Center).
		Bold(true)

	l.Title = titleStyle.Render(params.Label)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(true)

	// Filtreleme giriş stilini ayarla
	l.FilterInput.Prompt = pinkHighlight.Render("🔍 Search: ")
	l.FilterInput.Placeholder = "Ara..."
	l.FilterInput.TextStyle = filterInputStyle
	l.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(lipgloss.Color(filterCursorFg))

	return SelectionListModel{
		list: l,
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
		// Pencere boyutu değiştiğinde listeyi yeniden boyutlandır
		m.width = msg.Width
		m.list.SetSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyMsg:
		// Tuşlara göre işlem yap
		switch msg.String() {
		case "enter":
			if i, ok := m.list.SelectedItem().(listItem); ok {
				m.selected = string(i)
			}
			m.quitting = true
			return m, tea.Quit

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
func SelectionList(params internal.UiParams) (string, error) {
	// Yeni bir program başlat ve seçimi al
	p := tea.NewProgram(NewSelectionListModel(params), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		if params.Logger != nil {
			params.Logger.LogError(fmt.Errorf("bubbletea p.Run() error in SelectionList: %w", err))
		}
		return "", err
	}
	model := m.(SelectionListModel)
	if model.err != nil {
		return "", model.err
	}
	return model.selected, nil
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