// This model is shared between delete and update views
// probably not the best idea after all...

package main

import (
	"log"
	"slices"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	update_entries_view = iota
	update_action_view  = iota
	update_num_views    = iota
)

type UpdateEntriesModel struct {
	entry_to_search        Expense
	active_view            int
	feedback               string
	found_entries          []Expense
	entries                []ExpensePlaceholder
	found_entries_page_idx int
	cursor                 Cursor2D
	track_edits_table      TrackEditsTable
	prompt_text            string
	prompt_text_style      int
}

func createUpdateEntriesModel(found_entries []Expense, entry_to_search Expense) UpdateEntriesModel {
	model := UpdateEntriesModel{
		entry_to_search: entry_to_search,
		found_entries:   found_entries,
		feedback:        default_feedback,
		active_view:     update_entries_view,
		prompt_text:     default_feedback,
	}

	return populateUpdateEntries(model)
}

func populateUpdateEntries(m UpdateEntriesModel) UpdateEntriesModel {

	m.entries = make([]ExpensePlaceholder, len(m.found_entries))

	modified := make([][]bool, len(m.found_entries)) // Allocate the outer slice
	for i := range modified {
		modified[i] = make([]bool, expense_credit+1) // Allocate each inner slice
	}

	valid := make([][]valid_status, len(m.found_entries)) // Allocate the outer slice
	for i := range valid {
		valid[i] = make([]valid_status, expense_credit+1) // Allocate each inner slice
	}

	m.track_edits_table = TrackEditsTable{
		modified: modified,
		valid:    valid,
	}

	log.Printf("update form number of entries: %d", len(m.found_entries))

	// Convert found_entries all to string
	for idx, entry := range m.found_entries {
		m.entries[idx].Year = strconv.Itoa(entry.Year)
		m.entries[idx].Month = strconv.Itoa(entry.Month)
		m.entries[idx].Day = strconv.Itoa(entry.Day)
		m.entries[idx].Description = entry.Description
		m.entries[idx].Debit = strconv.FormatFloat(entry.Debit, 'f', 2, 64)
		m.entries[idx].Credit = strconv.FormatFloat(entry.Credit, 'f', 2, 64)
	}

	return m
}

func (m UpdateEntriesModel) Init() tea.Cmd {
	return nil
}

func (m UpdateEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			if m.active_view == update_entries_view {
				if m.cursor.y > 0 {
					m.cursor.y--
				}
			}
		case "down":
			if m.active_view == update_entries_view {
				num_entries_on_page := min(num_entries_per_page, len(m.found_entries)-(m.found_entries_page_idx*num_entries_per_page))
				if m.cursor.y < num_entries_on_page-1 {
					m.cursor.y++
				}
			}
		case "left":
			if m.cursor.x > 0 {
				m.cursor.x--
			} else {
				m.cursor.x = expense_credit
				m.cursor.y--
			}
		case "right":
			if m.cursor.x < expense_credit {
				m.cursor.x++
			} else {
				num_entries_on_page := min(num_entries_per_page, len(m.found_entries)-(m.found_entries_page_idx*num_entries_per_page))
				if m.cursor.y < num_entries_on_page-1 {
					m.cursor.y++
					m.cursor.x = 0
				}
			}
		case ">":
			if (m.found_entries_page_idx+1)*max_entries < len(m.entries) {
				m.found_entries_page_idx += 1
			}
		case "<":
			if m.found_entries_page_idx > 0 {
				m.found_entries_page_idx -= 1
			}
		case "backspace":
			start_idx := m.found_entries_page_idx * max_entries
			row := start_idx + m.cursor.y

			entry := &m.entries[row]
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

			checkIfEntryModified(&m, row)
		case "tab":
			m.active_view = (m.active_view + 1) % update_num_views
		case "enter":
			if m.active_view == update_action_view {

				valid_modified_entries, valid_modified_rows := getValidModifiedEntries(&m)

				// get entries being modified
				original_entries_being_modified := []Expense{}
				for row := 0; row < len(m.found_entries); row++ {
					if slices.Contains(valid_modified_rows, row) {
						original_entries_being_modified = append(original_entries_being_modified, m.found_entries[row])
					}
				}

				invalid := checkForInvalidEntries(&m) || len(valid_modified_entries) == 0

				if !invalid {

					log.Print(original_entries_being_modified)
					log.Print(valid_modified_entries)

					mongoUpdateEntries(original_entries_being_modified, valid_modified_entries)
					insertingCsvScreenModel := createPostInsertCSVScreenModel(valid_modified_entries)
					return insertingCsvScreenModel, nil
				} else {
					if len(original_entries_being_modified) == 0 {
						m.prompt_text = "No entries were modified."
						m.prompt_text_style = 1
					} else {
						m.prompt_text = "Some errors were detected (highlighted). Please fix and re-enter."
						m.prompt_text_style = 1
					}
					m.active_view = insert_table_view
				}
			}
		case "ctrl+c":
			return createHomeScreenModel(), nil
		default:
			start_idx := m.found_entries_page_idx * max_entries
			row := start_idx + m.cursor.y

			entry := &m.entries[row]
			switch m.cursor.x {
			case expense_year:
				if len(entry.Year) < DateWidth {
					entry.Year += msg.String()
				}
			case expense_month:
				if len(entry.Month) < DateWidth {
					entry.Month += msg.String()
				}
			case expense_day:
				if len(entry.Day) < DateWidth {
					entry.Day += msg.String()
				}
			case expense_description:
				if len(entry.Description) < DescriptionWidth {
					entry.Description += msg.String()
				}
			case expense_debit:
				if len(entry.Debit) < DefaultWidth {
					entry.Debit += msg.String()
				}
			case expense_credit:
				if len(entry.Credit) < DefaultWidth {
					entry.Credit += msg.String()
				}
			}

			checkIfEntryModified(&m, row)
		}
	}

	return m, nil
}

func checkForInvalidEntries(m *UpdateEntriesModel) bool {
	any_entry_invalid := false
	for y := 0; y < len(m.found_entries); y++ {
		for x := 0; x < (expense_credit + 1); x++ {
			if (m.track_edits_table.valid[y][x]) == valid_error {
				any_entry_invalid = true
				break
			}
		}
	}
	return any_entry_invalid
}

// Scans all entries to find valid modified entries
func getValidModifiedEntries(m *UpdateEntriesModel) ([]Expense, []int) {
	entries := []Expense{}
	valid_modified_rows := []int{}

	for row := 0; row < len(m.found_entries); row++ {

		entry := Expense{}

		// see if fields are valid
		for col := 0; col < (expense_credit + 1); col++ {
			switch col {
			case expense_year:
				if m.entries[row].Year != "" {
					year, err := strconv.Atoi(m.entries[row].Year)
					if err == nil {
						entry.Year = year
						m.track_edits_table.valid[row][col] = valid_selected
					} else {
						m.track_edits_table.valid[row][col] = valid_error
					}
				} else {
					if m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			case expense_month:
				if m.entries[row].Month != "" {
					month, err := time.Parse("Jan", m.entries[row].Month)
					if err == nil {
						entry.Month = int(month.Month())
						m.track_edits_table.valid[row][col] = valid_selected
					} else {
						// try parsing number
						month, err := strconv.Atoi(m.entries[row].Month)
						if err == nil && month >= 1 && month <= 12 {
							entry.Month = month
							m.track_edits_table.valid[row][col] = valid_selected
						} else {
							m.track_edits_table.valid[row][col] = valid_error
						}
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			case expense_day:
				if m.entries[row].Day != "" {
					day, err := strconv.Atoi(m.entries[row].Day)
					if err == nil && day >= 1 && day <= 31 {
						entry.Day = day
						m.track_edits_table.valid[row][col] = valid_selected
					} else {
						m.track_edits_table.valid[row][col] = valid_error
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			case expense_description:
				if m.entries[row].Description != "" {
					entry.Description = m.entries[row].Description
					m.track_edits_table.valid[row][col] = valid_selected
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			case expense_debit:
				if m.entries[row].Debit != "" {
					val, err := strconv.ParseFloat(m.entries[row].Debit, 64)
					if err == nil {
						entry.Debit = val
						m.track_edits_table.valid[row][col] = valid_selected
					} else {
						m.track_edits_table.valid[row][col] = valid_error
					}
				} else {
					if (m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "") && m.entries[row].Credit == "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			case expense_credit:
				if m.entries[row].Credit != "" {
					val, err := strconv.ParseFloat(m.entries[row].Credit, 64)
					if err == nil {
						entry.Credit = val
						m.track_edits_table.valid[row][col] = valid_selected
					} else {
						m.track_edits_table.valid[row][col] = valid_error
					}
				} else {
					if (m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "") && m.entries[row].Debit == "" {
						m.track_edits_table.valid[row][col] = valid_error
					} else {
						m.track_edits_table.valid[row][col] = valid_inactive
					}
				}
			}
		}

		// Check if entry is valid
		checkValidEntryValues(&entry)

		// Check if entry was modified
		for col := 0; col < (expense_credit + 1); col++ {
			if m.track_edits_table.modified[row][col] {
				entries = append(entries, entry)
				valid_modified_rows = append(valid_modified_rows, row)
				break
			}
		}

	}

	return entries, valid_modified_rows
}

func checkIfEntryModified(m *UpdateEntriesModel, row int) {

	if strconv.Itoa(m.found_entries[row].Year) != m.entries[row].Year {
		m.track_edits_table.modified[row][expense_year] = true
	} else {
		m.track_edits_table.modified[row][expense_year] = false
	}

	if strconv.Itoa(m.found_entries[row].Month) != m.entries[row].Month {
		m.track_edits_table.modified[row][expense_month] = true
	} else {
		m.track_edits_table.modified[row][expense_month] = false
	}

	if strconv.Itoa(m.found_entries[row].Day) != m.entries[row].Day {
		m.track_edits_table.modified[row][expense_day] = true
	} else {
		m.track_edits_table.modified[row][expense_day] = false
	}

	if m.found_entries[row].Description != m.entries[row].Description {
		m.track_edits_table.modified[row][expense_description] = true
	} else {
		m.track_edits_table.modified[row][expense_description] = false
	}

	if strconv.FormatFloat(m.found_entries[row].Debit, 'f', 2, 64) != m.entries[row].Debit {
		m.track_edits_table.modified[row][expense_debit] = true
	} else {
		m.track_edits_table.modified[row][expense_debit] = false
	}

	if strconv.FormatFloat(m.found_entries[row].Credit, 'f', 2, 64) != m.entries[row].Credit {
		m.track_edits_table.modified[row][expense_credit] = true
	} else {
		m.track_edits_table.modified[row][expense_credit] = false
	}

}

func (m UpdateEntriesModel) View() string {
	s := ""
	s = renderUpdateExpenses(m, s)
	s += selectPromptTextStyle(m).Render(m.prompt_text) + "\n"
	s = renderUpdateActions(m, s)
	return s
}

func renderUpdateExpenses(m UpdateEntriesModel, s string) string {
	sym := " "
	if m.active_view == update_entries_view {
		sym = "[x]"
	}

	s += "\n" + textStyle.Width((DateWidth+3)*3).Render("Matching Entries")

	if len(m.found_entries) > 0 {
		page_str := "Entries: " +
			strconv.Itoa(m.found_entries_page_idx*num_entries_per_page+1) + "-" +
			strconv.Itoa(min((m.found_entries_page_idx+1)*num_entries_per_page, len(m.found_entries))) + " / " +
			strconv.Itoa(len(m.found_entries))

		s += textStyle.Width(DescriptionWidth + 3).Render(page_str)
		s += textStyle.Width((DefaultWidth + 3) * 2).Render("Press < or > to switch pages")
		s += activeUpdateViewStyle(m.active_view, update_entries_view).Width(3).Render(sym)
	}

	s += "\n" + textStyle.Render("Press tab to switch between search, entry, and delete sections.")
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
	s += "\n"

	// slice entries
	start_idx := m.found_entries_page_idx * max_entries
	end_idx := min(start_idx+max_entries, len(m.entries))
	sliced_entries := m.entries[start_idx:end_idx]

	log.Printf("length of sliced entries %d", len(sliced_entries))

	for y, entry := range sliced_entries {

		row := start_idx + y

		line := selectUpdateEntryStyle(m, y, row, expense_year).Width(DateWidth).Render(entry.Year)
		line += " | "
		line += selectUpdateEntryStyle(m, y, row, expense_month).Width(DateWidth).Render(entry.Month)
		line += " | "
		line += selectUpdateEntryStyle(m, y, row, expense_day).Width(DateWidth).Render(entry.Day)
		line += " | "
		line += selectUpdateEntryStyle(m, y, row, expense_description).Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += selectUpdateEntryStyle(m, y, row, expense_debit).Width(DefaultWidth).Render(entry.Debit)
		line += " | "
		line += selectUpdateEntryStyle(m, y, row, expense_credit).Width(DefaultWidth).Render(entry.Credit)

		s += line + "\n"
	}

	return s
}

func renderUpdateActions(m UpdateEntriesModel, s string) string {
	s += "\n" + textStyle.PaddingRight(2).Render("Edit selected entries?")

	sym := ""
	if m.active_view == update_action_view {
		sym = "Press enter to edit selected entries"
	}
	s += activeUpdateViewStyle(m.active_view, update_action_view).Render(sym) + "\n"

	return s
}

func activeUpdateViewStyle(active_view int, view int) lipgloss.Style {
	if view == active_view {
		return selectedStyle
	}

	return textStyle
}

func selectPromptTextStyle(m UpdateEntriesModel) lipgloss.Style {
	if m.prompt_text_style == 0 {
		return textStyle
	} else {
		return errorStyle
	}
}

// highlights specific cell
func selectUpdateEntryStyle(m UpdateEntriesModel, y int, row int, col int) lipgloss.Style {

	log.Printf("y: %d x: %d", y, col)

	if m.cursor.x == col && m.cursor.y == y {
		return selectedStyle
	} else if m.track_edits_table.valid[row][col] == valid_error {
		return errorStyle
	} else if m.track_edits_table.modified[row][col] {
		return questionStyle
	}

	return inactiveStyle
}
