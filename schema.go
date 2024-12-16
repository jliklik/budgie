package main

type Expense struct {
	month       int
	day         int
	year        int
	description string
	debit       float64
	credit      float64
	total       float64
	valid       bool
}

const (
	expense_month       = iota
	expense_day         = iota
	expense_year        = iota
	expense_description = iota
	expense_debit       = iota
	expense_credit      = iota
	expense_total       = iota
	expense_valid       = iota
)

const (
	csv_date_col        = iota
	csv_description_col = iota
	csv_debit_col       = iota
	csv_credit_col      = iota
	csv_total_col       = iota
)
