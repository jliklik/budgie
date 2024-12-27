package main

import (
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	insert_table_view   = iota
	insert_confirm_view = iota
	insert_num_views    = iota
)

type cursor2D struct {
	x int
	y int
}

const (
	inactive_style = iota
	error_style    = iota
	selected_style = iota
)

type manualInsertModel struct {
	active_view int
	cursor      cursor2D
	valid       [max_entries][expense_credit + 1]int
	entries     []expensePlaceholder
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
			if m.active_view == insert_confirm_view {
				m.active_view = insert_table_view
			} else {
				if m.cursor.y > 0 {
					m.cursor.y--
				}
			}

		case "down":
			if m.cursor.y < max_entries-1 {
				m.cursor.y++
			} else {
				m.active_view = insert_confirm_view
			}

		case "left":
			if m.active_view == insert_table_view {
				if m.cursor.x > 0 {
					m.cursor.x--
				} else {
					m.cursor.x = expense_credit
					m.cursor.y--
				}
			}

		case "right":
			if m.active_view == insert_table_view {
				if m.cursor.x < expense_credit {
					m.cursor.x++
				} else {
					if m.cursor.y < max_entries-1 {
						m.cursor.y++
						m.cursor.x = 0
					} else {
						m.active_view = insert_confirm_view
					}
				}
			}

		case "tab":
			if m.active_view == insert_table_view {
				if m.cursor.x < expense_credit {
					m.cursor.x++
				} else {
					if m.cursor.y < max_entries-1 {
						m.cursor.y++
						m.cursor.x = 0
					} else {
						m.active_view = insert_confirm_view
					}
				}
			} else {
				m.active_view = insert_table_view
			}

		case "shift+tab":
			if m.active_view == insert_table_view {
				if m.cursor.x > 0 {
					m.cursor.x--
				} else {
					m.cursor.x = expense_credit
					m.cursor.y--
				}
			} else {
				m.active_view = insert_table_view
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

		case "enter":
			if m.active_view == insert_confirm_view {
				filtered := insertManualEntriesIntoMongo(&m)

				any_entry_invalid := false
				for y := 0; y < max_entries; y++ {
					for x := 0; x < (expense_credit + 1); x++ {
						if (m.valid[y][x]) == error_style {
							any_entry_invalid = true
							break
						}
					}
				}

				if !any_entry_invalid {
					mongoInsertEntries(filtered)
					insertingCsvScreenModel := createPostInsertCSVScreenModel(filtered)
					return insertingCsvScreenModel, nil
				} else {
					m.active_view = insert_table_view
				}

			} else {
				m.active_view = insert_confirm_view
			}

		default:
			entry := &m.entries[m.cursor.y]
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
	s = renderHeader(m, s)
	s = renderEntries(m, s)
	s = renderInsertAction(m, s)

	// Send the UI for rendering
	return s
}

func renderHeader(m manualInsertModel, s string) string {
	sym := " "
	if m.active_view == insert_table_view {
		sym = "[x]"
	}
	s += "\n" + textStyle.PaddingRight(1).Render("Please fill in entries to insert into database")
	s += activeViewStyle(m.active_view, insert_table_view).Width(3).Render(sym)
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
	} else if m.valid[y][x] == error_style {
		return errorStyle
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

func renderInsertAction(m manualInsertModel, s string) string {

	s += "\n" + textStyle.PaddingRight(2).Render("Insert selected entries?")

	sym := ""
	if m.active_view == insert_confirm_view {
		sym = "Press enter to delete selected entries [x]"
	}
	s += activeViewStyle(m.active_view, insert_confirm_view).Render(sym)

	return s
}

func insertManualEntriesIntoMongo(m *manualInsertModel) []Expense {
	entries := []Expense{}

	for row := 0; row < max_entries; row++ {

		entry := Expense{}

		for col := 0; col < (expense_credit + 1); col++ {
			switch col {
			case expense_year:
				if m.entries[row].Year != "" {
					year, err := strconv.Atoi(m.entries[row].Year)
					if err == nil {
						entry.Year = year
						m.valid[row][col] = selected_style
					} else {
						m.valid[row][col] = error_style
					}
				} else {
					if m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = inactive_style
					}
				}
			case expense_month:
				if m.entries[row].Month != "" {
					month, err := time.Parse("Jan", m.entries[row].Month)
					if err == nil {
						entry.Month = int(month.Month())
						m.valid[row][col] = selected_style
					} else {
						// try parsing number
						month, err := strconv.Atoi(m.entries[row].Month)
						if err == nil && month >= 1 && month <= 12 {
							entry.Month = month
							m.valid[row][col] = selected_style
						} else {
							m.valid[row][col] = error_style
						}
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = inactive_style
					}
				}
			case expense_day:
				if m.entries[row].Day != "" {
					day, err := strconv.Atoi(m.entries[row].Day)
					if err == nil && day >= 1 && day <= 31 {
						entry.Day = day
						m.valid[row][col] = selected_style
					} else {
						m.valid[row][col] = error_style
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = inactive_style
					}
				}
			case expense_description:
				if m.entries[row].Description != "" {
					entry.Description = m.entries[row].Description
					m.valid[row][col] = selected_style
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Debit != "" || m.entries[row].Credit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = inactive_style
					}
				}
			case expense_debit:
				if m.entries[row].Debit != "" {
					val, err := strconv.ParseFloat(m.entries[row].Debit, 64)
					if err == nil {
						entry.Debit = val
						m.valid[row][col] = selected_style
					} else {
						m.valid[row][col] = error_style
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Credit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = inactive_style
					}
				}
			case expense_credit:
				if m.entries[row].Credit != "" {
					val, err := strconv.ParseFloat(m.entries[row].Credit, 64)
					if err == nil {
						entry.Credit = val
						m.valid[row][col] = selected_style
					} else {
						m.valid[row][col] = error_style
					}
				} else {
					if m.entries[row].Year != "" || m.entries[row].Month != "" || m.entries[row].Day != "" || m.entries[row].Description != "" || m.entries[row].Debit != "" {
						m.valid[row][col] = error_style
					} else {
						m.valid[row][col] = selected_style // inactive_style
					}
				}
			}
		}

		// Check if entry is valid
		check_if_entry_is_valid(&entry)

		entries = append(entries, entry)
	}

	filtered := filter_empty_rows(entries)

	return filtered
}

func filter_empty_rows(entries []Expense) []Expense {

	filtered := []Expense{}

	for _, entry := range entries {

		empty := true
		if entry.Year != 0 {
			empty = false
		} else if entry.Month != 0 {
			empty = false
		} else if entry.Day != 0 {
			empty = false
		} else if entry.Description != "" {
			empty = false
		} else if entry.Debit != 0 {
			empty = false
		} else if entry.Credit != 0 {
			empty = false
		}

		if !empty {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}
