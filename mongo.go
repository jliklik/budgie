package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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

func checkIfEntryIsValid(entry *Expense) {
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

func mongoInsertEntries(entries []Expense) {
	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(MongoDb).Collection(MongoCollection)

	for _, entry := range entries {
		if entry.Valid {
			coll.InsertOne(ctx, entry)
		}
	}
}

func mongoFindMatchingEntries(entry Expense) []Expense {
	filters := bson.A{}

	if entry.Year != invalid {
		filters = append(filters, bson.M{"year": entry.Year}) //bson.M stands for Map type
	}
	if entry.Month != invalid {
		filters = append(filters, bson.M{"month": entry.Month})
	}
	if entry.Day != invalid {
		filters = append(filters, bson.M{"day": entry.Day})
	}
	if entry.Description != "" {
		filters = append(filters, bson.M{
			"description": bson.M{
				"$regex":   ".*" + entry.Description + ".*", // Matches any string containing entry description
				"$options": "i",                             // Case-insensitive search
			}})
	}
	if entry.Debit != invalid {
		filters = append(filters, bson.M{"debit": entry.Debit})
	}
	if entry.Credit != invalid {
		filters = append(filters, bson.M{"credit": entry.Credit})
	}

	filter := bson.D{} // bson.D is a list
	if len(filters) > 0 {
		filter = bson.D{{"$and", filters}}
	}

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(MongoDb).Collection(MongoCollection)

	search_cursor, err := coll.Find(ctx, filter)
	if err != nil {
		log.Fatal(err)
	}
	defer search_cursor.Close(ctx)

	var expenses []Expense
	if err = search_cursor.All(ctx, &expenses); err != nil {
		panic(err)
	}

	return expenses
}

func mongoUpdateEntries(old_entries []Expense, new_entries []Expense) {
	for idx, entry := range old_entries {
		mongoUpdateEntry(entry, new_entries[idx])
	}
}

func mongoUpdateEntry(old_entry Expense, new_entry Expense) {

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(MongoDb).Collection(MongoCollection)

	filter := bson.D{{"_id", old_entry.ID}}
	update := bson.D{{"$set",
		bson.D{
			{"year", new_entry.Year},
			{"month", new_entry.Month},
			{"day", new_entry.Day},
			{"description", new_entry.Description},
			{"debit", new_entry.Debit},
			{"credit", new_entry.Credit},
		}}}

	_, err = coll.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
	}
}

func mongoDeleteEntries(entries []Expense) {
	for _, entry := range entries {
		mongoDeleteEntry(entry)
	}
}

func mongoDeleteEntry(entry Expense) {

	ctx := context.TODO()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(MongoUri))
	if err != nil {
		panic(err)
	}

	defer func() {
		if err = client.Disconnect(ctx); err != nil {
			panic(err)
		}
	}()

	coll := client.Database(MongoDb).Collection(MongoCollection)

	filter := bson.M{"_id": entry.ID}

	// Delete the document
	_, err = coll.DeleteOne(ctx, filter)
	if err != nil {
		log.Fatalf("Error deleting document: %v", err)
	}
}
