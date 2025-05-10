package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type postInsertCSVScreenModel struct {
	entries          []Expense
	entries_page_idx int
}

const DateWidth = 5
const DefaultWidth = 15
const DescriptionWidth = 36
const LegendWidth = 50

func createPostInsertCSVScreenModel(entries []Expense) postInsertCSVScreenModel {
	return postInsertCSVScreenModel{
		entries: entries,
	}
}

func (m postInsertCSVScreenModel) Init() tea.Cmd {
	return nil
}

func (m postInsertCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "left":
			if m.entries_page_idx > 0 {
				m.entries_page_idx -= 1
			}

		case "right":
			if (m.entries_page_idx+1)*max_entries < len(m.entries) {
				m.entries_page_idx += 1
			}

		case "ctrl+c":
			return createHomeScreenModel(), nil
		}
	}

	return m, nil
}

func (m postInsertCSVScreenModel) View() string {
	s := ""
	s += displayLegend(&m)
	s += displayExpenses(&m)
	s += "\n" + textStyle.PaddingLeft(2).Render("Press Ctrl+C to go back to home screen.") + "\n"
	return s
}

func displayLegend(m *postInsertCSVScreenModel) string {
	s := textStyle.Width(LegendWidth).Render("Legend") + "\n"
	s += errorStyle.Width(LegendWidth).Render("Not inserted into DB - invalid or duplicate") + "\n"
	s += selectedStyle.Width(LegendWidth).Render("Successfully inserted into DB") + "\n\n"

	if len(m.entries) > 0 {
		page_str := "Entries: " +
			strconv.Itoa(m.entries_page_idx*num_entries_per_page+1) + "-" +
			strconv.Itoa(min((m.entries_page_idx+1)*num_entries_per_page, len(m.entries))) + " / " +
			strconv.Itoa(len(m.entries))

		s += textStyle.Width(DescriptionWidth + 3).Render(page_str)
		s += textStyle.Width((DefaultWidth + 3) * 2).Render("Press <- or -> to switch pages")
		s += "\n"
	}

	return s
}

func displayExpenses(m *postInsertCSVScreenModel) string {

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

	start_idx := m.entries_page_idx * max_entries
	end_idx := min(start_idx+max_entries, len(m.entries))
	sliced_entries := m.entries[start_idx:end_idx]

	for _, entry := range sliced_entries {

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
