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
				if m.fields[m.cursor] != "" {
					year, err := strconv.Atoi(m.fields[m.cursor])
					if err == nil {
						m.entry_to_search.year = year
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid year!"
					}
				} else {
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_month:
				if m.fields[m.cursor] != "" {
					month, err := time.Parse("Jan", m.fields[m.cursor])
					if err == nil {
						m.entry_to_search.month = int(month.Month())
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid month! Format: Jan, Feb, Mar, etc."
					}
				} else {
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_day:
				if m.fields[m.cursor] != "" {
					day, err := strconv.Atoi(m.fields[m.cursor])
					if err == nil && day >= 1 && day <= 31 {
						m.entry_to_search.day = day
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid day! Must be between 1 and 31."
					}
				} else {
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_description:
				m.entry_to_search.description = m.fields[m.cursor]
				m.validated[m.cursor] = true
				m.cursor++
				m.feedback = "Press Ctrl+C to go back."
			case expense_debit:
				if m.fields[m.cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
					if err == nil {
						m.entry_to_search.debit = val
						m.validated[m.cursor] = true
						m.cursor++
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.validated[m.cursor] = true
					m.cursor++
					m.feedback = "Press Ctrl+C to go back."
				}
			case expense_credit:
				if m.fields[m.cursor] != "" {
					val, err := strconv.ParseFloat(m.fields[m.cursor], 64)
					if err == nil {
						m.entry_to_search.debit = val
						m.validated[m.cursor] = true
						m.feedback = "Press Ctrl+C to go back."
					} else {
						m.feedback = "Invalid debit amount!"
					}
				} else {
					m.validated[m.cursor] = true
					m.feedback = "Press Ctrl+C to go back."
				}

				if allValid(m) {
					fmt.Println("all valid")
					findMatchingEntriesInMongo(m.entry_to_search)
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

func allValid(m findEntryScreenModel) bool {
	for _, value := range m.validated {
		if !value {
			return false
		}
	}
	return true
}

func findMatchingEntriesInMongo(entry Expense) {
	// filters := bson.A{}
	// if entry.year != 0 {
	// 	filters = append(filters, bson.D{{"year", entry.year}})
	// }
	// if entry.month != 0 {
	// 	filters = append(filters, bson.D{{"month", entry.month}})
	// }
	// if entry.day != 0 {
	// 	filters = append(filters, bson.D{{"day", entry.day}})
	// }
	// if entry.description != "" {
	// 	filters = append(filters, bson.D{{"description", entry.description}})
	// }
	// if entry.debit != 0 {
	// 	filters = append(filters, bson.D{{"debit", entry.debit}})
	// }
	// if entry.credit != 0 {
	// 	filters = append(filters, bson.D{{"credit", entry.credit}})
	// }

	// var filter bson.D
	// if len(filters) > 0 {
	// 	filter = bson.D{{"$and", filters}}
	// }

	filter := bson.D{
		{"$and", bson.A{
			bson.D{{"year", 2024}},
			// bson.D{{"month", 7}},
		}},
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
		fmt.Println("error")
		log.Fatal(err)
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result bson.M
		if err := cursor.Decode(&result); err != nil {
			log.Fatal(err)
		}
		fmt.Println(result)
	}

	if err := cursor.Err(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("complete")

}
