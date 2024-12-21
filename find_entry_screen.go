package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	search_view  = iota
	entries_view = iota
)

type findEntryScreenModel struct {
	fields                 [num_expense_search_fields]string
	validated              [num_expense_search_fields]bool
	feedback               string
	active_view            int
	search_cursor          int
	entry_to_search        Expense
	found_entries          []Expense
	selected_entries       []bool
	found_entries_page_idx int
	entries_cursor         int
}

const FindEntryScreenWidth = 30
const FindEntryScreenLabelWidth = 15
const num_expense_search_fields = expense_credit + 1
const invalid = -99
const num_entries_per_page = 10

func createFindEntryScreenModel() findEntryScreenModel {
	return findEntryScreenModel{
		entry_to_search: Expense{
			Month:  invalid,
			Day:    invalid,
			Year:   invalid,
			Debit:  invalid,
			Credit: invalid,
		},
		validated:        [num_expense_search_fields]bool{false, false, false, false, false, false},
		found_entries:    nil,
		selected_entries: nil,
		feedback:         "Press Ctrl+C to go back.",
		active_view:      search_view,
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
			if m.active_view == search_view {
				if m.search_cursor > expense_year {
					m.search_cursor--
				}
			} else {
				if m.entries_cursor > 0 {
					m.entries_cursor--
				} else {
					m.active_view = search_view
				}
			}
		case "down":
			if m.active_view == search_view {
				if m.search_cursor < expense_credit {
					m.search_cursor++
				} else {
					if len(m.found_entries) > 0 {
						m.active_view = entries_view
					}
				}
			} else {
				num_entries_on_page := min(num_entries_per_page, len(m.found_entries)-(m.found_entries_page_idx*num_entries_per_page))
				if m.entries_cursor < num_entries_on_page-1 {
					m.entries_cursor++
				}
			}
		case "left":
			if m.found_entries_page_idx > 0 {
				m.found_entries_page_idx--
				m.entries_cursor = 0
			}
		case "right":
			num_pages := len(m.found_entries) / num_entries_per_page
			if m.found_entries_page_idx < num_pages {
				m.found_entries_page_idx++
				m.entries_cursor = 0
			}
		case "backspace":
			sz := len(m.fields[m.search_cursor])
			if sz >= 1 {
				m.fields[m.search_cursor] = m.fields[m.search_cursor][:sz-1]
			}
		case "enter":
			if m.active_view == search_view {
				switch m.search_cursor {
				case expense_year:
					if m.fields[m.search_cursor] != "" {
						year, err := strconv.Atoi(m.fields[m.search_cursor])
						if err == nil {
							m.entry_to_search.Year = year
							m.validated[m.search_cursor] = true
							m.search_cursor++
							m.feedback = "Press Ctrl+C to go back."
						} else {
							m.feedback = "Invalid year!"
						}
					} else {
						m.entry_to_search.Year = invalid
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = "Press Ctrl+C to go back."
					}
				case expense_month:
					if m.fields[m.search_cursor] != "" {
						month, err := time.Parse("Jan", m.fields[m.search_cursor])
						if err == nil {
							m.entry_to_search.Month = int(month.Month())
							m.validated[m.search_cursor] = true
							m.search_cursor++
							m.feedback = "Press Ctrl+C to go back."
						} else {
							m.feedback = "Invalid month! Format: Jan, Feb, Mar, etc."
						}
					} else {
						m.entry_to_search.Month = invalid
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = "Press Ctrl+C to go back."
					}
				case expense_day:
					if m.fields[m.search_cursor] != "" {
						day, err := strconv.Atoi(m.fields[m.search_cursor])
						if err == nil && day >= 1 && day <= 31 {
							m.entry_to_search.Day = day
							m.validated[m.search_cursor] = true
							m.search_cursor++
							m.feedback = "Press Ctrl+C to go back."
						} else {
							m.feedback = "Invalid day! Must be between 1 and 31."
						}
					} else {
						m.entry_to_search.Day = invalid
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = "Press Ctrl+C to go back."
					}
				case expense_description:
					m.entry_to_search.Description = m.fields[m.search_cursor]
					m.validated[m.search_cursor] = true
					m.search_cursor++
					m.feedback = "Press Ctrl+C to go back."
				case expense_debit:
					if m.fields[m.search_cursor] != "" {
						val, err := strconv.ParseFloat(m.fields[m.search_cursor], 64)
						if err == nil {
							m.entry_to_search.Debit = val
							m.validated[m.search_cursor] = true
							m.search_cursor++
							m.feedback = "Press Ctrl+C to go back."
						} else {
							m.feedback = "Invalid debit amount!"
						}
					} else {
						m.entry_to_search.Debit = invalid
						m.validated[m.search_cursor] = true
						m.search_cursor++
						m.feedback = "Press Ctrl+C to go back."
					}
				case expense_credit:
					if m.fields[m.search_cursor] != "" {
						val, err := strconv.ParseFloat(m.fields[m.search_cursor], 64)
						if err == nil {
							m.entry_to_search.Credit = val
							m.validated[m.search_cursor] = true
							m.feedback = "Press Ctrl+C to go back."
						} else {
							m.feedback = "Invalid debit amount!"
						}
					} else {
						m.entry_to_search.Credit = invalid
						m.validated[m.search_cursor] = true
						m.feedback = "Press Ctrl+C to go back."
					}
				}
				if allValid(m) {
					m.found_entries = findMatchingEntriesInMongo(m.entry_to_search)
					m.selected_entries = make([]bool, len(m.found_entries))
				}
			} else {
				if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = true
				} else {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = false
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

func (m findEntryScreenModel) View() string {
	s := renderSearchBox(m, "")
	s = renderExpenses(m, s)
	return s
}

func renderSearchBox(m findEntryScreenModel, s string) string {
	sym := " "
	if m.active_view == search_view {
		sym = "[x]"
	}

	s += textStyle.PaddingRight(1).Render("Enter in details of entry to search for. Leave blank to search all.") +
		activeViewStyle(m, search_view).Width(3).Render(sym) + "\n"
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
	s += textStyle.Render(m.feedback) + "\n"

	return s
}

func renderExpenses(m findEntryScreenModel, s string) string {
	sym := " "
	if m.active_view == entries_view {
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
		s += activeViewStyle(m, entries_view).Width(3).Render(sym)
		// s += textStyle.Width(DefaultWidth).Render("Cursor " + strconv.Itoa(m.entries_cursor))
	}

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
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Selected")
	s += "\n"

	// slice entries
	sliced_entries := m.found_entries
	sliced_selected_entries := m.selected_entries
	if len(m.found_entries) > num_entries_per_page {
		end_idx := min(len(m.found_entries), (m.found_entries_page_idx+1)*num_entries_per_page)
		sliced_entries = m.found_entries[m.found_entries_page_idx*num_entries_per_page : end_idx]
		sliced_selected_entries = m.selected_entries[m.found_entries_page_idx*num_entries_per_page : end_idx]
	}

	for i, entry := range sliced_entries {
		line := selectEntryStyle(m, i).Width(DateWidth).Render(strconv.Itoa(entry.Year))
		line += " | "
		line += selectEntryStyle(m, i).Width(DateWidth).Render(strconv.Itoa(entry.Month))
		line += " | "
		line += selectEntryStyle(m, i).Width(DateWidth).Render(strconv.Itoa(entry.Day))
		line += " | "
		line += selectEntryStyle(m, i).Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += selectEntryStyle(m, i).Width(DefaultWidth).Render(strconv.FormatFloat(entry.Debit, 'f', 2, 64))
		line += " | "
		line += selectEntryStyle(m, i).Width(DefaultWidth).Render(strconv.FormatFloat(entry.Credit, 'f', 2, 64))
		line += " | "
		selected := " "
		if sliced_selected_entries[i] {
			selected = "X"
		}
		line += fmt.Sprintf("[%s]", selected)
		s += line + "\n"
	}

	return s
}

func activeViewStyle(m findEntryScreenModel, view int) lipgloss.Style {
	if view == m.active_view {
		return selectedStyle
	}

	return textStyle
}

func selectEntryStyle(m findEntryScreenModel, index int) lipgloss.Style {
	if m.entries_cursor == index {
		return selectedStyle
	} else {
		return inactiveStyle
	}
}

func selectStyle(m findEntryScreenModel, index int) lipgloss.Style {
	if m.search_cursor == index {
		return selectedStyle.PaddingLeft(2).PaddingRight(2)
	} else if !m.validated[index] {
		return errorStyle.PaddingLeft(2).PaddingRight(2)
	} else {
		return inactiveStyle.PaddingLeft(2).PaddingRight(2)
	}
}

func allValid(m findEntryScreenModel) bool {
	for _, value := range m.validated {
		if !value {
			return false
		}
	}
	return true
}

func findMatchingEntriesInMongo(entry Expense) []Expense {
	filters := bson.A{}

	if entry.Year != invalid {
		filters = append(filters, bson.M{"year": entry.Year}) //bson.M stands for Map type
	}
	if entry.Month != invalid {
		filters = append(filters, bson.M{"month": entry.Month})
	}
	if entry.Day != invalid {
		filters = append(filters, bson.M{"day": entry.Day})
	}
	if entry.Description != "" {
		filters = append(filters, bson.M{
			"description": bson.M{
				"$regex":   ".*" + entry.Description + ".*", // Matches any string containing entry description
				"$options": "i",                             // Case-insensitive search
			}})
	}
	if entry.Debit != invalid {
		filters = append(filters, bson.M{"debit": entry.Debit})
	}
	if entry.Credit != invalid {
		filters = append(filters, bson.M{"credit": entry.Credit})
	}

	filter := bson.D{} // bson.D is a list
	if len(filters) > 0 {
		filter = bson.D{{"$and", filters}}
	}

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(MongoDb).Collection(MongoCollection)

	search_cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer search_cursor.Close(ctx)

	var expenses []Expense
	if err = search_cursor.All(ctx, &expenses); err != nil {
		panic(err)
	}

	return expenses
}
