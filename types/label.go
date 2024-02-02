package types

import (
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
		MarginRight(1).
		MarginTop(1)
}
