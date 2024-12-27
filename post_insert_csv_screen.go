package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type postInsertCSVScreenModel struct {
	expenses []Expense
}

const DateWidth = 5
const DefaultWidth = 15
const DescriptionWidth = 36
const LegendWidth = 50

func createPostInsertCSVScreenModel(expenses []Expense) postInsertCSVScreenModel {
	return postInsertCSVScreenModel{
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
		if !entry.Valid {
			style = errorStyle
		}

		line := style.Width(DateWidth).Render(strconv.Itoa(entry.Year))
		line += " | "
		line += style.Width(DateWidth).Render(strconv.Itoa(entry.Month))
		line += " | "
		line += style.Width(DateWidth).Render(strconv.Itoa(entry.Day))
		line += " | "
		line += style.Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += style.Width(DefaultWidth).Render(strconv.FormatFloat(entry.Debit, 'f', 2, 64))
		line += " | "
		line += style.Width(DefaultWidth).Render(strconv.FormatFloat(entry.Credit, 'f', 2, 64))
		s += line + "\n"
	}

	return s
}
