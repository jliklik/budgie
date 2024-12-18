package main

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type findEntryScreenModel struct {
	fields          [num_expense_search_fields]string
	validated       [num_expense_search_fields]bool
	feedback        string
	cursor          int
	entry_to_search Expense
	found_entries   []Expense
}

const FindEntryScreenWidth = 30
const FindEntryScreenLabelWidth = 15
const num_expense_search_fields = expense_credit + 1

func createFindEntryScreenModel() findEntryScreenModel {
	return findEntryScreenModel{
		entry_to_search: Expense{},
		validated:       [num_expense_search_fields]bool{false, false, false, false, false, false},
		found_entries:   nil,
		feedback:        "Press Ctrl+C to go back.",
	}
}

func (m findEntryScreenModel) Init() tea.Cmd {
	return nil
}

func (m findEntryScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			if m.cursor > expense_year {
				m.cursor--
			}
		case "down":
			if m.cursor < num_expense_search_fields {
				m.cursor++
			}
		case "left":
			// do nothing
		case "right":
			// do nothing
		case "backspace":
			sz := len(m.fields[m.cursor])
			if sz >= 1 {
				m.fields[m.cursor] = m.fields[m.cursor][:sz-1]
			}
		case "enter":
			switch m.cursor {
			case expense_year:
				year, err := strconv.Atoi(m.fields[m.cursor])
				if err == nil {
					m.entry_to_search.year = year
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				} else {
					m.feedback = "Invalid year!"
				}
			case expense_month:
				month, err := time.Parse("Jan", m.fields[m.cursor])
				if err == nil {
					m.entry_to_search.month = int(month.Month())
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				} else {
					m.feedback = "Invalid month! Format: Jan, Feb, Mar, etc."
				}
			case expense_day:
				day, err := strconv.Atoi(m.fields[m.cursor])
				if err == nil && day >= 1 && day <= 31 {
					m.entry_to_search.day = day
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				} else {
					m.feedback = "Invalid day! Must be between 1 and 31."
				}
			case expense_description:
				m.entry_to_search.description = m.fields[m.cursor]
				m.validated[m.cursor] = true
				m.cursor++
				m.feedback = "Press Ctrl+C to go back."
			case expense_debit:
				val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
				if err == nil {
					m.entry_to_search.debit = val
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				} else {
					m.feedback = "Invalid debit amount!"
				}
			case expense_credit:
				val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
				if err == nil {
					m.entry_to_search.debit = val
					m.validated[m.cursor] = true
					m.feedback = "Press Ctrl+C to go back."
				} else {
					m.feedback = "Invalid debit amount!"
				}

			}
		case "ctrl+c":
			return createHomeScreenModel(), nil
		default:
			m.fields[m.cursor] += msg.String()
		}
	}

	return m, nil
}

func (m findEntryScreenModel) View() string {
	s := textStyle.Render("Enter in details of entry to search for. Leave blank to search all.") + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Year: ") +
		selectStyle(m, expense_year).Width(FindEntryScreenWidth).Render(m.fields[expense_year]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Month: ") +
		selectStyle(m, expense_month).Width(FindEntryScreenWidth).Render(m.fields[expense_month]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Day: ") +
		selectStyle(m, expense_day).Width(FindEntryScreenWidth).Render(m.fields[expense_day]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Description: ") +
		selectStyle(m, expense_description).Width(FindEntryScreenWidth).Render(m.fields[expense_description]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Debit: ") +
		selectStyle(m, expense_debit).Width(FindEntryScreenWidth).Render(m.fields[expense_debit]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryScreenLabelWidth).Render("Credit: ") +
		selectStyle(m, expense_credit).Width(FindEntryScreenWidth).Render(m.fields[expense_credit]) + "\n"
	s += textStyle.Render(m.feedback)
	return s
}

func selectStyle(m findEntryScreenModel, index int) lipgloss.Style {
	if m.cursor == index {
		return selectedStyle.PaddingLeft(2).PaddingRight(2)
	} else if !m.validated[index] {
		return errorStyle.PaddingLeft(2).PaddingRight(2)
	} else {
		return inactiveStyle.PaddingLeft(2).PaddingRight(2)
	}
}

// See if entry is already in DB before inserting it
// filter := bson.D{{"year", entry.year}, {"month", entry.month}, {"day", entry.day}, {"description", entry.description}, {"debit", entry.debit}, {"credit", entry.credit}}
// cursor, err := coll.Find(context.TODO(), filter)
// if err != nil {
// 	panic(err)
// }
// var results []Expense
// if err = cursor.All(context.TODO(), &results); err != nil {
// 	panic(err)
// }
