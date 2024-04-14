package fuzzyfinder

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/theme"
)

type Searchable interface {
	SearchResultString() string
	Kind() string
	Display() string
	Implemented() bool
	IsMatch(string) bool
}

type Sources []Searchable

func (s Sources) Len() int {
	return len(s)
}

func (s Sources) String(i int) string {
	return s[i].SearchResultString()
}

type Screen struct {
	Width  int
	Height int
}

type keyMap struct {
	NextItem     key.Binding
	PreviousItem key.Binding
	Select       key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.NextItem, k.PreviousItem, k.Select}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{k.NextItem, k.PreviousItem, k.Select}}
}

var defaultKeys = keyMap{
	NextItem: key.NewBinding(
		key.WithKeys("ctrl+n", "down"),
		key.WithHelp("ctrl+n", "Next"),
	),
	PreviousItem: key.NewBinding(
		key.WithKeys("ctrl+p", "up"),
		key.WithHelp("ctrl+p", "Previous"),
	),
	Select: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "Select"),
	),
}

type Model struct {
	styles         theme.FuzzyFinderComponentStyles
	help           help.Model
	viewport       viewport.Model
	keys           keyMap
	items          Sources
	matches        fuzzy.Matches
	search         textinput.Model
	ctx            ctx.Context
	matchIndex     int
	unmatchedIndex int
	searchStarted  bool
	debug          bool
}

type ModelOption func(m *Model)

func WithStyles(styles theme.FuzzyFinderComponentStyles) ModelOption {
	return func(m *Model) {
		m.styles = styles
		m.search.PromptStyle = styles.PromptStyles.PromptStyle
		m.search.PlaceholderStyle = styles.PromptStyles.PlaceholderStyle
		m.search.TextStyle = styles.PromptStyles.TextStyle
	}
}

func WithDebug(debug bool) ModelOption {
	return func(m *Model) {
		m.debug = debug
	}
}

func New(ctx ctx.Context, items Sources, maxHeight int, options ...ModelOption) Model {

	m := &Model{
		ctx:            ctx,
		searchStarted:  false,
		help:           help.New(),
		viewport:       viewport.New(0, min(maxHeight, len(items))),
		keys:           defaultKeys,
		matches:        fuzzy.Matches{},
		items:          items,
		search:         textinput.New(),
		matchIndex:     0,
		unmatchedIndex: 0,
		// FIXME: add default styles
		styles: defaultStyles(),
		debug:  false,
	}

	for _, opt := range options {
		opt(m)
	}
	// m.styles.SelectedItem.Width(ctx.ScreenWidth)
	m.search.Placeholder = "Search for something"
	m.search.Focus()
	return *m
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if m.search.Value() == "" {
		m.matchIndex = 0
		m.searchStarted = false
	}
	if m.searchStarted && m.matchIndex > len(m.matches) {
		m.matchIndex = 0
	}

	if m.search.Value() != "" {
		m.searchStarted = true
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// m.styles.SelectedItem.Width(msg.Width)
		m.viewport.Width = m.ctx.ScreenWidth
	case tea.KeyMsg:
		switch {
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		case key.Matches(msg, m.keys.NextItem):
			if m.atListBottom() {
				m.viewport.LineDown(1)
			}
			m.move(next)
		case key.Matches(msg, m.keys.PreviousItem):
			if m.atListTop() {
				m.viewport.LineUp(1)
			}
			m.move(previous)
		case key.Matches(msg, m.keys.Select):
			item := m.items[m.unmatchedIndex]
			if m.searchStarted {
				item, _ = lo.Find(m.items, func(s Searchable) bool {
					return m.matches[m.matchIndex].Str == s.SearchResultString()
				})
			}
			if !item.Implemented() {
				return m, nil
			}
			return m, selectItem(item)
		default:
			m.search, cmd = m.search.Update(msg)
			m.matches = fuzzy.FindFrom(m.search.Value(), m.items)
			cmds = append(cmds, cmd)
			// Want to stay at the top when fuzzy searching
			if m.searchStarted {
				m.viewport.GotoTop()
			}
		}
	case SelectionMade:
		return m, tea.Quit
	}
	m.viewport.SetContent(m.viewportContent())

	return m, tea.Batch(cmds...)
}

type action = int

const (
	next     action = 1
	previous action = -1
)

func (m *Model) move(action action) {
	switch action {
	case next:
		if m.searchStarted && m.matchIndex < len(m.matches)-1 {
			m.matchIndex++
		}
		if m.unmatchedIndex < len(m.items)-1 {
			m.unmatchedIndex++
		}
	case previous:
		if m.searchStarted && m.matchIndex > 0 {
			m.matchIndex--
		}
		if m.unmatchedIndex > 0 {
			m.unmatchedIndex--
		}
	}
}

func (m Model) atListBottom() bool {
	if m.searchStarted {
		return m.matchIndex+1 == m.viewport.VisibleLineCount()+m.viewport.YOffset
	}
	return m.unmatchedIndex+1 == m.viewport.VisibleLineCount()+m.viewport.YOffset
}

func (m Model) atListTop() bool {
	if m.searchStarted {
		return m.matchIndex-m.viewport.YOffset == 0
	}
	return m.unmatchedIndex-m.viewport.YOffset == 0
}

type SelectionMade struct {
	Selection Searchable
}

func selectItem(selection Searchable) tea.Cmd {
	return func() tea.Msg {
		return SelectionMade{Selection: selection}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func addSpacesUntil(maxWidth int, strs ...string) string {
	componentWidths := 0
	for _, component := range strs {
		componentWidths += lipgloss.Width(component)
	}
	spacing := maxWidth - componentWidths
	return strings.Repeat(" ", max(0, spacing))
}

// When rendering a line line this with multiple different styles, you can't
// just apply over them like selectedStyle(kindStyle(kind) + Name + Spacing + Implemented)
// you have to style one part, and style another part
// returns styledKind + styledRestOfContent
func (m Model) renderItems() []string {
	return lo.Map(m.items, func(s Searchable, index int) string {
		implemented := ""
		if !s.Implemented() {
			implemented = "Not Implemented"
		}
		name := s.SearchResultString()
		if m.unmatchedIndex == index {
			kind := m.styles.SelectedItem.Copy().Foreground(m.styles.KindColor).Render(fmt.Sprintf("[%s] ", s.Kind()))
			spaces := addSpacesUntil(m.ctx.ScreenWidth, kind, s.Display(), implemented)
			return kind + m.styles.SelectedItem.Copy().Width(m.ctx.ScreenWidth-lipgloss.Width(kind)).Render(name+spaces+implemented)
		}
		kind := m.styles.UnselectedItem.Copy().Foreground(m.styles.KindColor).Render(fmt.Sprintf("[%s] ", s.Kind()))
		spaces := addSpacesUntil(m.ctx.ScreenWidth, kind, name, implemented)
		return kind + m.styles.UnselectedItem.Copy().Width(m.ctx.ScreenWidth-lipgloss.Width(kind)).Render(name+spaces+implemented)
	})
}

// When rendering a line line this with multiple different styles, you can't
// just apply over them like selectedStyle(kindStyle(kind) + Name + Spacing + Implemented)
// you have to style one part, and style another part
// returns styledKind + styledRestOfContent
func (m Model) renderMatches() []string {
	if len(m.matches) == 0 {
		return []string{"No matches found"}
	}
	return lo.Map(m.matches, func(match fuzzy.Match, index int) string {
		content := match.Str
		item, found := lo.Find(m.items, func(e Searchable) bool {
			return e.IsMatch(content)
		})
		if !found {
			return ""
		}
		implemented := ""
		if !item.Implemented() {
			implemented = "Not Implemented"
		}
		if m.searchStarted && m.matchIndex == index {
			kind := m.styles.SelectedItem.Copy().Foreground(m.styles.KindColor).Render(fmt.Sprintf("[%s] ", item.Kind()))
			spaces := addSpacesUntil(m.ctx.ScreenWidth, kind, content, implemented)
			return kind + m.styles.SelectedItem.Copy().Width(m.ctx.ScreenWidth-lipgloss.Width(kind)).Render(content+spaces+implemented)
		}
		kind := m.styles.UnselectedItem.Copy().Foreground(m.styles.KindColor).Render(fmt.Sprintf("[%s] ", item.Kind()))
		return kind + m.styles.UnselectedItem.Render(content)
	})
}

func (m Model) viewportContent() string {
	if m.searchStarted {
		return lipgloss.JoinVertical(lipgloss.Left, m.renderMatches()...)
	}
	return lipgloss.JoinVertical(lipgloss.Left, m.renderItems()...)
}

func (m Model) debugInformation() string {

	firstMatch := ""
	if len(m.matches) > 0 {
		firstMatch = m.matches[m.matchIndex].Str
	}
	return lipgloss.JoinVertical(
		lipgloss.Left,
		fmt.Sprintf("matches[0]: %s", firstMatch),
		fmt.Sprintf("YOffset: %d", m.viewport.YOffset),
		fmt.Sprintf("AtTop %t", m.viewport.AtTop()),
		fmt.Sprintf("AtBottom %t", m.viewport.AtBottom()),
		fmt.Sprintf("PastBottom %t", m.viewport.PastBottom()),
		fmt.Sprintf("VisibleLineCount %d", m.viewport.VisibleLineCount()),
		fmt.Sprintf("TotalLineCount %d", m.viewport.TotalLineCount()),
		fmt.Sprintf("len(matches): %d", len(m.matches)),
		fmt.Sprintf("unmatchedIndex: %d", m.unmatchedIndex),
		fmt.Sprintf("matchedIndex: %d", m.matchIndex),
	)
}

func (m Model) statusLine() string {
	command := m.styles.StatusLineCommand.Render("Find All")
	if m.searchStarted {
		command = m.styles.StatusLineCommand.Render("FZF All")
	}
	selection := fmt.Sprintf("[%d/%d]", m.unmatchedIndex+1, len(m.items))
	if m.searchStarted {
		selection = fmt.Sprintf("[%d/%d]", m.matchIndex+1, len(m.matches))
	}

	selectionIndicator := m.styles.StatusLineSelectionIndicator.Render(selection)
	help := m.styles.StatusLineHelp.Render("Ctrl+? - Help")
	spacing := m.styles.StatusLineSpacing.Render(
		strings.Repeat(" ", max(0, m.ctx.ScreenWidth-lipgloss.Width(command)-lipgloss.Width(selectionIndicator)-lipgloss.Width(help))),
	)
	return command + selectionIndicator + spacing + help
}

func (m Model) View() string {
	sections := []string{}
	if m.debug {
		sections = append(sections, m.debugInformation())
	}
	sections = append(sections, lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		BorderForeground(m.styles.TopBarColor).
		Render(strings.Repeat(" ", max(0, m.ctx.ScreenWidth))))
	sections = append(sections, m.viewport.View())
	sections = append(sections, m.search.View())
	sections = append(sections, m.statusLine())
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func defaultStyles() theme.FuzzyFinderComponentStyles {
	return theme.FuzzyFinderComponentStyles{
		StatusLineCommand: lipgloss.NewStyle().
			PaddingRight(1).
			PaddingLeft(1),
		StatusLineSpacing:            lipgloss.Style{},
		StatusLineSelectionIndicator: lipgloss.Style{},
		StatusLineHelp:               lipgloss.Style{},
		SelectedItem: lipgloss.NewStyle().
			Background(lipgloss.AdaptiveColor{
				Light: "#303446",
				Dark:  "#303446",
			}).
			Foreground(lipgloss.AdaptiveColor{
				Light: "#c6d0f5",
				Dark:  "#c6d0f5",
			}),
		UnselectedItem: lipgloss.Style{},
		TopBarColor: lipgloss.AdaptiveColor{
			Light: "#303446",
			Dark:  "#303446",
		},
		PromptStyles: theme.FuzzyFinderPromptStyles{
			PromptStyle:      lipgloss.Style{},
			TextStyle:        lipgloss.Style{},
			PlaceholderStyle: lipgloss.Style{},
		},
	}
}
