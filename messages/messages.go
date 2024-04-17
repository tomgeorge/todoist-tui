package messages

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/services/sync"
	"github.com/tomgeorge/todoist-tui/types"
	"github.com/tomgeorge/todoist-tui/util"
)

// PushMessage pushes a new model onto the stack
// TODO: Document fields
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
// TODO: Document fields
type PopMessage struct{}

// Return a tea.Cmd that asks the root model to pop the top model off of the
// stack during Update
func Pop() tea.Cmd {
	return func() tea.Msg {
		return PopMessage{}
	}
}

// A StateMessage is sent back from the client
// TODO: Document fields
type StateMessage struct {
	State *sync.SyncResponse
	Error error
}

func StateUpdated() tea.Msg {
	return StateUpdatedMessage{}
}

// The state was updated and the view should redraw itself
// TODO: Document fields
type StateUpdatedMessage struct{}

func TaskOperation(client *sync.Client, commands sync.CommandList) tea.Cmd {
	return func() tea.Msg {
		request := client.NewSyncRequest([]string{"items"}, commands)
		response, err := client.Sync(context.Background(), *request)
		return OperationResponse{
			State:    response,
			Commands: commands,
			Error:    err,
		}
	}
}

// TODO: Document fields
type OperationPendingMessage struct{}

func OperationPending() tea.Cmd {
	return func() tea.Msg {
		return OperationPendingMessage{}
	}
}

// TODO: Document fields
type OperationCompleteMessage struct{}

func OperationComplete() tea.Cmd {
	return func() tea.Msg {
		return OperationCompleteMessage{}
	}
}

// Returned when an update is sent to the server
// TODO: Document fields
type OperationResponse struct {
	State    *sync.SyncResponse
	Commands []sync.Command
	Error    error
}

// Sent after a successful StateMessage to write to the filesystem
// TODO: Document fields
type SaveStateMessage struct {
	Success bool
	Error   error
}

func SaveState(ctx ctx.Context, state *sync.SyncResponse) tea.Cmd {
	return func() tea.Msg {
		buf, err := util.PrettyMarshal(state)
		if err != nil {
			ctx.Logger.Debug("Error marshalling json: %v", err)
			return SaveStateMessage{
				Success: false,
				Error:   fmt.Errorf("marshalling sync response into json: %v", err),
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
// TODO: Document fields
type LoadingMessage struct{}

// CreateItemMessage is sent when a task was attempted to be created
// TODO: Document fields
type TaskCreatedMessage struct {
	Task  *types.Item
	Error error
}

func OpenDrawer() tea.Msg {
	return OpenDrawerMessage{}
}

// TODO: Document fields
type OpenDrawerMessage struct {
}

type ChangeProjectDirection string

const NextProject ChangeProjectDirection = "forward"
const PreviousProject ChangeProjectDirection = "backwards"

// Send to ask the UI to change the currently shown project, should only be sent
// from the project view component
// TODO: Document fields
type ChangeProjectMessage struct {
	From      *types.Project
	Direction ChangeProjectDirection
}
