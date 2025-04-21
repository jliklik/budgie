package main

import (
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type postInsertCSVScreenModel struct {
	table    table.Model
	expenses []Expense
	lg       *lipgloss.Renderer
	styles   *Styles
	width    int
}

const DateWidth = 5
const DefaultWidth = 15
const DescriptionWidth = 36
const LegendWidth = 50

func createPostInsertCSVScreenModel(expenses []Expense) postInsertCSVScreenModel {

	columns := []table.Column{
		{Title: "Valid?", Width: DefaultWidth},
		{Title: "Year", Width: DateWidth},
		{Title: "Month", Width: DateWidth},
		{Title: "Day", Width: DateWidth},
		{Title: "Description", Width: DescriptionWidth},
		{Title: "Debit", Width: DefaultWidth},
		{Title: "Credit", Width: DefaultWidth},
	}

	// convert expense into "row"
	rows := []table.Row{}

	for _, e := range expenses {

		successful := "✅"
		if !e.Valid {
			successful = "❌"
		}

		rows = append(rows, table.Row{
			successful,
			strconv.Itoa(e.Year),
			strconv.Itoa(e.Month),
			strconv.Itoa(e.Day),
			e.Description,
			strconv.FormatFloat(e.Debit, 'f', 2, 64),
			strconv.FormatFloat(e.Credit, 'f', 2, 64)})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	s := table.DefaultStyles()
	s.Header = s.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(brown).
		BorderBottom(true).
		Bold(false)
	s.Selected = s.Selected.
		Foreground(lipgloss.Color("229")).
		Background(brown).
		Bold(false)
	t.SetStyles(s)

	m := postInsertCSVScreenModel{
		expenses: expenses,
		table:    t,
		width:    maxWidth,
	}

	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	return m
}

func (m postInsertCSVScreenModel) Init() tea.Cmd {
	return nil
}

func (m postInsertCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			if m.table.Focused() {
				m.table.Blur()
			} else {
				m.table.Focus()
			}
		case "q", "ctrl+c":
			new_model := createHomeScreenModel()
			return new_model, new_model.Init()
		case "enter":
			return m, tea.Batch(
				tea.Printf("Let's go to %s!", m.table.SelectedRow()[1]),
			)
		}
	}
	m.table, cmd = m.table.Update(msg)
	return m, cmd

}

func (m postInsertCSVScreenModel) View() string {
	s := m.styles
	body := s.Table.Render(m.table.View())

	any_invalid := false
	for _, e := range m.expenses {
		if !e.Valid {
			any_invalid = true
			break
		}
	}

	header := m.appBoundaryView("Budgie - Budget Manager v0.0.1 🪺 ")

	footer := m.appBoundaryView("Press Ctrl+C to return to home screen")
	if any_invalid {
		footer = m.appBoundaryView("Note: One or more entries not entered due to invalid fields.")
	}

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)

}

func (m postInsertCSVScreenModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(brown),
	)
}

func (m postInsertCSVScreenModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}
