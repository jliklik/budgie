// Select month, and process data for that month

package main

import (
	"regexp"
	"strconv"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ProcessEntriesModel struct {
	entry_to_search         Expense
	found_entries           []Expense
	summary                 Summary
	finished_processing     bool
	gas_and_groceries_table table.Model
	lg                      *lipgloss.Renderer
	styles                  *Styles
	width                   int
}

func createProcessEntriesModel(found_entries []Expense, entry_to_search Expense) ProcessEntriesModel {
	m := ProcessEntriesModel{
		entry_to_search: entry_to_search,
		found_entries:   found_entries,
		width:           maxWidth,
	}

	m.lg = lipgloss.DefaultRenderer()
	m.styles = NewStyles(m.lg)

	return m
}

func gasAndGroceryStrings() []string {
	return []string{
		"superstore",
		"supermarket",
		"chv", // chevron
		"super save",
		"flashfood",
		"FREEDOM MOBILE",
		"TELUS",
	}
}

func insuranceStrings() []string {
	return []string{
		"square one",
	}
}

func incomeStrings() []string {
	return []string{
		"CORVUS",
		"TAX REFUND",
		"REWARDS_REDEMPTION",
	}
}

func rentalIncomeStrings() []string {
	return []string{
		"E-TRANSFER",
	}
}

func donationStrings() []string {
	return []string{
		"CHRIST CITY CHURCH",
		"UNION GOSPEL",
		"UGM",
		"BC CANCER",
	}
}

func ignoreStrings() []string {
	return []string{
		"STEADYHAND",
		"PAYMENT - THANK YOU",
	}
}

func processEntries(m *ProcessEntriesModel) {

outer:
	for _, entry := range m.found_entries {
		for _, regex := range gasAndGroceryStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				m.summary.GasAndGroceries = append(m.summary.GasAndGroceries, entry)
				break outer
			}
		}
		for _, regex := range insuranceStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				m.summary.Insurance = append(m.summary.Insurance, entry)
				break outer
			}
		}
		for _, regex := range incomeStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				m.summary.Income = append(m.summary.Income, entry)
				break outer
			}
		}
		for _, regex := range rentalIncomeStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				if entry.Debit > 1000.0 {
					m.summary.RentalIncome = append(m.summary.RentalIncome, entry)
					break outer
				}
			}
		}
		for _, regex := range donationStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				m.summary.Donation = append(m.summary.Donation, entry)
				break outer
			}
		}
		for _, regex := range ignoreStrings() {
			regex_search := "(?i)" + regex

			if res, err := regexp.MatchString(regex_search, entry.Description); err != nil && res {
				// do nothing
				break outer
			}
		}
		// otherwise, assume entertainment
		m.summary.Entertainment = append(m.summary.Entertainment, entry)
	}

	columns := []table.Column{
		{Title: "Year", Width: DateWidth},
		{Title: "Month", Width: DateWidth},
		{Title: "Day", Width: DateWidth},
		{Title: "Description", Width: DescriptionWidth},
		{Title: "Debit", Width: DefaultWidth},
		{Title: "Credit", Width: DefaultWidth},
	}

	// convert expense into "row"
	rows := []table.Row{}

	total := 0.0
	for _, e := range m.summary.GasAndGroceries {
		total = total + e.Debit - e.Credit
		rows = append(rows, table.Row{
			strconv.Itoa(e.Year),
			strconv.Itoa(e.Month),
			strconv.Itoa(e.Day),
			e.Description,
			strconv.FormatFloat(e.Debit, 'f', 2, 64),
			strconv.FormatFloat(e.Credit, 'f', 2, 64)})
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(7),
	)

	m.gas_and_groceries_table = t
}

func (m ProcessEntriesModel) Init() tea.Cmd {
	return nil
}

func (m ProcessEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {

	if !m.finished_processing {
		processEntries(&m)
		m.finished_processing = true
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:

		switch msg.String() {

		case "ctrl+c":
			new_m := createHomeScreenModel()
			return new_m, new_m.Init()
		default:
		}

	}
	return m, nil
}

func (m ProcessEntriesModel) View() string {
	s := m.styles
	body := s.Table.Render(m.gas_and_groceries_table.View())

	header := m.appBoundaryView("Budgie - Budget Manager v0.0.1 🪺 ")

	footer := m.appBoundaryView("Press Ctrl+C to return to home screen")

	return s.Base.Render(header + "\n" + body + "\n\n" + footer)
}

func (m ProcessEntriesModel) appBoundaryView(text string) string {
	return lipgloss.PlaceHorizontal(
		m.width,
		lipgloss.Left,
		m.styles.HeaderText.Render(text),
		lipgloss.WithWhitespaceChars("/"),
		lipgloss.WithWhitespaceForeground(brown),
	)
}
