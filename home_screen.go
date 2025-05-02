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
	form                       *huh.Form
	lg                         *lipgloss.Renderer
	styles                     *Styles
	width                      int
	home_action                *string
	csv_file_path              *string
	search_expense_year        *string
	search_expense_month       *string
	search_expense_day         *string
	search_expense_description *string
	search_expense_debit       *string
	search_expense_credit      *string
	entry_to_search            Expense
}

type home_action string

const (
	insertCsvData home_action = "Insert CSV data"
	insertEntry   home_action = "Insert manual entry"
	updateEntry   home_action = "Update entry"
	deleteEntry   home_action = "Delete entries"
	quitProgram   home_action = "Quit"
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
		width:                      maxWidth,
		home_action:                new(string),
		csv_file_path:              new(string),
		search_expense_year:        new(string),
		search_expense_month:       new(string),
		search_expense_day:         new(string),
		search_expense_description: new(string),
		search_expense_debit:       new(string),
		search_expense_credit:      new(string),
	}

	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)
	m.form = huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Key("home_action").
				Options(huh.NewOptions(string(insertCsvData), string(insertEntry), string(updateEntry), string(deleteEntry), string(quitProgram))...).
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

		huh.NewGroup(
			huh.NewInput().
				Key("search_expense_year").
				Title("Expense year").
				Prompt("> ").
				Value(m.search_expense_year).
				Validate(func(str_year string) error {
					if str_year == "" {
						return nil
					}
					_, err := strconv.Atoi(str_year)
					if err != nil {
						return err
					}
					return nil
				}),

			huh.NewSelect[string]().
				Key("search_expense_month").
				Title("Expense month").
				Options(huh.NewOptions("All", "Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec")...).
				Value(m.search_expense_month).
				Validate(func(str_month string) error {
					if str_month == "All" {
						return nil
					}
					_, err := time.Parse("Jan", str_month)
					if err == nil {
						return nil
					}
					return err
				}),

			huh.NewInput().
				Key("search_expense_day").
				Title("Expense day").
				Prompt("> ").
				Value(m.search_expense_day).
				Validate(func(str_day string) error {
					if str_day == "" {
						return nil
					}
					day, err := strconv.Atoi(str_day)
					if err == nil && day < 1 && day > 31 {
						return fmt.Errorf("day must be between 1 and 31")
					}
					return nil
				}),

			huh.NewInput().
				Key("search_description").
				Title("Expense description").
				Prompt("> ").
				Value(m.search_expense_description),

			huh.NewInput().
				Key("search_debit").
				Title("Expense debit").
				Prompt("> ").
				Value(m.search_expense_debit).
				Validate(func(v string) error {
					if v == "" {
						return nil
					}
					_, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return err
					}
					return nil
				}),

			huh.NewInput().
				Key("search_credit").
				Title("Expense credit").
				Prompt("> ").
				Value(m.search_expense_credit).
				Validate(func(v string) error {
					if v == "" {
						return nil
					}
					_, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return err
					}
					return nil
				}),
		).WithHideFunc(
			func() bool {
				switch home_action(*m.home_action) {
				case deleteEntry:
					return false
				case updateEntry:
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
			new_m := createHomeScreenModel()
			return new_m, new_m.Init()
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
			m.entry_to_search.Year = invalid
			if *m.search_expense_year != "" {
				m.entry_to_search.Year, _ = strconv.Atoi(*m.search_expense_year)
			}
			m.entry_to_search.Month = invalid
			if *m.search_expense_month != "All" {
				month, _ := time.Parse("Jan", *m.search_expense_month)
				m.entry_to_search.Month = int(month.Month())
			}
			m.entry_to_search.Day = invalid
			if *m.search_expense_day != "" {
				m.entry_to_search.Day, _ = strconv.Atoi(*m.search_expense_day)
			}
			m.entry_to_search.Description = *m.search_expense_description
			m.entry_to_search.Debit = invalid
			if *m.search_expense_debit != "" {
				m.entry_to_search.Debit, _ = strconv.ParseFloat(*m.search_expense_debit, 64)
			}
			m.entry_to_search.Credit = invalid
			if *m.search_expense_credit != "" {
				m.entry_to_search.Credit, _ = strconv.ParseFloat(*m.search_expense_credit, 64)
			}

			found_entries := mongoFindMatchingEntries(m.entry_to_search)
			log.Printf("Found %d entries", len(found_entries))
			new_m := createUpdateEntriesModel(found_entries, m.entry_to_search)
			return new_m, new_m.Init()

		case deleteEntry:
			m.entry_to_search.Year = invalid
			if *m.search_expense_year != "" {
				m.entry_to_search.Year, _ = strconv.Atoi(*m.search_expense_year)
			}
			m.entry_to_search.Month = invalid
			if *m.search_expense_month != "All" {
				month, _ := time.Parse("Jan", *m.search_expense_month)
				m.entry_to_search.Month = int(month.Month())
			}
			m.entry_to_search.Day = invalid
			if *m.search_expense_day != "" {
				m.entry_to_search.Day, _ = strconv.Atoi(*m.search_expense_day)
			}
			m.entry_to_search.Description = *m.search_expense_description
			m.entry_to_search.Debit = invalid
			if *m.search_expense_debit != "" {
				m.entry_to_search.Debit, _ = strconv.ParseFloat(*m.search_expense_debit, 64)
			}
			m.entry_to_search.Credit = invalid
			if *m.search_expense_credit != "" {
				m.entry_to_search.Credit, _ = strconv.ParseFloat(*m.search_expense_credit, 64)
			}

			found_entries := mongoFindMatchingEntries(m.entry_to_search)
			new_m := createDeleteEntriesModel(found_entries, m.entry_to_search)
			return new_m, new_m.Init()

		case quitProgram:
			return m, tea.Quit

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
