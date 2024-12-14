package main

import (
	"fmt"
	"context"
	"bytes"
	"encoding/csv"
	"io"
	"os"
	// "encoding/json"
	// "log"


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

	data, err := readCSV("test/jan_cc.csv")
	if err != nil {
		fmt.Println("Error reading file: ", err)
		return
	}
	reader, err := createCSVReader(data)
	if err != nil {
		fmt.Println("Error creating CSV reader: ", err)
	}
	processCSV(reader)
}

func readCSV(filename string) ([]byte, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	data, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func createCSVReader(data []byte) (*csv.Reader, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	return reader, nil
}

func processCSV(reader *csv.Reader) {
	for {
		record, err := reader.Read()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Println("Error reading CSV data: ", err)
				break
			}
		}
		fmt.Println(record)
	}
}