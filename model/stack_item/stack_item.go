package stack_item

import (
	"github.com/charmbracelet/bubbles/help"
	tea "github.com/charmbracelet/bubbletea"
)

type StackItem interface {
	Component
	FooterInformation
	tea.Model
	HasHelp
}

// TODO: Document fields
type Component interface {
	ComponentName() string
	Model() tea.Model
	Resize(width, height int) tea.Model
}

// TODO: Document fields
type FooterInformation interface {
	FooterInformation() FooterSections
}

// TODO: Document fields
type HasHelp interface {
	KeyMap() help.KeyMap
}

// TODO: Document fields
type FooterSections struct {
	Location   string
	Dirty      string
	Pagination string
}
