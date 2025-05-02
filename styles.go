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

const maxWidth = 120

var (
	red = lipgloss.AdaptiveColor{Light: "#FE5F86", Dark: "#FE5F86"}
	// indigo = lipgloss.AdaptiveColor{Light: "#5A56E0", Dark: "#7571F9"}
	green = lipgloss.AdaptiveColor{Light: "#02BA84", Dark: "#02BF87"}
	brown = lipgloss.AdaptiveColor{Light: "#873e23", Dark: "#eab676"}
)

type Styles struct {
	Base,
	HeaderText,
	Status,
	StatusHeader,
	Highlight,
	ErrorHeaderText,
	Table,
	Help lipgloss.Style
}

func NewStyles(lg *lipgloss.Renderer) *Styles {
	s := Styles{}
	s.Base = lg.NewStyle().
		Padding(1, 4, 0, 1)
	s.HeaderText = lg.NewStyle().
		Foreground(brown).
		Bold(true).
		Padding(0, 1, 0, 2)
	s.Status = lg.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(brown).
		PaddingLeft(1).
		MarginTop(1)
	s.StatusHeader = lg.NewStyle().
		Foreground(green).
		Bold(true)
	s.Highlight = lg.NewStyle().
		Foreground(lipgloss.Color("212"))
	s.ErrorHeaderText = s.HeaderText.
		Foreground(red)
	s.Table = lg.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(brown)
	s.Help = lg.NewStyle().
		Foreground(lipgloss.Color("240"))
	return &s
}
