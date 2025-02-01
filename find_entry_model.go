package main

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const FindEntryLabelWidth = 20

type findEntryModel struct {
	fields          [num_expense_search_fields]string
	validated       [num_expense_search_fields]bool
	feedback        string
	search_cursor   int
	entry_to_search Expense
	found_entries   []Expense
	action          action
}

func createFindEntryModel(action action) findEntryModel {
	return findEntryModel{
		entry_to_search: Expense{
			Month:  invalid,
			Day:    invalid,
			Year:   invalid,
			Debit:  invalid,
			Credit: invalid,
		},
		validated: [num_expense_search_fields]bool{false, false, false, false, false, false},
		feedback:  default_feedback,
		action:    action,
	}
}

func (m findEntryModel) Init() tea.Cmd {
	return nil
}

func (m findEntryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			if m.search_cursor > expense_year {
				m.search_cursor--
			}

		case "down":
			if m.search_cursor < expense_credit {
				m.search_cursor++
			}

		case "backspace":
			sz := len(m.fields[m.search_cursor])
			if sz >= 1 {
				m.fields[m.search_cursor] = m.fields[m.search_cursor][:sz-1]
			}

		case "enter":
			switch m.search_cursor {
			case expense_year:
				if m.fields[m.search_cursor] != "" {
					year, err := strconv.Atoi(m.fields[m.search_cursor])
					if err == nil {
						m.entry_to_search.Year = year
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = default_feedback
					} else {
						m.validated[m.search_cursor] = false
						m.feedback = "Invalid year!"
					}
				} else {
					m.entry_to_search.Year = invalid
					m.validated[m.search_cursor] = true
					m.search_cursor++
					m.feedback = default_feedback
				}
			case expense_month:
				if m.fields[m.search_cursor] != "" {
					month, err := time.Parse("Jan", m.fields[m.search_cursor])
					if err == nil {
						m.entry_to_search.Month = int(month.Month())
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = default_feedback
					} else {
						// try parsing number
						month, err := strconv.Atoi(m.fields[m.search_cursor])
						if err == nil && month >= 1 && month <= 12 {
							m.entry_to_search.Month = month
							m.validated[m.search_cursor] = true
							m.search_cursor++
							m.feedback = default_feedback
						} else {
							m.validated[m.search_cursor] = false
							m.feedback = "Invalid month! Format: Jan, Feb, Mar, etc."
						}
					}
				} else {
					m.entry_to_search.Month = invalid
					m.validated[m.search_cursor] = true
					m.search_cursor++
					m.feedback = default_feedback
				}
			case expense_day:
				if m.fields[m.search_cursor] != "" {
					day, err := strconv.Atoi(m.fields[m.search_cursor])
					if err == nil && day >= 1 && day <= 31 {
						m.entry_to_search.Day = day
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = default_feedback
					} else {
						m.validated[m.search_cursor] = false
						m.feedback = "Invalid day! Must be between 1 and 31."
					}
				} else {
					m.entry_to_search.Day = invalid
					m.validated[m.search_cursor] = true
					m.search_cursor++
					m.feedback = default_feedback
				}
			case expense_description:
				m.entry_to_search.Description = m.fields[m.search_cursor]
				m.validated[m.search_cursor] = true
				m.search_cursor++
				m.feedback = default_feedback
			case expense_debit:
				if m.fields[m.search_cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.search_cursor], 64)
					if err == nil {
						m.entry_to_search.Debit = val
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = default_feedback
					} else {
						m.validated[m.search_cursor] = false
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.entry_to_search.Debit = invalid
					m.validated[m.search_cursor] = true
					m.search_cursor++
					m.feedback = default_feedback
				}
			case expense_credit:
				if m.fields[m.search_cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.search_cursor], 64)
					if err == nil {
						m.entry_to_search.Credit = val
						m.validated[m.search_cursor] = true
						m.feedback = default_feedback
					} else {
						m.validated[m.search_cursor] = false
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.entry_to_search.Credit = invalid
					m.validated[m.search_cursor] = true
					m.feedback = default_feedback
				}
			}
			if allValid(m) {
				m.found_entries = mongoFindMatchingEntries(m.entry_to_search)
				// TODO: transition to found_entries_screen
				if m.action.action_text == "delete" {
					return createDeleteEntriesModel(m.found_entries, m.entry_to_search), nil
				} else {
					return createUpdateEntriesModel(m.found_entries, m.entry_to_search), nil
				}
			}

		case "ctrl+c":
			return createHomeScreenModel(), nil
		default:
			m.fields[m.search_cursor] += msg.String()
		}
	}

	return m, nil
}

func (m findEntryModel) View() string {
	s := ""
	s = renderSearchBox(m, s)
	return s
}

func renderSearchBox(m findEntryModel, s string) string {
	s += textStyle.PaddingRight(1).Render("Enter in details of entry to search for. Leave blank to search all.") + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Year: ") +
		selectSearchBoxStyle(m, expense_year).Width(FindEntryLabelWidth).Render(m.fields[expense_year]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Month: ") +
		selectSearchBoxStyle(m, expense_month).Width(FindEntryLabelWidth).Render(m.fields[expense_month]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Day: ") +
		selectSearchBoxStyle(m, expense_day).Width(FindEntryLabelWidth).Render(m.fields[expense_day]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Description: ") +
		selectSearchBoxStyle(m, expense_description).Width(FindEntryLabelWidth).Render(m.fields[expense_description]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Debit: ") +
		selectSearchBoxStyle(m, expense_debit).Width(FindEntryLabelWidth).Render(m.fields[expense_debit]) + "\n"
	s += textStyle.PaddingLeft(2).Width(FindEntryLabelWidth).Render("Credit: ") +
		selectSearchBoxStyle(m, expense_credit).Width(FindEntryLabelWidth).Render(m.fields[expense_credit]) + "\n"
	s += textStyle.Render(m.feedback) + "\n"

	return s
}

func selectSearchBoxStyle(m findEntryModel, index int) lipgloss.Style {
	if m.search_cursor == index {
		return selectedStyle.PaddingLeft(2).PaddingRight(2)
	} else if !m.validated[index] {
		return errorStyle.PaddingLeft(2).PaddingRight(2)
	} else {
		return inactiveStyle.PaddingLeft(2).PaddingRight(2)
	}
}

func allValid(m findEntryModel) bool {
	for _, value := range m.validated {
		if !value {
			return false
		}
	}
	return true
}
