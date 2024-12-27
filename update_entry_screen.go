package main

const (
	update_search_view  = iota
	update_entries_view = iota
	update_action_view  = iota
)

type updateEntryScreenModel struct {
	fields                 [num_expense_search_fields]string
	validated              [num_expense_search_fields]bool
	feedback               string
	active_view            int
	search_cursor          int
	entry_to_search        Expense
	found_entries          []Expense
	selected_entries       []bool
	found_entries_page_idx int
	entries_cursor         int
}
