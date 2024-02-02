package events

import (
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/samber/lo"
)

type Model struct {
	Events []Event
}

type Event struct {
	Timer   timer.Model
	Message string
	Style   lipgloss.Style
	Quit    bool
}

type NewMessage struct {
	Timeout time.Duration
	Message string
	Style   lipgloss.Style
	Quit    bool
}

func (m *Model) Publish(message string, style lipgloss.Style, timeout time.Duration, quit bool) tea.Cmd {
	return func() tea.Msg {
		return NewMessage{timeout, message, style, quit}
	}
}

func New() *Model {
	return &Model{
		Events: []Event{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case NewMessage:
		log.Printf("events - NewMessage timeout %v", msg.Timeout.Seconds())
		timer := timer.New(msg.Timeout)
		event := Event{
			Message: msg.Message,
			Style:   msg.Style,
			Timer:   timer,
			Quit:    msg.Quit,
		}
		m.Events = append(m.Events, event)
		cmd := m.Events[len(m.Events)-1].Timer.Init()
		cmds = append(cmds, cmd)
	case timer.TickMsg:
		for i, event := range m.Events {
			if event.Timer.ID() == msg.ID {
				updated, cmd := event.Timer.Update(msg)
				m.Events[i].Timer = updated
				cmds = append(cmds, cmd)
			}
		}
	case timer.TimeoutMsg:
		m.Events = lo.Filter(m.Events, func(e Event, _ int) bool {
			return e.Timer.ID() != msg.ID
		})
		return m, nil
	}
	quit := lo.Filter(m.Events, func(e Event, _ int) bool {
		return e.Quit == true
	})
	if len(quit) > 0 {
		return m, tea.Quit
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
