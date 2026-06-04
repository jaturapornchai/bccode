package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	uri := os.Getenv("MONGODB_DEV_URI")
	if uri == "" {
		log.Fatal("MONGODB_DEV_URI is required")
	}
	dbName := "bcaiclouddb"

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	defer client.Disconnect(ctx)

	// List collections to see if organizationCompanies or companies exists
	db := client.Database(dbName)
	cols, err := db.ListCollectionNames(ctx, bson.M{})
	if err != nil {
		log.Fatalf("list collections error: %v", err)
	}
	fmt.Printf("Collections: %v\n\n", cols)

	// We will try finding from organizationCompanies
	collection := db.Collection("organizationCompanies")
	var results []bson.M
	cursor, err := collection.Find(ctx, bson.M{})
	if err != nil {
		// Fallback to checking companies if organizationCompanies doesn't exist
		collection = db.Collection("companies")
		cursor, err = collection.Find(ctx, bson.M{})
		if err != nil {
			log.Fatalf("find error: %v", err)
		}
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("decode error: %v", err)
	}

	output, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(output))
}
