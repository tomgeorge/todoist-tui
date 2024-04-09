package footer

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
)

type Model struct {
	ctx              ctx.Context
	showExtendedHelp bool
}

func New(ctx ctx.Context) Model {
	return Model{
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
	m.showExtendedHelp = showExtendedHelp
}

func (m Model) ShowExtendedHelp() bool {
	return m.showExtendedHelp
}

func (m Model) View() string {
	var sb strings.Builder
	mode := lipgloss.
		NewStyle().
		Background(lipgloss.Color("#8caaee")).
		Foreground(lipgloss.Color("#232634")).
		PaddingLeft(1).
		PaddingRight(1).
		Bold(true).
		Render("PROJECT VIEW")

	modeline := lipgloss.NewStyle().
		Background(lipgloss.Color("#232634")).Width(m.ctx.ScreenWidth - len(mode)).Render("             ")
	sb.WriteString(mode + modeline)
	sb.WriteString("\n")

	if m.showExtendedHelp {
		for i := 0; i <= 8; i++ {
			sb.WriteString("Hello!\n")
		}
	}
	return sb.String()
}
