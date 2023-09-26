package main

import "github.com/charmbracelet/lipgloss"

var (
	screenWidth        int
	screenHeight       int
	projectTableHeight = 25
	projectTableWidth  = 50
)

var (
	selectedBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.ThickBorder())
	unselectedBoxStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder())
)
