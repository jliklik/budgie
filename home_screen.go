package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type homeScreenModel struct {
	form          *huh.Form
	lg            *lipgloss.Renderer
	styles        *Styles
	width         int
	home_action   *string
	csv_file_path *string
}

type home_action string

const (
	insertCsvData home_action = "Insert CSV data"
	insertEntry   home_action = "Insert manual entry"
	updateEntry   home_action = "Update entry"
	deleteEntry   home_action = "Delete entries"
)

func GetCurrentDirectory() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("unable to get caller information")
	}

	dir := filepath.Dir(filename)
	return dir
}

func createHomeScreenModel() homeScreenModel {
	m := homeScreenModel{
		width:         maxWidth,
		home_action:   new(string),
		csv_file_path: new(string),
	}

	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("home_action").
				Options(huh.NewOptions(string(insertCsvData), string(insertEntry), string(updateEntry), string(deleteEntry))...).
				Title("What would you like to do?").
				Description("Select action").
				Value(m.home_action),
		),

		huh.NewGroup(
			huh.NewFilePicker().
				Key("csv_file").
				Title("Select CSV file").
				CurrentDirectory(fmt.Sprintf("%s/data", GetCurrentDirectory())).
				Value(m.csv_file_path),
		).WithHideFunc(
			func() bool {
				switch home_action(*m.home_action) {
				case insertCsvData:
					return false
				default:
					return true
				}
			},
		),
	).WithWidth(maxWidth).
		WithShowHelp(false).
		WithShowErrors(false).WithTheme(huh.ThemeCatppuccin())

	return m
}

func (m homeScreenModel) Init() tea.Cmd {
	return m.form.Init()
}

func (m homeScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = min(msg.Width, maxWidth) - m.styles.Base.GetHorizontalFrameSize()
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Interrupt
		case "esc", "q":
			return m, tea.Quit
		}
	}

	var cmds []tea.Cmd

	// Process the form
	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		switch home_action(*m.home_action) {
		case insertCsvData:
			log.Printf("picked: %s", m.form.GetString("csv_file"))
			log.Printf("current directory: %s", GetCurrentDirectory())
			return EnterCSV(m)
		case insertEntry:
			return createManualInsertScreenModel(), nil
		case updateEntry:
			return createFindEntryModel(action{
				action_text: "edit",
				next_model:  nil,
			}), nil
		case deleteEntry:
			return createFindEntryModel(action{
				action_text: "delete",
				next_model:  nil,
			}), nil
		}
	}

	return m, tea.Batch(cmds...)
}

func (m homeScreenModel) View() string {

	s := m.styles

	// Form (left side)
	v := strings.TrimSuffix(m.form.View(), "\n\n")
	form := m.lg.NewStyle().Margin(1, 0).Render(v)

	errors := m.form.Errors()
	header := m.appBoundaryView("Budgie - Budget Manager v0.0.1 🪺 ")
	if len(errors) > 0 {
		header = m.appErrorBoundaryView(m.errorView())
	}

	footer := m.appBoundaryView(m.form.Help().ShortHelpView(m.form.KeyBinds()))
	if len(errors) > 0 {
		footer = m.appErrorBoundaryView("")
	}

	return s.Base.Render(header + "\n" + form + "\n\n" + footer)

}

func (m homeScreenModel) errorView() string {
	var s string
	for _, err := range m.form.Errors() {
		s += err.Error()
	}
	return s
}

func (m homeScreenModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(brown),
	)
}

func (m homeScreenModel) appErrorBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.ErrorHeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(red),
	)
}

func EnterCSV(m homeScreenModel) (tea.Model, tea.Cmd) {
	log.Printf("Reading csv: %s", m.csv_file_path)
	data, err := readCSV(m.csv_file_path)
	if err != nil {
		log.Printf("Error reading file! %v", err)
		return m, tea.Quit
	}
	reader, err := createCSVReader(data)
	if err != nil {
		log.Printf("Error creating CSV reader: %v", err)
		return m, tea.Quit
	}
	expenses_inserted := insertCSVIntoMongo(reader)
	return createPostInsertCSVScreenModel(expenses_inserted), nil
}

func readCSV(filename *string) ([]byte, error) {
	f, err := os.Open(*filename)
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

func insertCSVIntoMongo(reader *csv.Reader) []Expense {

	entries := []Expense{}

	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Printf("Error reading CSV data: %v", err)
				break
			}
		}

		entry := Expense{}

		for i, str := range record {
			switch i {
			case csv_date_col:
				// parse date
				layout := "01/02/2006"
				parsedDate, err := time.Parse(layout, str)
				if err == nil {
					entry.Month = int(parsedDate.Month())
					entry.Day = parsedDate.Day()
					entry.Year = parsedDate.Year()
				}
			case csv_description_col:
				entry.Description = str
			case csv_debit_col:
				val, err := strconv.ParseFloat(str, 64)
				if err == nil {
					entry.Debit = val
				}
			case csv_credit_col:
				val, err := strconv.ParseFloat(str, 64)
				if err == nil {
					entry.Credit = val
				}
			case csv_total_col:
				val, err := strconv.ParseFloat(str, 64)
				if err == nil {
					entry.Total = val
				}
			}
		}

		// Check if entry is valid
		checkValidEntryValues(&entry)

		entries = append(entries, entry)
	}

	// context.TODO() creates an empty context
	// options.Client().ApplyURI() is part of mongo-driver/mongo/options package

	err := mongoInsertEntries(entries)
	if err != nil {
		log.Printf("Error inserting data into mongodb: %v", err)
	}

	return entries
}
