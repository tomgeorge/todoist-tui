package cmd

import (
	tea "github.com/charmbracelet/bubbletea"
)

// Can send to child components to notify them to resize.
// Includes the parent containing component's frame size.
// containing component size is the size of all margins, borders, padding, etc.
type ResizeChildMessage struct {
	// Screen width
	Width int
	// Screen height
	Height int
	// Containing component width
	ParentFrameWidth int
	// Containing component height
	ParentFrameHeight int
}

func NotifyResize(x, y, px, py int) tea.Cmd {
	return func() tea.Msg {
		return ResizeChildMessage{
			Width:             x,
			Height:            y,
			ParentFrameWidth:  px,
			ParentFrameHeight: py,
		}
	}
}
