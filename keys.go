package main

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up       key.Binding
	Down     key.Binding
	PageUp   key.Binding
	PageDown key.Binding
	Top      key.Binding
	Bottom   key.Binding
	Enter    key.Binding
	Quit     key.Binding
	Help     key.Binding
}

var Keys = keyMap{
	Up: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("'↑/k'", "move up"),
	),
	Down: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("'↓/j'", "move down"),
	),
	PageUp: key.NewBinding(
		key.WithKeys("ctrl+u"),
		key.WithHelp("ctrl+u", "[NOT IMPLEMENTED] page up"),
	),
	PageDown: key.NewBinding(
		key.WithKeys("ctrl+d"),
		key.WithHelp("ctrl+d", "[NOT IMPLEMENTED] page down"),
	),
	Top: key.NewBinding(
		key.WithKeys("g"),
		key.WithHelp("g", "top of list"),
	),
	Bottom: key.NewBinding(
		key.WithKeys("G"),
		key.WithHelp("G", "bottom of list"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q/esc/ctrl+c", "Quit"),
	),
  Help: key.NewBinding(
    key.WithKeys("?"),
    key.WithHelp("?", "Show full help"),
  ),
}

func (k keyMap) ShortHelp() []key.Binding {
  return []key.Binding{
    k.Help,
    k.Quit,
  }
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
    {k.Up, k.Down, k.PageUp, k.PageDown, k.Top, k.Bottom, k.Enter},
    {k.Help, k.Quit},
  }
}
