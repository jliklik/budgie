package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type homeScreenModel struct {
	choices  []string         // items on list
	cursor   int              // which item our cursor is pointing at
	selected map[int]struct{} // which items are selected
}

const HomeScreenWidth = 30

func createHomeScreenModel() homeScreenModel {
	return homeScreenModel{
		choices:  []string{"Insert csv data", "Delete csv data", "Insert entry", "Delete entry", "Find entry"},
		selected: make(map[int]struct{}), // map of int to struct
	}
}

func (m homeScreenModel) Init() tea.Cmd {
	return nil
}

func (m homeScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c", "q":
			return m, tea.Quit

		// The "up" and "k" keys move the cursor up
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}

		// The "down" and "j" keys move the cursor down
		case "down", "j":
			if m.cursor < len(m.choices)-1 {
				m.cursor++
			}

		// The "enter" key and the spacebar (a literal space) toggle
		// the selected state for the item that the cursor is pointing at.
		case "enter", " ":

			switch m.cursor {
			case insertCsvData:
				return createInsertCSVScreenModel(), nil
			case deleteCsvData:
			case insertEntry:
			case deleteEntry:
			case findEntry:
			}

			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
			} else {
				m.selected[m.cursor] = struct{}{}
				fmt.Println(m.selected)
			}
		}
	}

	// Return the updated model to the Bubble Tea runtime for processing.
	// Note that we're not returning a command.
	return m, nil
}

func (m homeScreenModel) View() string {
	// The header
	s := textStyle.Width(HomeScreenWidth).PaddingLeft(2).Render("What would you like to do?") + "\n"

	content := ""
	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
			content += selectedStyle.Width(HomeScreenWidth).Render(fmt.Sprintf("%s %s", cursor, choice))
		} else {
			content += promptStyle.Width(HomeScreenWidth).Render(fmt.Sprintf("%s %s", cursor, choice))
		}

		if i < len(m.choices)-1 {
			content += "\n"
		}
	}

	s += promptStyle.Width(HomeScreenWidth).Render(content)

	// The footer
	s += "\n" + textStyle.Width(HomeScreenWidth).PaddingLeft(2).Render("Press q to quit.") + "\n"

	// Send the UI for rendering
	return s
}
