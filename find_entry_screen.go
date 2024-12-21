package main

import (
	"context"
	"log"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type findEntryScreenModel struct {
	fields                 [num_expense_search_fields]string
	validated              [num_expense_search_fields]bool
	feedback               string
	cursor                 int
	entry_to_search        Expense
	found_entries          []Expense
	found_entries_page_idx int
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
		validated:     [num_expense_search_fields]bool{false, false, false, false, false, false},
		found_entries: nil,
		feedback:      "Press Ctrl+C to go back.",
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
			if m.cursor < expense_credit {
				m.cursor++
			}
		case "left":
			if m.found_entries_page_idx > 0 {
				m.found_entries_page_idx--
			}
		case "right":
			num_pages := len(m.found_entries) / num_entries_per_page
			if m.found_entries_page_idx < num_pages {
				m.found_entries_page_idx++
			}
		case "backspace":
			sz := len(m.fields[m.cursor])
			if sz >= 1 {
				m.fields[m.cursor] = m.fields[m.cursor][:sz-1]
			}
		case "enter":
			switch m.cursor {
			case expense_year:
				if m.fields[m.cursor] != "" {
					year, err := strconv.Atoi(m.fields[m.cursor])
					if err == nil {
						m.entry_to_search.Year = year
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid year!"
					}
				} else {
					m.entry_to_search.Year = invalid
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_month:
				if m.fields[m.cursor] != "" {
					month, err := time.Parse("Jan", m.fields[m.cursor])
					if err == nil {
						m.entry_to_search.Month = int(month.Month())
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid month! Format: Jan, Feb, Mar, etc."
					}
				} else {
					m.entry_to_search.Month = invalid
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_day:
				if m.fields[m.cursor] != "" {
					day, err := strconv.Atoi(m.fields[m.cursor])
					if err == nil && day >= 1 && day <= 31 {
						m.entry_to_search.Day = day
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid day! Must be between 1 and 31."
					}
				} else {
					m.entry_to_search.Day = invalid
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_description:
				m.entry_to_search.Description = m.fields[m.cursor]
				m.validated[m.cursor] = true
				m.cursor++
				m.feedback = "Press Ctrl+C to go back."
			case expense_debit:
				if m.fields[m.cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
					if err == nil {
						m.entry_to_search.Debit = val
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.entry_to_search.Debit = invalid
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_credit:
				if m.fields[m.cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
					if err == nil {
						m.entry_to_search.Credit = val
						m.validated[m.cursor] = true
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.entry_to_search.Credit = invalid
					m.validated[m.cursor] = true
					m.feedback = "Press Ctrl+C to go back."
				}
			}
			if allValid(m) {
				m.found_entries = findMatchingEntriesInMongo(m.entry_to_search)
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
	s := renderSearchBox(m, "")
	s = renderExpenses(m, s)
	return s
}

func renderSearchBox(m findEntryScreenModel, s string) string {
	s += textStyle.Render("Enter in details of entry to search for. Leave blank to search all.") + "\n"
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

	s += "\n" + textStyle.Width((DateWidth+3)*3).Render("Matching Entries")

	if len(m.found_entries) > 0 {
		page_str := "Entries: " +
			strconv.Itoa(m.found_entries_page_idx*num_entries_per_page+1) + "-" +
			strconv.Itoa(min((m.found_entries_page_idx+1)*num_entries_per_page, len(m.found_entries))) + " / " +
			strconv.Itoa(len(m.found_entries))

		s += textStyle.Width(DescriptionWidth + 3).Render(page_str)
		s += textStyle.Width(DefaultWidth*2 + 3).Render("Press < or > to switch pages")
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
	s += "\n"

	// slice entries
	sliced_entries := m.found_entries
	if len(m.found_entries) > num_entries_per_page {
		end_idx := min(len(m.found_entries), (m.found_entries_page_idx+1)*num_entries_per_page)
		sliced_entries = m.found_entries[m.found_entries_page_idx*num_entries_per_page : end_idx]
	}

	for _, entry := range sliced_entries {
		line := inactiveStyle.Width(DateWidth).Render(strconv.Itoa(entry.Year))
		line += " | "
		line += inactiveStyle.Width(DateWidth).Render(strconv.Itoa(entry.Month))
		line += " | "
		line += inactiveStyle.Width(DateWidth).Render(strconv.Itoa(entry.Day))
		line += " | "
		line += inactiveStyle.Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += inactiveStyle.Width(DefaultWidth).Render(strconv.FormatFloat(entry.Debit, 'f', 2, 64))
		line += " | "
		line += inactiveStyle.Width(DefaultWidth).Render(strconv.FormatFloat(entry.Credit, 'f', 2, 64))
		s += line + "\n"
	}

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

	cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	var expenses []Expense
	if err = cursor.All(ctx, &expenses); err != nil {
		panic(err)
	}

	return expenses
}
