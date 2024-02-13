package messages

import (
	"encoding/json"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
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
	Error error
}

// Returned when an update is sent to the server
type OperationResponse struct {
	State *sync.SyncResponse
	Error error
}

// Sent after a successful StateMessage to write to the filesystem
type SaveStateMessage struct {
	Success bool
	Error   error
}

func SaveState(ctx ctx.Context, state *sync.SyncResponse) tea.Cmd {
	return func() tea.Msg {
		buf, err := json.Marshal(state)
		ctx.Logger.Infof("in savestate, len(buf): %d", len(buf))
		if err != nil {
			return SaveStateMessage{
				Success: false,
				Error:   err,
			}
		}
		path := filepath.Join(ctx.Config.StateDir, "state.json")
		ctx.Logger.Infof("savestate, path: %s", path)
		err = os.WriteFile(path, buf, 0644)
		if err != nil {
			return SaveStateMessage{
				Success: false,
				Error:   err,
			}
		}
		return SaveStateMessage{
			Success: true,
			Error:   nil,
		}
	}
}

// LoadingMessage is sent on init when loading the application
type LoadingMessage struct{}

// CreateItemMessage is sent when a task was attempted to be created
type TaskCreatedMessage struct {
	Task  *types.Item
	Error error
}
