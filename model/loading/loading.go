package loading

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tomgeorge/todoist-tui/ctx"
	"github.com/tomgeorge/todoist-tui/messages"
	"github.com/tomgeorge/todoist-tui/model/stack_item"
	"github.com/tomgeorge/todoist-tui/services/sync"
)

func (m Model) Resize(_, _ int) tea.Model { return m }

func (m Model) Model() tea.Model { return m }

func (m Model) ComponentName() string { return "loading" }

// TODO: Document fields
type Model struct {
	ctx          ctx.Context
	loading      bool
	spinner      spinner.Model
	errorMessage string
	errorStyle   lipgloss.Style
}

type ModelOption func(m *Model)

func New(ctx ctx.Context, opts ...ModelOption) Model {
	var (
		defaultLoading = true
		defaultSpinner = spinner.New(
			spinner.WithSpinner(spinner.Dot),
			spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("#939ab7"))),
		)
		defaultErrorStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#ed8796"))
	)
	m := &Model{
		ctx:          ctx,
		loading:      defaultLoading,
		spinner:      defaultSpinner,
		errorMessage: "",
		errorStyle:   defaultErrorStyle,
	}

	for _, opt := range opts {
		opt(m)
	}
	return *m
}

func WithSpinnerModel(spinnerModel spinner.Model) ModelOption {
	return func(m *Model) {
		m.spinner = spinnerModel
	}
}

func WithErrorStyle(errorStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.errorStyle = errorStyle
	}
}

func (m *Model) performSync() tea.Cmd {
	return func() tea.Msg {
		stateFile, err := os.ReadFile(filepath.Join(m.ctx.Config.StateDir, "state.json"))
		if err != nil && !os.IsNotExist(err) {
			return messages.StateMessage{State: nil, Error: err}
		}

		m.ctx.Logger.Debug("found a state file on the machine")
		// m.ctx.Logger.Debugf("found a state file on the machine:\n---\n%s\n---\n", util.PrettyJSON(stateFile))

		// If the state file has stuff init, we should try and merge the existing
		// state file with the response from the server
		// TODO: We could also check for a zero value of the SyncResponse struct,
		// e.g. if the file contains `{}` but that might throw an error when trying
		// to unmarshal
		// FIXME: Redact the user's api token from the state file on the machine
		if len(stateFile) != 0 {
			state := &sync.SyncResponse{}
			err := json.Unmarshal(stateFile, state)
			if err != nil {
				m.ctx.Logger.Debugf("Error unmarshalling: %v", err)
				return messages.StateMessage{State: nil, Error: err}
			}
			m.ctx.Logger.Debug("Found a state file, doing weird mergey shit now")
			recentUpdates, err := m.ctx.Client.FullSync(context.Background(), sync.WithSyncToken(state.SyncToken))
			if err != nil {
				m.ctx.Logger.Debugf("Full sync error: %v", err)
				return messages.StateMessage{State: state, Error: err}
			}
			return messages.StateMessage{State: sync.Merge(state, recentUpdates), Error: nil}
		}

		// If the state file was not found, do a full sync
		state, err := m.ctx.Client.FullSync(context.Background())
		m.ctx.Logger.Debugf("No state file sync: %v", err)
		return messages.StateMessage{State: state, Error: err}
	}
}

// FIXME: I'm not quite sure why this is needed, and why I can't return
// tea.Batch(performSync(), m.spinner.Tick) in Init(), but it doesn't work and
// I can't remember why at the moment
func load() tea.Cmd {
	return func() tea.Msg { return messages.LoadingMessage{} }
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(m.performSync(), load())
}

// TODO: Document fields
type LoadingCompleteMessage struct{}

func LoadingComplete() tea.Msg {
	return LoadingCompleteMessage{}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	switch msg := msg.(type) {
	case messages.LoadingMessage:
		m.ctx.Logger.Debug("LoadingMessage")
		return m, m.spinner.Tick
	case messages.StateMessage:
		if msg.Error != nil {
			m.ctx.Logger.Debug("Got an error %s", msg.Error.Error())
			m.errorMessage = fmt.Sprintf("Error getting data from todoist: %s", msg.Error.Error())
			m.loading = false
			return m, tea.Sequence(tea.Quit)
		}
		return m, tea.Batch(
			LoadingComplete,
			messages.SaveState(m.ctx, msg.State),
		)
	}
	return m, cmd
}

func (m Model) View() string {
	if m.loading {
		return lipgloss.NewStyle().PaddingLeft(1).Render(fmt.Sprintf("\n%s fetching data from todoist\n",
			m.spinner.View()))
	} else {
		return lipgloss.JoinVertical(lipgloss.Left, m.errorStyle.Render(m.errorMessage, "\n"))
	}
}

func (m Model) FooterInformation() stack_item.FooterSections {
	return stack_item.FooterSections{
		Location: fmt.Sprintf("%s:", m.ComponentName()),
	}
}

// TODO: Document fields
type keyMap struct {
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{}
}

func (m Model) KeyMap() help.KeyMap {
	return keyMap{}
}
