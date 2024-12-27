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

type manualInsertModel struct {
	active_view int
	cursor      cursor2D
	entries     []expensePlaceholder
}

type expensePlaceholder struct {
	Month       string
	Day         string
	Year        string
	Description string
	Debit       string
	Credit      string
	Valid       bool
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
				expenses_inserted := insertManualEntriesIntoMongo(m)
				insertingCsvScreenModel := createPostInsertCSVScreenModel(expenses_inserted)
				return insertingCsvScreenModel, nil
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

func insertManualEntriesIntoMongo(m manualInsertModel) []Expense {
	entries := []Expense{}

	for row := range max_entries {

		entry := Expense{}

		for col := range expense_credit {
			switch col {
			case expense_year:
				year, err := strconv.Atoi(m.entries[row].Year)
				if err == nil {
					entry.Year = year
				}
			case expense_month:
				month, err := time.Parse("Jan", m.entries[row].Month)
				if err == nil {
					entry.Month = int(month.Month())
				} else {
					// try parsing number
					month, err := strconv.Atoi(m.entries[row].Month)
					if err == nil && month >= 1 && month <= 12 {
						entry.Month = month
					}
				}
			case expense_day:
				day, err := strconv.Atoi(m.entries[row].Day)
				if err == nil && day >= 1 && day <= 31 {
					entry.Day = day
				}
			case expense_description:
				entry.Description = m.entries[row].Description
			case expense_debit:
				val, err := strconv.ParseFloat(m.entries[row].Debit, 64)
				if err == nil {
					entry.Debit = val
				}
			case expense_credit:
				val, err := strconv.ParseFloat(m.entries[row].Credit, 64)
				if err == nil {
					entry.Credit = val
				}
			}
		}

		// Check if entry is valid
		check_if_entry_is_valid(&entry)

		entries = append(entries, entry)
	}

	mongoInsertEntries(entries)

	return entries
}
