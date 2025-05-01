package main

import (
	"fmt"
	"log"
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

const interface_log_file = "budgie.log"

type valid_status int

const (
	valid_inactive valid_status = iota
	valid_error    valid_status = iota
	valid_selected valid_status = iota
)

type Cursor2D struct {
	x int
	y int
}

type TrackEditsTable struct {
	modified [][]bool
	valid    [][]valid_status
}

type ExpensePlaceholder struct {
	Month       string
	Day         string
	Year        string
	Description string
	Debit       string
	Credit      string
}

func main() {

	file, err := os.OpenFile(interface_log_file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Failed to open log file: %v", err)
		os.Exit(1)
	}
	defer file.Close()

	// Set log output to the file
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetOutput(file)

	p := tea.NewProgram(createHomeScreenModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
