package drawer

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Pickable interface {
	Info() string
}
type Model struct {
	items   []Pickable
	focused bool
	enabled bool
}

func New(items []Pickable) Model {
	return Model{
		items:   items,
		focused: false,
		enabled: true,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m Model) View() string {
	sections := []string{}
	sections = append(sections, fmt.Sprintf("Hey, I'm the drawer, of length %d", len(m.items)))
	for i, item := range m.items {
		if i == 0 {
			sections = append(sections, lipgloss.NewStyle().Background(lipgloss.Color("1")).Render(item.Info()))
		} else {
			sections = append(sections, item.Info())
		}
	}
	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

func (m Model) Enabled() bool {
	return m.enabled
}
