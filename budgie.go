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
	Width(30).
	Align(lipgloss.Left)

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

const (
	insertCsvData = iota
	deleteCsvData = iota
	insertEntry   = iota
	deleteEntry   = iota
	findEntry     = iota
)

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
