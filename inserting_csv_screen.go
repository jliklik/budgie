package main

import (
	"encoding/csv"

	tea "github.com/charmbracelet/bubbletea"
)

type insertingCSVScreenModel struct {
	filename string
	reader   *csv.Reader
}

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
