package picker

import (
	"fmt"
	"log"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

/* Keymaps and help */
type keyMap struct {
	New            key.Binding
	Confirm        key.Binding
	NextSuggestion key.Binding
	CancelInput    key.Binding
	Help           key.Binding
	ShowAvailable  key.Binding
}

var defaultKeyMap = keyMap{
	New: key.NewBinding(
		key.WithKeys("a"),
		key.WithHelp("a", "add item"),
	),
	Confirm: key.NewBinding(
		key.WithKeys("enter", "confirm"),
		key.WithHelp("enter", "select/deselect"),
	),
	CancelInput: key.NewBinding(
		key.WithKeys("esc", "cancel input"),
		key.WithHelp("esc", "cancel input"),
	),
	NextSuggestion: key.NewBinding(
		key.WithKeys("ctrl+n", "next suggestion"),
		key.WithHelp("ctrl+n", "next suggestion"),
	),
	ShowAvailable: key.NewBinding(
		key.WithKeys("ctrl+l", "show available items"),
		key.WithHelp("ctrl+l", "show available items"),
	),
	Help: key.NewBinding(
		key.WithKeys("?", "help"),
		key.WithHelp("?", "help"),
	),
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Help}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.New, k.Confirm, k.CancelInput, k.ShowAvailable},
		{k.Help},
	}
}

func (m Model) Help() help.Model {
	return m.help
}

func (m Model) HelpKeys() help.KeyMap {
	return m.keys
}

/* PickerItem interface */
type PickerItem interface {
	Render() string
	Style() lipgloss.Style
}

// This is a possible way to genericize some of the picker item stuff and still
// have the underlying type abvailable. You wouldn't have to coerce it to the
// interface type all the time which might be useful. The picker could also
// maybe have access to the underlying type if it needs it, but I don't know if
// it does yet.
//
// The interface would describe the behavior, and the generic item type would
// maybe make it's usage easier.
// type GenericPickerItem[T PickerItem] struct {
// 	Underlying T
// }
//
// func (a *GenericPickerItem[T]) Render() string {
// 	return a.Underlying.Render()
// }
//
// func (a *GenericPickerItem[T]) Style() lipgloss.Style {
// 	return a.Underlying.Style()
// }
//
// func NewList[T PickerItem](elems ...T) []GenericPickerItem[T] {
// 	items := make([]GenericPickerItem[T], len(elems))
// 	for i, e := range elems {
// 		items[i] = GenericPickerItem[T]{Underlying: e}
// 	}
// 	return items
// }
//
// func TryWithLabels() {
// 	list := NewList[types.Label]()
// }

// Construct a new picker item list
func NewPickerItem(elements []interface{}) []PickerItem {
	items := make([]PickerItem, len(elements))
	for i, item := range items {
		elements[i] = item
	}
	return items
}

/* Model, constructor, and options */

type Model struct {
	focused           bool
	textInput         textinput.Model
	requiredSelection int
	textInputVisible  bool
	keys              keyMap
	items             []PickerItem
	selectedItems     []PickerItem
	showAvailable     bool
	label             string
	labelStyle        lipgloss.Style
	focusedLabelStyle lipgloss.Style
	validationStyle   lipgloss.Style
	multipleSelection bool
	help              help.Model
}

type ModelOption func(*Model)

func InitTextInput(items []PickerItem) textinput.Model {
	log.Printf("items are %v", items)
	suggestions := make([]string, len(items))
	for i, item := range items {
		suggestions[i] = item.Render()
	}
	textInput := textinput.New()
	textInput.ShowSuggestions = true
	textInput.SetSuggestions(suggestions)
	return textInput
}

func (m *Model) UpdateSuggestions(items []PickerItem) {
	suggestions := make([]string, len(items))
	for i, item := range items {
		suggestions[i] = item.Render()
	}
	m.textInput.SetSuggestions(suggestions)
}

func NewModel(opts ...ModelOption) *Model {
	const (
		defaultLabel             = "Items"
		defaultMultipleSelection = true
		defaultRequiredSelection = 0
	)
	defaultLabelStyle := lipgloss.NewStyle().
		Underline(true).
		MarginBottom(1)

	defaultFocusedLabelStyle := defaultLabelStyle.
		Copy().
		Bold(true).
		Italic(true)

	defaultValidationStyle := defaultFocusedLabelStyle.Copy()

	defaultItems := []PickerItem{}

	model := &Model{
		items:             defaultItems,
		label:             defaultLabel,
		labelStyle:        defaultLabelStyle,
		focusedLabelStyle: defaultFocusedLabelStyle,
		validationStyle:   defaultValidationStyle,
		help:              help.New(),
		textInput:         InitTextInput(defaultItems),
		textInputVisible:  false,
		multipleSelection: defaultMultipleSelection,
		keys:              defaultKeyMap,
		showAvailable:     false,
	}
	for _, opt := range opts {
		opt(model)
	}
	return model
}

func WithItems(items []PickerItem) ModelOption {
	return func(m *Model) {
		m.items = items
		m.UpdateSuggestions(items)
	}
}

func WithSelected(selectedItems []PickerItem) ModelOption {
	return func(m *Model) {
		m.selectedItems = selectedItems
	}
}

func WithLabel(label string) ModelOption {
	return func(m *Model) {
		m.label = label
	}
}

func WithLabelStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.labelStyle = labelStyle
	}
}

func WithFocusedLabelStyle(labelStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.focusedLabelStyle = labelStyle
	}
}

func WithValidationStyle(validationStyle lipgloss.Style) ModelOption {
	return func(m *Model) {
		m.validationStyle = validationStyle
	}
}

func WithPlaceholder(placeholder string) ModelOption {
	return func(m *Model) {
		m.textInput.Placeholder = placeholder
	}
}

func WithMultipleSelection(multiSelect bool) ModelOption {
	return func(m *Model) {
		m.multipleSelection = multiSelect
	}
}

func WithRequiredSelection(requiredSelection int) ModelOption {
	return func(m *Model) {
		m.requiredSelection = requiredSelection
	}
}

/* Lifecycle methods */

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	log.Printf("In picker update")
	var cmd tea.Cmd
	var cmds []tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keys.Help):
			m.help.ShowAll = !m.help.ShowAll
		case key.Matches(msg, m.keys.New):
			if !m.textInput.Focused() {
				log.Printf("Got the a key")
				m.textInputVisible = true
				m.textInput.Focus()
				return m, nil
			}
		case key.Matches(msg, m.keys.CancelInput):
			if m.textInputVisible && m.textInput.Focused() {
				log.Printf("Got cancel input command, unfocusing")
				m.textInput.SetValue("")
				m.textInputVisible = false
				m.textInput.Blur()
				return m, nil
			}
		case key.Matches(msg, m.keys.Confirm):
			value := m.textInput.Value()
			log.Printf("Confirm key %s", value)
			// If the value matches something
			if containsRenderedItem(m.items, value) {
				log.Printf("items contains value")
				if containsRenderedItem(m.selectedItems, value) {
					m.Deselect(value)
				} else {
					m.Select(value)
				}
				m.textInput.SetValue("")
			}
		case key.Matches(msg, m.keys.ShowAvailable):
			m.showAvailable = !m.showAvailable
		}
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		}

	}
	if m.textInput.Focused() {
		log.Println("Text input is focused, running update")
		log.Printf("Message is %v", msg)
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)
	}
	log.Printf("I'm returning from update")
	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	sections := []string{}
	if m.focused {
		sections = append(sections, m.focusedLabelStyle.Render(m.label))
	} else {
		if m.requiredSelection > 0 && len(m.selectedItems) != m.requiredSelection {
			sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Left, m.labelStyle.Render(m.label), " ", m.validationStyle.Render(" 1")))
		} else {
			sections = append(sections, m.labelStyle.Render(m.label))
		}
	}
	sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Left, m.styledItems("Selected", m.selectedItems)...))
	if m.textInputVisible && len(m.items) > 0 {
		sections = append(sections, m.textInput.View())
	}
	if m.showAvailable {
		sections = append(sections, lipgloss.NewStyle().Bold(true).MarginTop(1).Render("Available:"))
		sections = append(sections, lipgloss.JoinHorizontal(lipgloss.Left, m.styledItems(m.label, m.items)...))
	}
	if m.focused && m.requiredSelection > 0 && len(m.selectedItems) != m.requiredSelection {
		sections = append(sections, m.validationStyle.Render(fmt.Sprintf("This section requires %d selection(s)", m.requiredSelection)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, sections...)
}

// Select the item in model.items indicated by value
// and add to model.selectedItems
func (m *Model) Select(value string) {
	item := m.getItemByValue(value)
	if m.multipleSelection {
		m.selectedItems = append(m.selectedItems, item)
	} else {
		m.selectedItems = []PickerItem{item}
	}
}

// Deselect the item in model.items indicated by value
// and remove from model.selectedItems
func (m *Model) Deselect(value string) {
	item := m.getItemByValue(value)
	m.selectedItems = removeFromSlice(m.selectedItems, item)
}

/* List manipulation */

// Return the PickerItem represented by value
func (m Model) getItemByValue(value string) PickerItem {
	var item PickerItem
	for _, element := range m.items {
		if element.Render() == value {
			item = element
		}
	}
	return item
}

func removeFromSlice(items []PickerItem, item PickerItem) []PickerItem {
	for i, element := range items {
		if element.Render() == item.Render() {
			copy(items[i:], items[i+1:])
			items[len(items)-1] = nil
			items = items[:len(items)-1]
			break
		}
	}
	return items
}

func containsRenderedItem(items []PickerItem, item string) bool {
	found := false
	for _, element := range items {
		if element.Render() == item {
			found = true
		}
	}
	return found
}

func (m Model) styledItems(label string, items []PickerItem) []string {
	content := []string{}
	if len(items) == 0 {
		return append(content, "None")
	}
	// content = append(content, fmt.Sprintf("%s: ", label))
	for _, item := range items {
		content = append(content, item.Style().Render(item.Render()))
	}
	return content
}

// Don't do anything when focused
func (m *Model) FocusOn() {
	m.focused = true
}

func (m *Model) FocusOff() {
	m.focused = false
	if m.textInput.Focused() {
		m.textInputVisible = false
		m.textInput.Blur()
	}
}

func (m *Model) GetSelectedItems() []PickerItem {
	return m.selectedItems
}

func (m *Model) SetItems(items []PickerItem) {
	m.items = items
	if m.textInput.ShowSuggestions {
		m.UpdateSuggestions(items)
	}
}

func (m *Model) SetSelected(items []PickerItem) {
	m.selectedItems = items
}

func (m *Model) SetHelp(help bool) {
	m.help.ShowAll = help
}