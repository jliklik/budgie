package main

import "go.mongodb.org/mongo-driver/bson/primitive"

// Fields have to start with capital letter or else they
// will not be properly entered into MongoDB!
type Expense struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Month       int                `bson:"month"`
	Day         int                `bson:"day"`
	Year        int                `bson:"year"`
	Description string             `bson:"description"`
	Debit       float64            `bson:"debit"`
	Credit      float64            `bson:"credit"`
	Total       float64            `bson:"total,omitempty"`
	Valid       bool               `bson:"valid,omitempty"`
}

// could use reflection, but mapping struct fields to index is clearer
const (
	expense_year        = iota
	expense_month       = iota
	expense_day         = iota
	expense_description = iota
	expense_debit       = iota
	expense_credit      = iota
	expense_total       = iota
	expense_valid       = iota
	num_expense_fields  = iota
)

const (
	csv_date_col        = iota
	csv_description_col = iota
	csv_debit_col       = iota
	csv_credit_col      = iota
	csv_total_col       = iota
)

func check_if_entry_is_valid(entry *Expense) {
	if entry.Month == 0 {
		return
	} else if entry.Day == 0 {
		return
	} else if entry.Year == 0 {
		return
	} else if entry.Description == "" {
		return
	} else if entry.Debit == 0 && entry.Credit == 0 {
		return
	}

	entry.Valid = true
}
