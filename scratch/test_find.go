package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DocIdentity struct {
	GuidFixed string `bson:"guidfixed"`
}

type UnitInfo struct {
	DocIdentity `bson:"inline"`
	UnitCode    string `bson:"unitcode"`
}

type UnitData struct {
	HoldingCode string `bson:"holdingcode"`
	UnitInfo    `bson:"inline"`
}

type UnitDoc struct {
	ID       interface{} `bson:"_id"`
	UnitData `bson:"inline"`
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

	// Simulate FindByGuid
	holdingCode := "3EL6B3jlbAcZTxiMLkGMwGNzzUo"
	guid := "3EL6Bv3bDVDjiiyYZpLdQBPsLfD"

	filter := bson.M{
		"guidfixed":   guid,
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	var doc UnitDoc
	err = collection.FindOne(ctx, filter).Decode(&doc)
	if err != nil {
		fmt.Printf("FindOne error: %v\n", err)
		return
	}

	fmt.Printf("Found Document: ID=%v, HoldingCode=%s, GuidFixed=%s, UnitCode=%s\n", doc.ID, doc.HoldingCode, doc.GuidFixed, doc.UnitCode)
}
