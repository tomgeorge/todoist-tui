package types

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type Priority int

// The todoist API does this kind of funny
// in that Priority is an integer but they
// correspond to the opposite numbers in the UI
const (
	P1 Priority = 4
	P2 Priority = 3
	P3 Priority = 2
	P4 Priority = 1
)

var Priorities = []Priority{
	P1,
	P2,
	P3,
	P4,
}

func (p Priority) Render() string {
	return fmt.Sprintf("P%d", p)
}

func (p Priority) Style() lipgloss.Style {
	return lipgloss.NewStyle().MarginRight(1)
}
