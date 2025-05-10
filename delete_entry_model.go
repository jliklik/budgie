// This model is shared between delete and update views
// probably not the best idea after all...

package main

import (
	"fmt"
	"log"
	"slices"
	"strconv"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	delete_entries_view    = iota
	delete_select_all_view = iota
	delete_action_view     = iota
	delete_num_views       = iota
)

type DeleteEntriesModel struct {
	entry_to_search        Expense
	active_view            int
	feedback               string
	found_entries          []Expense
	entries                []ExpensePlaceholder
	found_entries_page_idx int
	selected_entries       []bool
	cursor                 Cursor2D
	prompt_text            string
}

func createDeleteEntriesModel(found_entries []Expense, entry_to_search Expense) DeleteEntriesModel {
	model := DeleteEntriesModel{
		entry_to_search:  entry_to_search,
		found_entries:    found_entries,
		selected_entries: make([]bool, len(found_entries)),
		feedback:         default_feedback,
		active_view:      delete_entries_view,
		prompt_text:      default_feedback,
	}

	return populateDeleteEntries(model)
}

func populateDeleteEntries(m DeleteEntriesModel) DeleteEntriesModel {

	log.Printf("update form number of entries: %d", len(m.found_entries))

	m.entries = make([]ExpensePlaceholder, len(m.found_entries))

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

func (m DeleteEntriesModel) Init() tea.Cmd {
	return nil
}

func (m DeleteEntriesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:

		switch msg.String() {

		case "up":
			if m.active_view == delete_entries_view {
				if m.cursor.y > 0 {
					m.cursor.y--
				}
			}

		case "down":
			if m.active_view == delete_entries_view {
				num_entries_on_page := min(num_entries_per_page, len(m.found_entries)-(m.found_entries_page_idx*num_entries_per_page))
				if m.cursor.y < num_entries_on_page-1 {
					m.cursor.y++
				}
			}

		case "left":
			if m.found_entries_page_idx > 0 {
				m.found_entries_page_idx -= 1
			}

		case "right":
			if (m.found_entries_page_idx+1)*max_entries < len(m.entries) {
				m.found_entries_page_idx += 1
			}

		case "tab":
			m.active_view = (m.active_view + 1) % delete_num_views
			if m.active_view == delete_action_view && !slices.Contains(m.selected_entries, true) {
				log.Print("skipping delete action view")
				m.active_view = delete_entries_view
			}
		case "x":
			// toggle selected entries
			if m.active_view == delete_entries_view {
				if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] = true
				} else {
					m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] = false
				}
			}
		case "enter":
			if m.active_view == delete_entries_view {
				if len(m.entries) > 0 {
					if !m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] = true
					} else {
						m.selected_entries[m.found_entries_page_idx*num_entries_per_page+m.cursor.y] = false
					}
				}
			} else if m.active_view == delete_select_all_view {
				for i := 0; i < len(m.selected_entries); i++ {
					m.selected_entries[i] = true
				}
			} else { // action view
				selected_entries := make([]Expense, 0)
				for idx, selected := range m.selected_entries {
					if selected {
						selected_entries = append(selected_entries, m.found_entries[idx])
					}
				}

				mongoDeleteEntries(selected_entries)

				// reset page
				m.found_entries = mongoFindMatchingEntries(m.entry_to_search)
				m.selected_entries = make([]bool, len(m.found_entries))

				m.active_view = delete_entries_view
				m.cursor.y = 0

				m = populateDeleteEntries(m)
			}

		case "ctrl+c":
			new_m := createHomeScreenModel()
			return new_m, new_m.Init()
		default:
			// do nothing in delete view
		}
	}

	return m, nil
}

func (m DeleteEntriesModel) View() string {
	s := ""
	s = renderDeleteExpenses(m, s)
	s += textStyle.Render(m.prompt_text) + "\n"
	s = renderDeleteSelectAll(m, s)
	s = renderDeleteActions(m, s)
	return s
}

func renderDeleteExpenses(m DeleteEntriesModel, s string) string {
	sym := " "
	if m.active_view == delete_entries_view {
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
		s += activeDeleteViewStyle(m.active_view, delete_entries_view).Width(3).Render(sym)
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

	start_idx := m.found_entries_page_idx * max_entries
	end_idx := min(start_idx+max_entries, len(m.entries))
	sliced_entries := m.entries[start_idx:end_idx]

	sliced_selected_entries := m.selected_entries[start_idx:end_idx]

	for row, entry := range sliced_entries {
		line := selectDeleteEntryStyle(m, row).Width(DateWidth).Render(entry.Year)
		line += " | "
		line += selectDeleteEntryStyle(m, row).Width(DateWidth).Render(entry.Month)
		line += " | "
		line += selectDeleteEntryStyle(m, row).Width(DateWidth).Render(entry.Day)
		line += " | "
		line += selectDeleteEntryStyle(m, row).Width(DescriptionWidth).Render(entry.Description)
		line += " | "
		line += selectDeleteEntryStyle(m, row).Width(DefaultWidth).Render(entry.Debit)
		line += " | "
		line += selectDeleteEntryStyle(m, row).Width(DefaultWidth).Render(entry.Credit)
		line += " | "

		selected := " "
		selected_entry_style := selectDeleteEntryStyle(m, row)
		// check len to prevent erroring if no entries found
		if len(sliced_selected_entries) > 0 && sliced_selected_entries[row] {
			selected = "X"
			selected_entry_style = selectedStyle
		}
		line += selected_entry_style.Render(fmt.Sprintf("[%s]", selected))
		s += line + "\n"
	}

	return s
}

func numDeleteSelectedEntries(m DeleteEntriesModel) int {
	num_selected := 0
	for _, selected := range m.selected_entries {
		if selected {
			num_selected++
		}
	}
	return num_selected
}

func renderDeleteSelectAll(m DeleteEntriesModel, s string) string {

	s += "\n" + textStyle.PaddingRight(2).Render("Select all entries?")

	sym := ""
	if m.active_view == delete_select_all_view {
		sym = "Press enter to select all entries [x]"
	}
	s += activeDeleteViewStyle(m.active_view, delete_select_all_view).Render(sym) + "\n"

	return s
}

func renderDeleteActions(m DeleteEntriesModel, s string) string {

	if numDeleteSelectedEntries(m) > 0 {
		s += "\n" + textStyle.PaddingRight(2).Render("Delete selected entries?")

		sym := ""
		if m.active_view == delete_action_view {
			sym = "Press enter to delete selected entries [x]"
		}
		s += activeDeleteViewStyle(m.active_view, delete_action_view).Render(sym) + "\n"
	}

	return s
}

func activeDeleteViewStyle(active_view int, view int) lipgloss.Style {
	if view == active_view {
		return selectedStyle
	}

	return textStyle
}

// highlights entire row
func selectDeleteEntryStyle(m DeleteEntriesModel, row int) lipgloss.Style {
	if m.cursor.y == row {
		return selectedStyle
	} else {
		return inactiveStyle
	}
}
