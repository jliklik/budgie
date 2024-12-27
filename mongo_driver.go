package main

import (
	"context"
	"log"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

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
