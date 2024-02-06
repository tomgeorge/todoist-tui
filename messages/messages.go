package messages

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tomgeorge/todoist-tui/services/sync"
)

// PushMessage pushes a new model onto the stack
type PushMessage struct {
	Id    string
	Model tea.Model
}

// Return a tea.Cmd that asks the root model to push a model onto the stack during
// Update
func Push(id string, model tea.Model) tea.Cmd {
	return func() tea.Msg {
		return PushMessage{
			Id:    id,
			Model: model,
		}
	}
}

// PopMessage pops the top model off the stack
type PopMessage struct{}

// Return a tea.Cmd that asks the root model to pop the top model off of the
// stack during Update
func Pop() tea.Cmd {
	return func() tea.Msg {
		return PopMessage{}
	}
}

// A StateMessage is sent back from the client
type StateMessage struct {
	State *sync.SyncResponse
	Err   error
}

// LoadingMessage is sent on init when loading the application
type LoadingMessage struct{}
