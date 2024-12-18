package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

const MongoDb = "budgie"
const MongoCollection = "expenses"
const MongoUri = "mongodb://127.0.0.1:27017" // running this on localhost

func main() {

	p := tea.NewProgram(createHomeScreenModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}

}
