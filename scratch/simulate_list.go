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

type DocIdentity struct {
	GuidFixed string `bson:"guid_fixed"`
}

type UnitInfo struct {
	DocIdentity `bson:"inline"`
	UnitCode    string `bson:"unitcode"`
}

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

	holdingCode := "3EL6B3jlbAcZTxiMLkGMwGNzzUo"

	filterQuery := bson.M{
		"holding_code": holdingCode,
		"deleted_at":   bson.M{"$exists": false},
	}

	tempOptions := options.Find()
	tempOptions.SetProjection(bson.M{
		"guid_fixed":    1,
		"unitcode":      1,
		"company_guids": 1,
		"names":         1,
	})

	var docList []bson.M
	cursor, err := collection.Find(ctx, filterQuery, tempOptions)
	if err != nil {
		log.Fatalf("find error: %v", err)
	}
	defer cursor.Close(ctx)

	if err = cursor.All(ctx, &docList); err != nil {
		log.Fatalf("decode error: %v", err)
	}

	// Find if UFL is there
	for _, doc := range docList {
		if doc["unitcode"] == "UFL" {
			output, _ := json.MarshalIndent(doc, "", "  ")
			fmt.Println("Found UFL in Simulate List Query:")
			fmt.Println(string(output))
		}
	}
}
