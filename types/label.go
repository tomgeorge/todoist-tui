package types

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// A todoist label
type Label struct {
	// The ID of the label
	Id string `json:"id"`
	// The name of the label
	Name string `json:"name"`
	// The color of the label icon. Refer to the name column in the todoist Colors
	// guide for more info
	Color string `json:"color"`
	// Label's order in the label list, where the smallest value should place the
	// label at the top
	ItemOrder int `json:"item_order"`
	// Whether the label is marked as deleted
	IsDeleted bool `json:"is_deleted"`
	// Whether the label is a favorite
	IsFavorite bool `json:"is_favorite"`
}

func (l Label) Render() string {
	return l.Name
}

func (l Label) Style() lipgloss.Style {
	return lipgloss.NewStyle().
		Background(lipgloss.Color(Colors[l.Color])).
		Foreground(lipgloss.AdaptiveColor{Dark: "#ffffff", Light: "#000000"}).
		MarginRight(1)
	// MarginTop(1)
}

func (l Label) Display() string {
	return l.Name
}

func (l Label) SearchResultString() string {
	return fmt.Sprintf("%s", l.Name)
}

func (l Label) Implemented() bool {
	return false
}

func (l Label) Kind() string {
	return "Label"
}

func (l Label) IsMatch(name string) bool {
	return l.Name == name
}
