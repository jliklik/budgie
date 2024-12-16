package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type postInsertCSVScreenModel struct {
	filename string
	expenses []Expense
}

const DateWidth = 5
const DefaultWidth = 15
const DescriptionWidth = 30
const LegendWidth = 50

const (
	date        = iota
	description = iota
	debit       = iota
	credit      = iota
	total       = iota
)

func createPostInsertCSVScreenModel(filename string, expenses []Expense) postInsertCSVScreenModel {
	return postInsertCSVScreenModel{
		filename: filename,
		expenses: expenses,
	}
}

func (m postInsertCSVScreenModel) Init() tea.Cmd {
	return nil
}

func (m postInsertCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c":
			return createHomeScreenModel(), nil
		}
	}

	return m, nil
}

func (m postInsertCSVScreenModel) View() string {
	s := ""
	s += displayLegend(s)
	s += displayExpenses(m.expenses)
	s += "\n" + textStyle.Width(HomeScreenWidth).PaddingLeft(2).Render("Press Ctrl+C to go back.") + "\n"
	return s
}

func displayLegend(s string) string {
	s += textStyle.Width(LegendWidth).Render("Legend") + "\n"
	s += errorStyle.Width(LegendWidth).Render("Not inserted into DB - invalid or duplicate") + "\n"
	s += selectedStyle.Width(LegendWidth).Render("Successfully inserted into DB") + "\n\n"
	return s
}

func displayExpenses(expenses []Expense) string {

	s := ""
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
	s += "\n"

	for _, entry := range expenses {

		style := selectedStyle
		if !entry.valid {
			style = errorStyle
		}

		line := style.Width(DateWidth).Render(strconv.Itoa(entry.year))
		line += " | "
		line += style.Width(DateWidth).Render(strconv.Itoa(entry.month))
		line += " | "
		line += style.Width(DateWidth).Render(strconv.Itoa(entry.day))
		line += " | "
		line += style.Width(DescriptionWidth).Render(entry.description)
		line += " | "
		line += style.Width(DefaultWidth).Render(strconv.FormatFloat(entry.debit, 'f', 2, 64))
		line += " | "
		line += style.Width(DefaultWidth).Render(strconv.FormatFloat(entry.credit, 'f', 2, 64))
		s += line + "\n"
	}

	return s
}
