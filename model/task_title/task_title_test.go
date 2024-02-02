package task_title_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	title "github.com/tomgeorge/todoist-tui/model/task_title"
)

func TestIsUnfocusedByDefault(t *testing.T) {
	model := title.New()
	if model.Focused() {
		t.Errorf("Should be unfocused by default")
	}
}

func TestFocus(t *testing.T) {
	model := title.New()
	m := *model
	if m.Focused() {
		t.Errorf("Should be unfocused by default")
	}

	m.FocusOn()

	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	if !m.Focused() && !m.Editing() {
		t.Errorf("Expected to be focused and editing the textinput")
	}
	if cmd != nil {
		t.Errorf("Did not expect a command")
	}
}

func TestEditing(t *testing.T) {
	model := title.New(title.WithContent("foo"))
	m := *model
	m.FocusOn()
	m.SetEditing(true)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if got := m.GetContent(); got != "foob" {
		t.Fatalf("Expected foob but got %s", got)
	}
}

func TestFocusToggles(t *testing.T) {
	model := title.New()
	model.FocusOn()
	if !model.Focused() {
		t.Fatal("Should be focused")
	}
	model.FocusOff()
	if model.Focused() {
		t.Fatal("Should be unfocused")
	}

}
