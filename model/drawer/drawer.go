package drawer

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/types"
)

type keyMap struct {
	Confirm    key.Binding
	Cancel     key.Binding
	ScrollDown key.Binding
	ScrollUp   key.Binding
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.ScrollUp, k.ScrollDown, k.Confirm, k.Cancel},
	}
}

var defaultKeys = keyMap{
	ScrollUp: key.NewBinding(
		key.WithKeys("ctrl+k", "up"),
		key.WithHelp("ctrl+k", "scroll up"),
	),
	ScrollDown: key.NewBinding(
		key.WithKeys("ctrl+j", "down"),
		key.WithHelp("ctrl+j", "scroll down"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "confirm"),
	),
	Cancel: key.NewBinding(
		key.WithKeys("esc", "ctrl+o"),
		key.WithHelp("esc/ctrl+o", "cancel"),
	),
}

type Pickable interface {
	Info() string
	Thing() any
}
type Model struct {
	statusLineStyle    lipgloss.Style
	selectedItemStyle  lipgloss.Style
	topItemStyle       lipgloss.Style
	helpIndicatorStyle lipgloss.Style
	keys               keyMap
	items              []Pickable
	ctx                ctx.Context
	index              int
	focused            bool
	enabled            bool
}

func New(ctx ctx.Context, items []Pickable) Model {
	var (
		// FIXME: move these to themes, double-check colors
		defaultStatusLineStyle = lipgloss.NewStyle().
					Background(lipgloss.Color("#e78284")).
					Foreground(lipgloss.Color("#232634")).
					PaddingLeft(1).
					PaddingRight(1).
					Bold(true)
		defaultSelectedItemStyle = lipgloss.NewStyle().
						Background(lipgloss.Color("#414559")).
						Width(ctx.ScreenWidth)
		defaultTopItemStyle = lipgloss.NewStyle().
					BorderTop(true).
					BorderStyle(lipgloss.NormalBorder()).
					Width(ctx.ScreenWidth)
		defaultHelpIndicatorStyle = lipgloss.NewStyle().
						Background(lipgloss.Color("#8caaee")).
						Foreground(lipgloss.Color("#232634")).
						PaddingLeft(1).
						PaddingRight(1).
						Bold(true)
	)
	return Model{
		ctx:                ctx,
		index:              0,
		items:              items,
		focused:            false,
		keys:               defaultKeys,
		statusLineStyle:    defaultStatusLineStyle,
		selectedItemStyle:  defaultSelectedItemStyle,
		topItemStyle:       defaultTopItemStyle,
		helpIndicatorStyle: defaultHelpIndicatorStyle,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.ScrollDown):
			if m.index == len(m.items)-1 {
				return m, nil
			}
			m.index += 1
		case key.Matches(msg, m.keys.ScrollUp):
			if m.index == 0 {
				return m, nil
			}
			m.index -= 1
		case key.Matches(msg, m.keys.Confirm):
			m.ctx.Logger.Debug("They hit the enter key")
			return m, Select(m.items[m.index])
		case key.Matches(msg, m.keys.Cancel):
			return m, DrawerClose
		}
	}
	return m, nil
}

type DrawerClosedMessage struct {
	Item Pickable
}

type DrawerSelectionMessage struct {
	Project types.Project
}

func DrawerChoice(thing any) tea.Cmd {
	return func() tea.Msg {
		switch thing := thing.(type) {
		case types.Project:
			return DrawerSelectionMessage{Project: thing}
		default:
			return DrawerSelectionMessage{}
		}
	}
}

func DrawerClose() tea.Msg {
	return DrawerClosedMessage{}
}

func Select(item Pickable) tea.Cmd {
	return func() tea.Msg {
		return DrawerClosedMessage{
			Item: item,
		}
	}
}

func (m Model) View() string {
	sections := []string{}
	max := func(a, b int) int {
		if a > b {
			return a
		}
		return b
	}
	statusLine := m.statusLineStyle.Render("Quick Select Project")
	helpIndicator := m.helpIndicatorStyle.Render("Ctrl+? - Help")
	spacing := lipgloss.NewStyle().
		Background(lipgloss.Color("#232634")).
		Render(strings.Repeat(" ", max(0, m.ctx.ScreenWidth-lipgloss.Width(statusLine)-lipgloss.Width(helpIndicator))))
	for i, item := range m.items {
		content := item.Info()
		if i == 0 && m.index == 0 {
			sections = append(sections, m.topItemStyle.Render(m.selectedItemStyle.Render(content)))
		} else if i == 0 {
			sections = append(sections, m.topItemStyle.Render(content))
		} else if i == m.index {
			sections = append(sections, m.selectedItemStyle.Render(content))
		} else {
			// I need a default item style
			sections = append(sections, content)
		}
	}
	sections = append(sections, statusLine+spacing+helpIndicator)
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}
