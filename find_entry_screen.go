package main

import (
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
)

type findEntryScreenModel struct {
	cursor          int
	entry_to_search Expense
	found_entries   []Expense
}

const FindEntryScreenWidth = 20

func createFindEntryScreenModel() findEntryScreenModel {
	return findEntryScreenModel{
		entry_to_search: Expense{},
		found_entries:   nil,
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
			if m.cursor < expense_valid {
				m.cursor++
			}
		case "down":
			if m.cursor > 0 {
				m.cursor--
			}
		case "left":
			// do nothing
		case "right":
			// do nothing
		case "backspace":
		case "enter":
		case "ctrl+c":
			return createHomeScreenModel(), nil
		default:
			// append to string
		}
	}

	return m, nil
}

func (m findEntryScreenModel) View() string {
	s := textStyle.Render("Enter in details of entry to search for. Leave blank to search all.") + "\n"
	s += textStyle.Render("Year: ")
	s += strconv.Itoa(m.entry_to_search.year)
	return s
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
