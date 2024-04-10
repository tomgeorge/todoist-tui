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

type Component interface {
	ComponentName() string
	Model() tea.Model
	Resize(width, height int) tea.Model
}

type FooterInformation interface {
	FooterInformation() FooterSections
}

type HasHelp interface {
	KeyMap() help.KeyMap
}

type FooterSections struct {
	Location   string
	Dirty      string
	Pagination string
}
