package main

import "github.com/charmbracelet/lipgloss"

var textStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#1f4fd1")).
	Align(lipgloss.Left)

var inactiveStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#b268f2")).
	Align(lipgloss.Left)

var selectedStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#32d147")).
	Align(lipgloss.Left)

var errorStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#f542c2")).
	Align(lipgloss.Left)

var questionStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#ffbf00")).
	Align(lipgloss.Left)
