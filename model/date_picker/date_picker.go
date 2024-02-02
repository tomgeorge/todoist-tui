package date_picker

import (
	"log"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type keyMap struct {
	MoveLeft                     key.Binding
	MoveRight                    key.Binding
	Increment                    key.Binding
	Decrement                    key.Binding
	SwitchInput                  key.Binding
	NewNaturalLanguageTodoOrEdit key.Binding
	NewAbsoluteTodo              key.Binding
	ShowHoursAndMinutes          key.Binding
	StopEditingOrDisable         key.Binding
	Help                         key.Binding
}

var defaultKeyMap = keyMap{
	MoveLeft: key.NewBinding(
		key.WithKeys("h"),
		key.WithHelp("l", "move right"),
	),
	MoveRight: key.NewBinding(
		key.WithKeys("l"),
		key.WithHelp("h", "move left"),
	),
	Increment: key.NewBinding(
		key.WithKeys("j"),
		key.WithHelp("j", "increment section"),
	),
	Decrement: key.NewBinding(
		key.WithKeys("k"),
		key.WithHelp("k", "decrement section"),
	),
	ShowHoursAndMinutes: key.NewBinding(
		key.WithKeys("H"),
		key.WithHelp("H", "set time"),
	),
	SwitchInput: key.NewBinding(
		key.WithKeys("ctrl+l"),
		key.WithHelp("ctrl+l", "toggle between human-defined date and absolute date"),
	),
	NewNaturalLanguageTodoOrEdit: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "edit due date with natural language"),
	),
	NewAbsoluteTodo: key.NewBinding(
		key.WithKeys("d"),
		key.WithHelp("d", "create due date with absolute date"),
	),
	StopEditingOrDisable: key.NewBinding(
		key.WithKeys("esc"),
		key.WithKeys("esc", "Stop editing"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "show help"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.MoveRight, k.MoveLeft, k.Increment, k.Decrement, k.ShowHoursAndMinutes, k.NewAbsoluteTodo, k.NewNaturalLanguageTodoOrEdit, k.StopEditingOrDisable, k.SwitchInput},
		{},
	}
}

func (m Model) Help() help.Model {
	return m.help
}

func (m Model) HelpKeys() help.KeyMap {
	return m.keys
}

type Model struct {
	editing                bool
	hasDueDate             bool
	focused                bool
	absolute               bool
	hasTime                bool
	focusedLabelStyle      lipgloss.Style
	labelStyle             lipgloss.Style
	focusedTextStyle       lipgloss.Style
	textStyle              lipgloss.Style
	label                  string
	absoluteDueDate        time.Time
	naturalLanguageDueDate textinput.Model
	focusIndex             int
	help                   help.Model
	showHelpUnderComponent bool
	keys                   keyMap
}

type ModelOption func(*Model)

func NewModel(opts ...ModelOption) *Model {
	const (
		defaultFocused                = false
		defaultEditing                = false
		defaultLabel                  = "Date"
		defaultAbsolute               = true
		defaultHasDueDate             = false
		defaultHasTime                = false
		defaultShowHelpUnderComponent = true
	)

	defaultLabelStyle := lipgloss.NewStyle().
		Underline(true).
		MarginBottom(1)

	defaultFocusedLabelStyle := defaultLabelStyle.
		Copy().
		Italic(true).
		Bold(true)

	defaultFocusedTextStyle := lipgloss.NewStyle().
		Background(lipgloss.Color("#626880"))

	defaultAbsoluteDate := time.Now().Local()

	defaultTextStyle := lipgloss.NewStyle()
	model := &Model{
		focused:                defaultFocused,
		editing:                defaultEditing,
		absolute:               defaultAbsolute,
		hasDueDate:             defaultHasDueDate,
		hasTime:                defaultHasTime,
		focusedLabelStyle:      defaultFocusedLabelStyle,
		focusedTextStyle:       defaultFocusedTextStyle,
		labelStyle:             defaultLabelStyle,
		textStyle:              defaultTextStyle,
		label:                  defaultLabel,
		absoluteDueDate:        defaultAbsoluteDate,
		naturalLanguageDueDate: textinput.New(),
		focusIndex:             0,
		help:                   help.New(),
		keys:                   defaultKeyMap,
		showHelpUnderComponent: defaultShowHelpUnderComponent,
	}

	for _, opt := range opts {
		opt(model)
	}
	return model
}

func WithLabel(label string) ModelOption {
	return func(m *Model) {
		m.label = label
	}
}

func WithDueDate(hasDueDate bool) ModelOption {
	return func(m *Model) {
		m.hasDueDate = hasDueDate
	}
}

func WithAbsoluteDueDate(dueDate time.Time) ModelOption {
	return func(m *Model) {
		m.hasDueDate = true
		m.absolute = true
		m.absoluteDueDate = dueDate
	}
}

func WithNaturalLanguageDueDate(humanDueDate string) ModelOption {
	return func(m *Model) {
		m.hasDueDate = true
		m.absolute = false
		m.naturalLanguageDueDate.SetValue(humanDueDate)
	}
}

func WithLabelStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.labelStyle = labelStyle
	}
}

func WithFocusedLabelStyle(focusedLabelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedLabelStyle = focusedLabelStyle
	}
}

func WithTextStyle(textStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.textStyle = textStyle
		m.naturalLanguageDueDate.TextStyle = textStyle
	}
}

func WithFocusedTextStyle(focusedTextStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedTextStyle = focusedTextStyle
	}
}

func WithShowHelpUnderComponent(showHelpUnderComponent bool) ModelOption {
	return func(m *Model) {
		m.showHelpUnderComponent = showHelpUnderComponent
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	log.Printf("DatePicker - Update - absolute %v\n", m.absoluteDueDate)
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.SwitchInput):
			m.setAbsolute(!m.absolute)
		case m.absolute && key.Matches(msg, m.keys.MoveLeft):
			m.moveLeft()
		case m.absolute && key.Matches(msg, m.keys.MoveRight):
			m.moveRight()
		case m.absolute && key.Matches(msg, m.keys.Increment):
			m.increment()
		case m.absolute && key.Matches(msg, m.keys.Decrement):
			m.decrement()
		case m.absolute && m.hasDueDate && key.Matches(msg, m.keys.ShowHoursAndMinutes):
			m.setHasTime(!m.hasTime)
		case !m.absolute && m.hasDueDate && !m.naturalLanguageDueDate.Focused() && key.Matches(msg, m.keys.NewNaturalLanguageTodoOrEdit):
			m.naturalLanguageDueDate.Focus()
			m.setEscapeBehavior()
		case !m.hasDueDate && key.Matches(msg, m.keys.NewNaturalLanguageTodoOrEdit):
			log.Println("no due date and got an A")
			m.hasDueDate = true
			m.naturalLanguageDueDate.Focus()
			m.setAbsolute(false)
		case !m.hasDueDate && key.Matches(msg, m.keys.NewAbsoluteTodo):
			m.hasDueDate = true
			m.setAbsolute(true)

		case key.Matches(msg, m.keys.StopEditingOrDisable):
			if !m.absolute && m.naturalLanguageDueDate.Focused() {
				m.naturalLanguageDueDate.Blur()
				m.updateNavigationKeys()
			} else {
				m.hasDueDate = false
				m.updateNavigationKeys()
			}
		case msg.String() == "ctrl+c":
			return m, tea.Quit
		default:
			if !m.absolute {
				var cmd tea.Cmd
				m.naturalLanguageDueDate, cmd = m.naturalLanguageDueDate.Update(msg)
				return m, cmd
			}
		}
	}
	return m, nil
}

func (m *Model) moveLeft() {
	if m.focusIndex != 0 {
		m.focusIndex--
	}
}

func (m *Model) moveRight() {
	switch {
	case !m.hasTime:
		if m.focusIndex != int(year) {
			m.focusIndex++
		}
	default:
		if m.focusIndex != int(lastSection) {
			m.focusIndex++
		}
	}
}

func (m *Model) increment() {
	switch m.focusIndex {
	case int(year):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(1, 0, 0)
	case int(month):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(0, 1, 0)
	case int(day), int(dayOfWeek):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(0, 0, 1)
	case int(hour):
		m.absoluteDueDate = m.absoluteDueDate.Add(1 * time.Hour)
	case int(minute):
		m.absoluteDueDate = m.absoluteDueDate.Add(1 * time.Minute)
	}
}

func (m *Model) decrement() {
	switch m.focusIndex {
	case int(year):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(-1, 0, 0)
	case int(month):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(0, -1, 0)
	case int(day), int(dayOfWeek):
		m.absoluteDueDate = m.absoluteDueDate.AddDate(0, 0, -1)
	case int(hour):
		m.absoluteDueDate = m.absoluteDueDate.Add(-1 * time.Hour)
	case int(minute):
		m.absoluteDueDate = m.absoluteDueDate.Add(-1 * time.Minute)
	}
}

func (m *Model) setHasTime(hasTime bool) {
	m.hasTime = hasTime
	if !m.hasTime && m.focusIndex > int(year) {
		m.focusIndex = int(year)
	}
}

func (m *Model) updateNavigationKeys() {
	switch {
	case !m.hasDueDate:
		log.Println("No due date")
		m.keys.MoveLeft.SetEnabled(false)
		m.keys.MoveRight.SetEnabled(false)
		m.keys.Increment.SetEnabled(false)
		m.keys.Decrement.SetEnabled(false)
		m.keys.StopEditingOrDisable.SetEnabled(false)
		m.keys.NewNaturalLanguageTodoOrEdit.SetEnabled(true)
		log.Printf("Setting newabsolutetodo to enabled")
		m.keys.NewAbsoluteTodo.SetEnabled(true)
		m.keys.NewNaturalLanguageTodoOrEdit.SetHelp("a", "set due date with natural language")
		m.keys.ShowHoursAndMinutes.SetEnabled(false)
	case m.absolute:
		log.Println("enabling absolute keys")
		m.keys.MoveLeft.SetEnabled(true)
		m.keys.MoveRight.SetEnabled(true)
		m.keys.Increment.SetEnabled(true)
		m.keys.Decrement.SetEnabled(true)
		m.keys.ShowHoursAndMinutes.SetEnabled(true)
		m.keys.StopEditingOrDisable.SetEnabled(true)
		m.keys.NewNaturalLanguageTodoOrEdit.SetEnabled(false)
		log.Printf("Setting newabsolutetodo to disabled")
		m.keys.NewAbsoluteTodo.SetEnabled(false)
	case !m.absolute:
		log.Println("enabling human keys")
		m.keys.MoveLeft.SetEnabled(false)
		m.keys.MoveRight.SetEnabled(false)
		m.keys.Increment.SetEnabled(false)
		m.keys.Decrement.SetEnabled(false)
		m.keys.ShowHoursAndMinutes.SetEnabled(false)
		m.keys.StopEditingOrDisable.SetEnabled(true)
		m.keys.NewNaturalLanguageTodoOrEdit.SetEnabled(true)
		log.Printf("Setting newabsolutetodo to disabled")
		m.keys.NewAbsoluteTodo.SetEnabled(false)
		m.keys.NewNaturalLanguageTodoOrEdit.SetHelp("a", "edit due date using natural language")
	}
	m.setEscapeBehavior()
}

func (m *Model) setEscapeBehavior() {
	switch {
	case !m.absolute && m.naturalLanguageDueDate.Focused():
		log.Printf("Setting escape to stop edit")
		m.keys.StopEditingOrDisable.SetHelp("esc", "stop editing")
	case !m.absolute && !m.naturalLanguageDueDate.Focused():
		m.keys.StopEditingOrDisable.SetHelp("esc", "remove due date")
	case m.absolute:
		m.keys.StopEditingOrDisable.SetHelp("esc", "remove due date")
	}
}

func (m *Model) setAbsolute(absolute bool) {
	m.absolute = absolute
	m.updateNavigationKeys()
}

func (m *Model) SetHelp(help bool) {
	m.help.ShowAll = help
}

type section int

const (
	dayOfWeek section = iota
	month
	day
	year
	hour
	minute
	lastSection = minute
)

// Expects dates to be in the format of
// Mon January 2 2006 3:04 PM
func makeArrayOfSections(date string) []string {
	split := strings.Split(date, " ")
	sections := []string{}
	for _, section := range split {
		if strings.Contains(section, ":") {
			sections = append(sections, strings.Split(section, ":")...)
		} else {
			sections = append(sections, section)
		}
	}
	return sections
}

func (m Model) renderLabel() string {
	if m.focused {
		return m.focusedLabelStyle.Render(m.label)
	} else {
		return m.labelStyle.Render(m.label)
	}
}
func (m Model) renderAbsoluteDate(sections []string) string {
	if m.absolute {
		var sb strings.Builder
		for i, section := range sections {
			if m.focusIndex == i {
				sb.WriteString(m.focusedTextStyle.Render(section))
			} else {
				sb.WriteString(m.textStyle.Render(section))
			}
			if i != len(sections) {
				sb.WriteString(" ")
			}
		}
		return sb.String()
	}
	return ""
}

func (m Model) renderNaturalLanguageDueDate() string {
	switch {
	case m.absolute:
		return ""
	case !m.absolute && m.focused && m.naturalLanguageDueDate.Focused():
		return lipgloss.JoinVertical(lipgloss.Left,
			m.naturalLanguageDueDate.View(),
			"WARNING: This doesn't do any checking of the contents of this due date.",
			"Please make sure you're typing something that todoist can understand here.",
			"For more information, see https://todoist.com/help/articles/introduction-to-due-dates-and-due-times-q7VobO",
		)
	case !m.absolute && m.focused && !m.naturalLanguageDueDate.Focused():
		return lipgloss.JoinVertical(lipgloss.Left,
			"Press 'a' to add a due date, or 'esc' to cancel")
	default:
		return m.naturalLanguageDueDate.Value()
	}
}

func (m Model) renderHelp() string {
	if m.help.ShowAll && m.showHelpUnderComponent {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.help.View(m.keys))
	}
	return ""
}

func (m *Model) FocusOn() {
	m.focused = true
	m.setAbsolute(m.absolute)
}

func (m *Model) FocusOff() {
	m.focused = false
}

func (m Model) View() string {

	var sections []string
	if m.hasTime {
		sections = makeArrayOfSections(m.absoluteDueDate.Format("Mon January 2 2006 3:04 PM"))
	} else {
		sections = makeArrayOfSections(m.absoluteDueDate.Format("Mon January 2 2006"))
	}

	label := m.renderLabel()
	dateSection := m.renderAbsoluteDate(sections)
	naturalLanguageDueDateSection := m.renderNaturalLanguageDueDate()
	help := m.renderHelp()
	if !m.hasDueDate {
		if m.focused {
			return lipgloss.JoinVertical(lipgloss.Left,
				label,
				"No due date",
				"Press 'a' to set a natural language due date",
				"Press 'd' to set an absolute due date",
				help)

		} else {
			return lipgloss.JoinVertical(lipgloss.Left,
				label,
				"No due date")
		}
	}
	if m.absolute {
		return lipgloss.JoinVertical(lipgloss.Left,
			label,
			dateSection,
			help)
	} else {
		return lipgloss.JoinVertical(lipgloss.Left,
			label,
			naturalLanguageDueDateSection,
			help)
	}
}

type DueDateContent struct {
	HasDueDate             bool
	AbsoluteDate           time.Time
	IncludeHoursAndMinutes bool
	HumanInputDate         string
}

func (m Model) GetContent() DueDateContent {
	// Natural language wins
	switch {
	case !m.hasDueDate:
		return DueDateContent{HasDueDate: false}
	case m.absolute && m.hasTime:
		return DueDateContent{HasDueDate: true, AbsoluteDate: m.absoluteDueDate, IncludeHoursAndMinutes: true}
	case m.absolute && !m.hasTime:
		return DueDateContent{HasDueDate: true, AbsoluteDate: m.absoluteDueDate, IncludeHoursAndMinutes: false}
	default:
		return DueDateContent{HasDueDate: true, HumanInputDate: m.naturalLanguageDueDate.Value()}
	}
}

func (m *Model) SetAbsoluteDueDate(dueDate time.Time) {
	m.hasDueDate = true
	m.absolute = true
	// FIXME: Can we assume that an absolute due date here will have hours and
	// minutes set to it?
	m.hasTime = true
	m.absoluteDueDate = dueDate
}

func (m *Model) SetNaturalLanguageDueDate(dueString string) {
	m.hasDueDate = true
	m.absolute = false
	m.naturalLanguageDueDate.Focus()
	m.naturalLanguageDueDate.SetValue(dueString)
}
