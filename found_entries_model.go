package main

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	found_entries_view = iota
	found_action_view  = iota
	found_num_views    = iota
)

type action struct {
	action_text string
	next_model  tea.Model
}

type edit_table struct {
	cursor cursor2D
	valid  [max_entries][expense_credit + 2]int
}

type foundEntriesModel struct {
	entry_to_search        Expense
	active_view            int
	feedback               string
	found_entries          []Expense
	entries                []expensePlaceholder
	selected_entries       []bool
	found_entries_page_idx int
	entries_cursor         int
	action                 action
	edit_table             edit_table
}

func createFoundEntriesModel(found_entries []Expense, action action, entry_to_search Expense) foundEntriesModel {
	model := foundEntriesModel{
		entry_to_search:  entry_to_search,
		found_entries:    found_entries,
		selected_entries: make([]bool, len(found_entries)),
		feedback:         default_feedback,
		active_view:      found_entries_view,
		action:           action,
	}

	// TODO: populate entries with found_entries

	return populate_entries(model)
}

func populate_entries(m foundEntriesModel) foundEntriesModel {

	m.entries = make([]expensePlaceholder, len(m.found_entries))

	for idx, entry := range m.found_entries {
		m.entries[idx].Year = strconv.Itoa(entry.Year)
		m.entries[idx].Month = strconv.Itoa(entry.Month)
		m.entries[idx].Day = strconv.Itoa(entry.Day)
		m.entries[idx].Description = entry.Description
		m.entries[idx].Debit = strconv.FormatFloat(entry.Debit, 'f', 2, 64)
		m.entries[idx].Credit = strconv.FormatFloat(entry.Credit, 'f', 2, 64)
	}

	return m
}

func (m foundEntriesModel) Init() tea.Cmd {
	return nil
}

func (m foundEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			if m.active_view == found_entries_view {
				if m.entries_cursor > 0 {
					m.entries_cursor--
				}
			}
		case "down":
			if m.active_view == found_entries_view {
				num_entries_on_page := min(num_entries_per_page, len(m.found_entries)-(m.found_entries_page_idx*num_entries_per_page))
				if m.entries_cursor < num_entries_on_page-1 {
					m.entries_cursor++
				}
			}
		case "left":
			if m.action.action_text == "delete" {
				if m.found_entries_page_idx > 0 {
					m.found_entries_page_idx--
					m.entries_cursor = 0
				}
			} else {
				if m.edit_table.cursor.x > 0 {
					m.edit_table.cursor.x--
				} else {
					m.edit_table.cursor.x = expense_credit
					m.edit_table.cursor.y--
				}
			}
		case "right":
			if m.action.action_text == "delete" {
				num_pages := len(m.found_entries) / num_entries_per_page
				if m.found_entries_page_idx < num_pages {
					m.found_entries_page_idx++
					m.entries_cursor = 0
				}
			} else {
				if m.edit_table.cursor.x < expense_credit+1 {
					m.edit_table.cursor.x++
				} else {
					if m.edit_table.cursor.y < max_entries-1 {
						m.edit_table.cursor.y++
						m.edit_table.cursor.x = 0
					}
				}
			}
		case "backspace":
			if m.action.action_text == "edit" {
				entry := &m.entries[m.edit_table.cursor.y]
				switch m.edit_table.cursor.x {
				case expense_year:
					entry.Year = removeLastChar(entry.Year)
				case expense_month:
					entry.Month = removeLastChar(entry.Month)
				case expense_day:
					entry.Day = removeLastChar(entry.Day)
				case expense_description:
					entry.Description = removeLastChar(entry.Description)
				case expense_debit:
					entry.Debit = removeLastChar(entry.Debit)
				case expense_credit:
					entry.Credit = removeLastChar(entry.Credit)
				}
			}
		case "tab":
			m.active_view = (m.active_view + 1) % found_num_views
		case "x":
			// does same thing as enter for entries view
			if m.active_view == found_entries_view {
				if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = true
				} else {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = false
				}
			}
		case "enter":
			if m.action.action_text == "delete" {
				if m.active_view == found_entries_view {
					if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = true
					} else {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.entries_cursor] = false
					}
				} else {
					selected_entries := make([]Expense, 0)
					for idx, selected := range m.selected_entries {
						if selected {
							selected_entries = append(selected_entries, m.found_entries[idx])
						}
					}
					if m.action.action_text == "delete" {
						mongoDeleteEntries(selected_entries)
					}

					// reset page
					m.found_entries = mongoFindMatchingEntries(m.entry_to_search)
					m.selected_entries = make([]bool, len(m.found_entries))

					m.active_view = found_entries_view
					m.entries_cursor = 0

					m = populate_entries(m)
				}
			} else {
				if m.edit_table.cursor.x == expense_credit+1 {
					if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.edit_table.cursor.y] {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.edit_table.cursor.y] = true
					} else {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.edit_table.cursor.y] = false
					}
				}
			}
		case "ctrl+c":
			return createHomeScreenModel(), nil
		default:
			if m.action.action_text == "edit" {
				entry := &m.entries[m.edit_table.cursor.y]
				switch m.edit_table.cursor.x {
				case expense_year:
					if len(entry.Year) < DateWidth {
						entry.Year += msg.String()
					}
				case expense_month:
					if len(entry.Month) < DateWidth {
						entry.Month += msg.String()
					}
				case expense_day:
					if len(entry.Day) < DateWidth {
						entry.Day += msg.String()
					}
				case expense_description:
					if len(entry.Description) < DescriptionWidth {
						entry.Description += msg.String()
					}
				case expense_debit:
					if len(entry.Debit) < DefaultWidth {
						entry.Debit += msg.String()
					}
				case expense_credit:
					if len(entry.Credit) < DefaultWidth {
						entry.Credit += msg.String()
					}
				}
			}
		}
	}

	return m, nil
}

func (m foundEntriesModel) View() string {
	s := ""
	s = renderExpenses(m, s)
	s = renderActions(m, s)
	return s
}

func renderExpenses(m foundEntriesModel, s string) string {
	sym := " "
	if m.active_view == found_entries_view {
		sym = "[x]"
	}

	s += "\n" + textStyle.Width((DateWidth+3)*3).Render("Matching Entries")

	if len(m.found_entries) > 0 {
		page_str := "Entries: " +
			strconv.Itoa(m.found_entries_page_idx*num_entries_per_page+1) + "-" +
			strconv.Itoa(min((m.found_entries_page_idx+1)*num_entries_per_page, len(m.found_entries))) + " / " +
			strconv.Itoa(len(m.found_entries))

		s += textStyle.Width(DescriptionWidth + 3).Render(page_str)
		s += textStyle.Width((DefaultWidth + 3) * 2).Render("Press < or > to switch pages")
		s += activeViewStyle(m.active_view, found_entries_view).Width(3).Render(sym)
	}

	s += "\n" + textStyle.Render("Press tab to switch between search, entry, and delete sections.")
	s += "\n"
	s += textStyle.Width(DateWidth).Render("Year")
	s += " | "
	s += textStyle.Width(DateWidth).Render("Month")
	s += " | "
	s += textStyle.Width(DateWidth).Render("Day")
	s += " | "
	s += textStyle.Width(DescriptionWidth).Render("Description")
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Debit")
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Credit")
	s += " | "
	s += textStyle.Width(DefaultWidth).Render("Selected")
	s += "\n"

	// slice entries
	sliced_entries := m.entries
	sliced_selected_entries := m.selected_entries
	if len(m.found_entries) > num_entries_per_page {
		end_idx := min(len(m.found_entries), (m.found_entries_page_idx+1)*num_entries_per_page)
		sliced_entries = m.entries[m.found_entries_page_idx*num_entries_per_page : end_idx]
		sliced_selected_entries = m.selected_entries[m.found_entries_page_idx*num_entries_per_page : end_idx]
	}

	for row, entry := range sliced_entries {
		line := selectEntryStyle(m, row, expense_year).Width(DateWidth).Render(entry.Year)
		line += " | "
		line += selectEntryStyle(m, row, expense_month).Width(DateWidth).Render(entry.Month)
		line += " | "
		line += selectEntryStyle(m, row, expense_day).Width(DateWidth).Render(entry.Day)
		line += " | "
		line += selectEntryStyle(m, row, expense_description).Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += selectEntryStyle(m, row, expense_debit).Width(DefaultWidth).Render(entry.Debit)
		line += " | "
		line += selectEntryStyle(m, row, expense_credit).Width(DefaultWidth).Render(entry.Credit)
		line += " | "

		selected := " "
		selected_entry_style := selectEntryStyle(m, row, expense_credit+1)
		if len(sliced_selected_entries) > 0 && sliced_selected_entries[row] {
			selected = "X"
			selected_entry_style = selectedStyle
		}
		line += selected_entry_style.Render(fmt.Sprintf("[%s]", selected))
		s += line + "\n"
	}

	return s
}

func numSelectedEntries(m foundEntriesModel) int {
	num_selected := 0
	for _, selected := range m.selected_entries {
		if selected {
			num_selected++
		}
	}
	return num_selected
}

func renderActions(m foundEntriesModel, s string) string {

	if numSelectedEntries(m) > 0 {
		s += "\n" + textStyle.PaddingRight(2).Render(fmt.Sprintf("%s selected entries?", string(byte(unicode.ToUpper(rune(m.action.action_text[0]))))+
			strings.ToLower(m.action.action_text[1:])))

		sym := ""
		if m.active_view == found_action_view {
			sym = fmt.Sprintf("Press enter to %s selected entries [x]", strings.ToLower(m.action.action_text))
		}
		s += activeViewStyle(m.active_view, found_action_view).Render(sym)
	}

	return s
}

func activeViewStyle(active_view int, view int) lipgloss.Style {
	if view == active_view {
		return selectedStyle
	}

	return textStyle
}

func selectEntryStyle(m foundEntriesModel, row int, col int) lipgloss.Style {
	if m.action.action_text == "delete" {
		return selectDeleteEntryStyle(m, row)
	} else {
		return selectUpdateEntryStyle(m, row, col)
	}
}

func selectDeleteEntryStyle(m foundEntriesModel, index int) lipgloss.Style {
	if m.entries_cursor == index {
		return selectedStyle
	} else {
		return inactiveStyle
	}
}

func selectUpdateEntryStyle(m foundEntriesModel, y int, x int) lipgloss.Style {
	if m.edit_table.cursor.x == x && m.edit_table.cursor.y == y {
		return selectedStyle
	} else if m.edit_table.valid[y][x] == error_style {
		return errorStyle
	}

	return inactiveStyle
}
