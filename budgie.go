package main

import (
	"fmt"
	"context"
	// "encoding/json"
	// "log"
	// "os"

	// "github.com/joho/godotenv"
	// "go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	fmt.Println("Hello, world")
	uri := "mongodb://127.0.0.1:27017"

	// context.TODO() creates an empty context
	// options.Client().ApplyURI() is part of mongo-driver/mongo/options package 
	_, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	fmt.Println("Connected to mongodb server")
}
