package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type cursor2D struct {
	x int
	y int
}

type manualInsertModel struct {
	cursor  cursor2D
	entries []expensePlaceholder
}

type expensePlaceholder struct {
	Month       string
	Day         string
	Year        string
	Description string
	Debit       string
	Credit      string
}

const max_entries = 10

func createManualInsertScreenModel() manualInsertModel {
	return manualInsertModel{
		cursor: cursor2D{
			x: 0,
			y: 0,
		},
		entries: make([]expensePlaceholder, max_entries),
	}
}

func (m manualInsertModel) Init() tea.Cmd {
	return nil
}

func (m manualInsertModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c", "q":
			return m, tea.Quit

		case "up":
			if m.cursor.y > 0 {
				m.cursor.y--
			}

		case "down":
			if m.cursor.y < max_entries-1 {
				m.cursor.y++
			}

		case "left":
			if m.cursor.x > 0 {
				m.cursor.x--
			}

		case "right":
			if m.cursor.x < expense_credit {
				m.cursor.x++
			}

		case "tab":
			if m.cursor.x < expense_credit {
				m.cursor.x++
			} else {
				m.cursor.x = 0
				m.cursor.y++
			}

		case "shift+tab":
			if m.cursor.x > 0 {
				m.cursor.x--
			} else {
				m.cursor.x = expense_credit
				m.cursor.y--
			}

		case "backspace":
			entry := &m.entries[m.cursor.y]
			switch m.cursor.x {
			case expense_year:
				entry.Year = removeLastChar(entry.Year)
			case expense_month:
				entry.Month = removeLastChar(entry.Month)
			case expense_day:
				entry.Day = removeLastChar(entry.Day)
			case expense_description:
				entry.Description = removeLastChar(entry.Description)
			case expense_debit:
				entry.Debit = removeLastChar(entry.Debit)
			case expense_credit:
				entry.Credit = removeLastChar(entry.Credit)
			}

		default:
			entry := &m.entries[m.cursor.y]
			switch m.cursor.x {
			case expense_year:
				entry.Year += msg.String()
			case expense_month:
				entry.Month += msg.String()
			case expense_day:
				entry.Day += msg.String()
			case expense_description:
				entry.Description += msg.String()
			case expense_debit:
				entry.Debit += msg.String()
			case expense_credit:
				entry.Credit += msg.String()
			}

		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func removeLastChar(s string) string {
	sz := len(s)
	if sz >= 1 {
		s = s[:sz-1]
	}

	return s
}

func (m manualInsertModel) View() string {
	s := ""
	s = renderHeader(s)
	s = renderEntries(m, s)

	// Send the UI for rendering
	return s
}

func renderHeader(s string) string {
	s += "\n" + textStyle.Render("Please fill in entries to insert into database")
	s += "\n"
	s += textStyle.Width(DateWidth).Render("Year")
	s += " | "
	s += textStyle.Width(DateWidth).Render("Month")
	s += " | "
	s += textStyle.Width(DateWidth).Render("Day")
	s += " | "
	s += textStyle.Width(DescriptionWidth).Render("Description")
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Debit")
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Credit")

	return s
}

func styleIfCursorIsHere(m manualInsertModel, x int, y int) lipgloss.Style {
	if m.cursor.x == x && m.cursor.y == y {
		return selectedStyle
	}

	return inactiveStyle
}

func renderEntries(m manualInsertModel, s string) string {
	s += "\n"
	for row, entry := range m.entries {
		line := styleIfCursorIsHere(m, expense_year, row).Width(DateWidth).Render(entry.Year)
		line += " | "
		line += styleIfCursorIsHere(m, expense_month, row).Width(DateWidth).Render(entry.Month)
		line += " | "
		line += styleIfCursorIsHere(m, expense_day, row).Width(DateWidth).Render(entry.Day)
		line += " | "
		line += styleIfCursorIsHere(m, expense_description, row).Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += styleIfCursorIsHere(m, expense_debit, row).Width(DefaultWidth).Render(entry.Debit)
		line += " | "
		line += styleIfCursorIsHere(m, expense_credit, row).Width(DefaultWidth).Render(entry.Credit)
		s += line + "\n"
	}

	return s
}
