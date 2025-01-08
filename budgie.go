package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const MongoDb = "budgie"
const MongoCollection = "expenses"
const MongoUri = "mongodb://127.0.0.1:27017" // running this on localhost

// other constants
const default_feedback = "Press Ctrl+C to go back to home screen."
const num_expense_search_fields = expense_credit + 1
const invalid = -99
const num_entries_per_page = 10

func main() {

	p := tea.NewProgram(createHomeScreenModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
