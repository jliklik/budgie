package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
)

type insertCSVScreenModel struct {
	filename string
}

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

		case "ctrl+c":
			return createHomeScreenModel(), nil

		case "backspace":
			sz := len(m.filename)
			if sz >= 1 {
				m.filename = m.filename[:sz-1]
			}

		case "enter":
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
			m := createInsertingCSVScreenmodel(m.filename, reader)
			return m, nil

		default:
			m.filename += msg.String()

		}
	}

	return m, nil
}

func (m insertCSVScreenModel) View() string {

	s := "Press Ctrl + C to go back to home screen\n\n Filename: "
	s += style.Render(m.filename)

	return s
}
