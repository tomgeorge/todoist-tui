package footer

import (
	"strings"

	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/model/stack_item"
)

type Model struct {
	help             help.Model
	ctx              ctx.Context
	showExtendedHelp bool
}

func New(ctx ctx.Context) Model {
	return Model{
		help.New(),
		ctx,
		false,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	return m, nil
}

func (m *Model) SetShowExtendedHelp(showExtendedHelp bool) {
	m.help.ShowAll = showExtendedHelp
}

func (m Model) ShowExtendedHelp() bool {
	return m.help.ShowAll
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// FIXME: This should be called statusline or something
// FIXME: Is there a better way to conditionally render things like pagination?
func (m Model) View(top stack_item.StackItem, screenWidth int) string {
	var sb strings.Builder
	dirty := top.FooterInformation().Dirty
	mode := lipgloss.
		NewStyle().
		Background(lipgloss.Color("#8caaee")).
		Foreground(lipgloss.Color("#232634")).
		PaddingLeft(1).
		PaddingRight(1).
		Bold(true).
		Render(top.FooterInformation().Location + dirty)

	// FIXME: Needs to be incorporated into the theme
	pagination := lipgloss.NewStyle().
		Background(lipgloss.Color("#f2d5cf")).
		Foreground(lipgloss.Color("#232634")).
		PaddingLeft(1).
		PaddingRight(1).
		Render(top.FooterInformation().Pagination)

	help := lipgloss.
		NewStyle().
		Background(lipgloss.Color("#8caaee")).
		Foreground(lipgloss.Color("#232634")).
		PaddingLeft(1).
		PaddingRight(1).
		Bold(true).
		Render("Ctrl+? - Help")

		// FIXME: This needs to be its own function
	spacerLength := func() int {
		modeWidth := lipgloss.Width(mode)
		helpWidth := lipgloss.Width(help)
		paginationWidth := lipgloss.Width(pagination)
		if paginationWidth == 2 {
			paginationWidth = 0
		}
		return max(0, screenWidth-modeWidth-helpWidth-paginationWidth)
	}
	spacing := strings.Repeat(" ", spacerLength())
	spacer := lipgloss.NewStyle().
		Background(lipgloss.Color("#232634")).Render(spacing)
	sb.WriteString(mode + spacer)
	if lipgloss.Width(pagination) > 2 {
		sb.WriteString(pagination)
	}
	sb.WriteString(help)
	sb.WriteString("\n")

	if m.help.ShowAll {
		sb.WriteString(m.help.View(top.KeyMap()))
	}
	return sb.String()
}
