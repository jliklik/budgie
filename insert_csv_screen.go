package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

type insertCSVScreenModel struct {
	filename string
}

const InsertScreenWidth = 20

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

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			// do nothing
		case "down":
			// do nothing
		case "left":
			// do nothing
		case "right":
			// do nothing

		case "ctrl+c":
			return createHomeScreenModel(), nil

		case "backspace":
			sz := len(m.filename)
			if sz >= 1 {
				m.filename = m.filename[:sz-1]
			}

		case "enter":
			return m.enterCSV()

		default:
			m.filename += msg.String()

		}
	}

	return m, nil
}

func (m insertCSVScreenModel) View() string {
	s := selectedStyle.Width(HomeScreenWidth).Render("> Insert csv data") + "\n"
	s += textStyle.Width(InsertScreenWidth).PaddingLeft(2).Render("Enter filename:")
	s += errorStyle.PaddingLeft(2).PaddingRight(2).Render(m.filename)
	s += "\n\n" + textStyle.Width(HomeScreenWidth).PaddingLeft(2).Render("Press Ctrl+C to go back to home screen.") + "\n"
	return s
}

func (m insertCSVScreenModel) enterCSV() (tea.Model, tea.Cmd) {
	data, err := readCSV(m.filename)
	if err != nil {
		fmt.Println("Error reading file! ", err)
		return m, tea.Quit
	}
	reader, err := createCSVReader(data)
	if err != nil {
		fmt.Println("Error creating CSV reader: ", err)
		return m, tea.Quit
	}
	expenses_inserted := insertCSVIntoMongo(reader)
	insertingCsvScreenModel := createPostInsertCSVScreenModel(expenses_inserted)
	return insertingCsvScreenModel, nil
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

func insertCSVIntoMongo(reader *csv.Reader) []Expense {

	entries := []Expense{}

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

	mongoInsertEntries(entries)

	return entries
}
