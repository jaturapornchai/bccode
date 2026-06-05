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

	collection := client.Database(dbName).Collection("units")

	var results []bson.M
	cursor, err := collection.Find(ctx, bson.M{"holdingcode": "3EL6B3jlbAcZTxiMLkGMwGNzzUo"})
	if err != nil {
		log.Fatalf("find error: %v", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &results); err != nil {
		log.Fatalf("decode error: %v", err)
	}

	output, _ := json.MarshalIndent(results, "", "  ")
	fmt.Println(string(output))
}
