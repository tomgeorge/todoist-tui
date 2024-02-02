package events

import (
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
	"github.com/tomgeorge/todoist-tui/ctx"
)

type Model struct {
	Events []Event
	Ctx    ctx.Context
}

type Event struct {
	Timer   timer.Model
	Timeout bool
	Message string
	Style   lipgloss.Style
}

type NewMessage struct {
	Timeout  bool
	Duration time.Duration
	Message  string
	Style    lipgloss.Style
}

func (m *Model) Publish(message string, style lipgloss.Style, timeout bool, duration time.Duration) tea.Cmd {
	return func() tea.Msg {
		return NewMessage{timeout, duration, message, style}
	}
}

func New(ctx ctx.Context) *Model {
	return &Model{
		Events: []Event{},
		Ctx:    ctx,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case NewMessage:
		log.Printf("events - NewMessage timeout %v", msg.Duration.Seconds())
		timer := timer.New(msg.Duration)
		event := Event{
			Message: msg.Message,
			Style:   msg.Style,
			Timer:   timer,
		}
		m.Events = append(m.Events, event)
		if event.Timeout {
			cmd := m.Events[len(m.Events)-1].Timer.Init()
			cmds = append(cmds, cmd)
		}
	case timer.TickMsg:
		for i, event := range m.Events {
			if event.Timer.ID() == msg.ID {
				updated, cmd := event.Timer.Update(msg)
				m.Events[i].Timer = updated
				cmds = append(cmds, cmd)
			}
		}
	case timer.TimeoutMsg:
		m.Ctx.Logger.Info("timeout")
		m.Events = lo.Filter(m.Events, func(e Event, _ int) bool {
			return e.Timer.ID() != msg.ID
		})
		return m, nil
	}
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if len(m.Events) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, event := range m.Events {
		sb.WriteString(event.Style.Render(event.Message))
	}
	return sb.String()
}
