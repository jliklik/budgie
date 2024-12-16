package main

import (
	"encoding/csv"
	"fmt"
	"io"

	tea "github.com/charmbracelet/bubbletea"
)

type insertingCSVScreenModel struct {
	filename string
	reader   *csv.Reader
}

// type Book struct {
// 	Title  string
// 	Author string
// }

const CsvEntryWidth = 15
const DescriptionWidth = 30

func createInsertingCSVScreenmodel(filename string, reader *csv.Reader) insertingCSVScreenModel {
	return insertingCSVScreenModel{
		filename: filename,
		reader:   reader,
	}
}

func (m insertingCSVScreenModel) Init() tea.Cmd {
	return nil
}

func (m insertingCSVScreenModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

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

func processCSV(reader *csv.Reader) string {

	s := ""

	header := true

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

		line := ""
		for i, str := range record {
			width := CsvEntryWidth
			if i == 1 {
				width = DescriptionWidth
			}
			if header {
				line += textStyle.Width(width).Render(str)
			} else {
				line += selectedStyle.Width(width).Render(str)
			}

			line += " | "
		}
		// s += csvLineStyle.Render(line) + "\n"
		header = false
		s += line + "\n"
	}

	return s
}
