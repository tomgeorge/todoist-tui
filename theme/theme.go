// Basically completely ripped from charmbracelet/huh
package theme

import (
	catppuccin "github.com/catppuccin/go"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/lipgloss"
)

// I ripped a lot of this from github.com/charmbracelet/huh
type Theme struct {
	Form           lipgloss.Style
	Group          lipgloss.Style
	FieldSeparator lipgloss.Style
	Blurred        FieldStyles
	Focused        FieldStyles
	Help           help.Styles
}

// FieldStyles are the styles for input fields.
type FieldStyles struct {
	Base           lipgloss.Style
	Title          lipgloss.Style
	Description    lipgloss.Style
	ErrorIndicator lipgloss.Style
	ErrorMessage   lipgloss.Style

	// Select styles.
	SelectSelector lipgloss.Style // Selection indicator
	Option         lipgloss.Style // Select options

	// Multi-select styles.
	MultiSelectSelector lipgloss.Style
	SelectedOption      lipgloss.Style
	SelectedPrefix      lipgloss.Style
	UnselectedOption    lipgloss.Style
	UnselectedPrefix    lipgloss.Style

	// Textinput and teatarea styles.
	TextInput TextInputStyles

	// Confirm styles.
	FocusedButton lipgloss.Style
	BlurredButton lipgloss.Style

	// Card styles.
	Card      lipgloss.Style
	NoteTitle lipgloss.Style
	Next      lipgloss.Style
}

// copy returns a copy of a FieldStyles with all children styles copied.
func (f FieldStyles) copy() FieldStyles {
	return FieldStyles{
		Base:                f.Base.Copy(),
		Title:               f.Title.Copy(),
		Description:         f.Description.Copy(),
		ErrorIndicator:      f.ErrorIndicator.Copy(),
		ErrorMessage:        f.ErrorMessage.Copy(),
		SelectSelector:      f.SelectSelector.Copy(),
		Option:              f.Option.Copy(),
		MultiSelectSelector: f.MultiSelectSelector.Copy(),
		SelectedOption:      f.SelectedOption.Copy(),
		SelectedPrefix:      f.SelectedPrefix.Copy(),
		UnselectedOption:    f.UnselectedOption.Copy(),
		UnselectedPrefix:    f.UnselectedPrefix.Copy(),
		FocusedButton:       f.FocusedButton.Copy(),
		BlurredButton:       f.BlurredButton.Copy(),
		TextInput:           f.TextInput.copy(),
		Card:                f.Card.Copy(),
		NoteTitle:           f.NoteTitle.Copy(),
		Next:                f.Next.Copy(),
	}
}

// TextInputStyles are the styles for text inputs.
type TextInputStyles struct {
	Cursor      lipgloss.Style
	Placeholder lipgloss.Style
	Prompt      lipgloss.Style
	Text        lipgloss.Style
}

// copy returns a copy of a TextInputStyles with all children styles copied.
func (t TextInputStyles) copy() TextInputStyles {
	return TextInputStyles{
		Cursor:      t.Cursor.Copy(),
		Placeholder: t.Placeholder.Copy(),
		Prompt:      t.Prompt.Copy(),
		Text:        t.Text.Copy(),
	}
}

const (
	buttonPaddingHorizontal = 2
	buttonPaddingVertical   = 0
)

// copy returns a copy of a theme with all children styles copied.
func (t Theme) copy() Theme {
	return Theme{
		Form:           t.Form.Copy(),
		Group:          t.Group.Copy(),
		FieldSeparator: t.FieldSeparator.Copy(),
		Blurred:        t.Blurred.copy(),
		Focused:        t.Focused.copy(),
		Help: help.Styles{
			Ellipsis:       t.Help.Ellipsis.Copy(),
			ShortKey:       t.Help.ShortKey.Copy(),
			ShortDesc:      t.Help.ShortDesc.Copy(),
			ShortSeparator: t.Help.ShortSeparator.Copy(),
			FullKey:        t.Help.FullKey.Copy(),
			FullDesc:       t.Help.FullDesc.Copy(),
			FullSeparator:  t.Help.FullSeparator.Copy(),
		},
	}
}

// ThemeBase returns a new base theme with general styles to be inherited by
// other themes.
func ThemeBase() *Theme {
	var t Theme

	t.FieldSeparator = lipgloss.NewStyle().SetString("\n\n")

	button := lipgloss.NewStyle().
		Padding(buttonPaddingVertical, buttonPaddingHorizontal).
		MarginRight(1)

	// Focused styles.
	f := &t.Focused
	f.Base = lipgloss.NewStyle().
		PaddingLeft(1).
		BorderStyle(lipgloss.ThickBorder()).
		BorderLeft(true).
		MarginBottom(2)
	f.Card = lipgloss.NewStyle().
		PaddingLeft(1)
	f.ErrorIndicator = lipgloss.NewStyle().
		SetString(" *")
	f.ErrorMessage = lipgloss.NewStyle().
		SetString(" *")
	f.SelectSelector = lipgloss.NewStyle().
		SetString("> ")
	f.MultiSelectSelector = lipgloss.NewStyle().
		SetString("> ")
	f.SelectedPrefix = lipgloss.NewStyle().
		SetString("[•] ")
	f.UnselectedPrefix = lipgloss.NewStyle().
		SetString("[ ] ")
	f.FocusedButton = button.Copy().
		Foreground(lipgloss.Color("0")).
		Background(lipgloss.Color("7"))
	f.BlurredButton = button.Copy().
		Foreground(lipgloss.Color("7")).
		Background(lipgloss.Color("0"))
	f.TextInput.Placeholder = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))

	t.Help = help.New().Styles

	// Blurred styles.
	t.Blurred = f.copy()
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.MultiSelectSelector = lipgloss.NewStyle().SetString("  ")

	return &t
}

func ThemeCatppuccin() *Theme {
	t := ThemeBase().copy()

	light := catppuccin.Latte
	dark := catppuccin.Mocha
	var (
		base     = lipgloss.AdaptiveColor{Light: light.Base().Hex, Dark: dark.Base().Hex}
		text     = lipgloss.AdaptiveColor{Light: light.Text().Hex, Dark: dark.Text().Hex}
		subtext1 = lipgloss.AdaptiveColor{Light: light.Subtext1().Hex, Dark: dark.Subtext1().Hex}
		subtext0 = lipgloss.AdaptiveColor{Light: light.Subtext0().Hex, Dark: dark.Subtext0().Hex}
		overlay1 = lipgloss.AdaptiveColor{Light: light.Overlay1().Hex, Dark: dark.Overlay1().Hex}
		overlay0 = lipgloss.AdaptiveColor{Light: light.Overlay0().Hex, Dark: dark.Overlay0().Hex}
		green    = lipgloss.AdaptiveColor{Light: light.Green().Hex, Dark: dark.Green().Hex}
		red      = lipgloss.AdaptiveColor{Light: light.Red().Hex, Dark: dark.Red().Hex}
		pink     = lipgloss.AdaptiveColor{Light: light.Pink().Hex, Dark: dark.Pink().Hex}
		mauve    = lipgloss.AdaptiveColor{Light: light.Mauve().Hex, Dark: dark.Mauve().Hex}
		cursor   = lipgloss.AdaptiveColor{Light: light.Rosewater().Hex, Dark: dark.Rosewater().Hex}
	)

	f := &t.Focused
	f.Base.BorderForeground(subtext1)
	f.Title.Foreground(mauve)
	f.NoteTitle.Foreground(mauve)
	f.Description.Foreground(subtext0)
	f.ErrorIndicator.Foreground(red)
	f.ErrorMessage.Foreground(red)
	f.SelectSelector.Foreground(pink)
	f.Option.Foreground(text)
	f.MultiSelectSelector.Foreground(pink)
	f.SelectedOption.Foreground(green)
	f.SelectedPrefix.Foreground(green)
	f.UnselectedPrefix.Foreground(text)
	f.UnselectedOption.Foreground(text)
	f.FocusedButton.Foreground(base).Background(pink)
	f.BlurredButton.Foreground(text).Background(base)

	f.TextInput.Cursor.Foreground(cursor)
	f.TextInput.Placeholder.Foreground(overlay0)
	f.TextInput.Prompt.Foreground(pink)

	t.Blurred = f.copy()
	t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())

	t.Help.Ellipsis.Foreground(subtext0)
	t.Help.ShortKey.Foreground(subtext0)
	t.Help.ShortDesc.Foreground(overlay1)
	t.Help.ShortSeparator.Foreground(subtext0)
	t.Help.FullKey.Foreground(subtext0)
	t.Help.FullDesc.Foreground(overlay1)
	t.Help.FullSeparator.Foreground(subtext0)

	return &t
}
