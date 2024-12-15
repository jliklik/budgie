package main

import (
	"bytes"
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"os"

	// "encoding/json"
	// "log"

	// MongoDB
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	// TUI
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("#FAFAFA")).
	Background(lipgloss.Color("#7D56F4")).
	PaddingLeft(4).
	Width(22)

func main() {
	fmt.Println("Hello, world")
	uri := "mongodb://127.0.0.1:27017" // running this on localhost

	// context.TODO() creates an empty context
	// options.Client().ApplyURI() is part of mongo-driver/mongo/options package
	_, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to mongodb server")

	p := tea.NewProgram(createHomeScreenModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}

type homeScreenModel struct {
	choices  []string         // items on list
	cursor   int              // which item our cursor is pointing at
	selected map[int]struct{} // which items are selected
}

type insertCSVScreenModel struct {
	filename string
}

type insertingCSVScreenModel struct {
	filename string
	reader   *csv.Reader
}

const (
	insertCsvData = iota
	deleteCsvData = iota
	insertEntry   = iota
	deleteEntry   = iota
	findEntry     = iota
)

// ------------------------------------------------------------
func createInsertingCSVScreenmodel(filename string) insertingCSVScreenModel {
	return insertingCSVScreenModel{
		filename: filename,
		reader:   nil,
	}
}

type csvReaderMsg *csv.Reader

func createCsvReaderMsg(reader *csv.Reader) tea.Cmd {
	return func() tea.Msg {
		return csvReaderMsg(reader)
	}
}

func (m insertingCSVScreenModel) Init() tea.Cmd {

	fmt.Println("Trying to open filename ", m.filename)

	data, err := readCSV(m.filename)
	if err != nil {
		fmt.Println("Error reading file! ", err)
		return tea.Quit
	}
	reader, err := createCSVReader(data)
	if err != nil {
		fmt.Println("Error creating CSV reader: ", err)
		return tea.Quit
	}
	return createCsvReaderMsg(reader)
}

func (m insertingCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case csvReaderMsg:
		fmt.Println("Updating insertingCSVScreenModel with file: ", m.filename)
		m.reader = (*csv.Reader)(msg)

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c":
			return createHomeScreenModel(), nil
		}
	}

	return m, nil
}

func (m insertingCSVScreenModel) View() string {
	s := "Press Ctrl + C to go back to home screen\n\n"
	s += processCSV(m.reader)
	return s
}

// ------------------------------------------------------------
func createInsertCSVScreenModel() insertCSVScreenModel {
	return insertCSVScreenModel{
		filename: "",
	}
}

func (m insertCSVScreenModel) Init() tea.Cmd {
	return nil
}

func (m insertCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	// Is it a key press?
	case tea.KeyMsg:

		// Cool, what was the actual key pressed?
		switch msg.String() {

		// These keys should exit the program.
		case "ctrl+c":
			return createHomeScreenModel(), nil

		case "backspace":
			sz := len(m.filename)
			if sz >= 1 {
				m.filename = m.filename[:sz-1]
			}

		case "enter":
			m := createInsertingCSVScreenmodel(m.filename)
			msg := m.Init()
			return m.Update(msg)

		default:
			m.filename += msg.String()

		}
	}

	return m, nil
}

func (m insertCSVScreenModel) View() string {

	s := "Press Ctrl + C to go back to home screen\n\n Filename: "
	s += m.filename

	return s
}

// -----------------------------------------------------------
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
	s := "What would you like to do?\n\n"

	// Iterate over our choices
	for i, choice := range m.choices {

		// Is the cursor pointing at this choice?
		cursor := " " // no cursor
		if m.cursor == i {
			cursor = ">" // cursor!
		}

		// Render the row
		s += style.Render(fmt.Sprintf("%s %s\n", cursor, choice))
	}

	// The footer
	s += "\nPress q to quit.\n"

	// Send the UI for rendering
	return s
}

func readCSV(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createCSVReader(data []byte) (*csv.Reader, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	return reader, nil
}

func processCSV(reader *csv.Reader) string {

	s := ""

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("Error reading CSV data: ", err)
				break
			}
		}
		for _, str := range record {
			s += (str)
		}
		s += "\n"
	}

	return s
}
