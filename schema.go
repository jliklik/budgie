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
