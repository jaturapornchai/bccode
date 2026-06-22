package knowledgebase

// Knowledge Base document metadata — stored in MongoDB.
//
// RAGFlow tracks doc id/name/size/chunks/status, but BC needs additional
// per-document fields (branch, on/off, schedule). We keep them in a Mongo
// collection keyed by (holdingcode, ragflowdocid).
//
// On every list call we fetch RAGFlow's docs THEN merge with this metadata.

import (
	"context"
	"fmt"
	serviceConfig "smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const kbMetaCollection = "kbdocumentmetadata"

// KBDocMeta — bccaccount-specific metadata that lives outside RAGFlow
type KBDocMeta struct {
	HoldingCode   string    `bson:"holdingcode" json:"holdingcode"`
	RagflowDocID  string    `bson:"ragflowdocid" json:"ragflowdocid"`
	Filename      string    `bson:"filename" json:"filename"`
	BranchID      string    `bson:"branchid" json:"branchid"`
	ContentType   string    `bson:"contenttype" json:"contenttype"`
	Size          int64     `bson:"size" json:"size"`
	UploadedBy    string    `bson:"uploadedby" json:"uploadedby"`
	Description   string    `bson:"description" json:"description"`
	Tags          []string  `bson:"tags" json:"tags"`
	Status        bool      `bson:"status" json:"status"` // enabled?
	AllDay        bool      `bson:"allday" json:"allday"` // active 24/7?
	StartDateTime *string   `bson:"startdatetime,omitempty" json:"startdatetime,omitempty"`
	EndDateTime   *string   `bson:"enddatetime,omitempty" json:"enddatetime,omitempty"`
	Version       int       `bson:"version" json:"version"`
	UploadedAt    time.Time `bson:"uploadedat" json:"uploadedat"`
	UpdatedAt     time.Time `bson:"updatedat" json:"updatedat"`
}

var (
	kbMetaIndexOnce sync.Once
)

func getCollection() (*mongo.Collection, error) {
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("mongodb client not available")
	}
	cfg := serviceConfig.NewServiceConfig()
	col := mongoClient.Database(cfg.MongodbDatabaseName()).Collection(kbMetaCollection)
	kbMetaIndexOnce.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
			Keys: bson.D{
				{Key: "holdingcode", Value: 1},
				{Key: "ragflowdocid", Value: 1},
			},
			Options: options.Index().SetUnique(true),
		})
		if err != nil {
			logger.Warn("[KB Meta] index create: %v", err)
		}
	})
	return col, nil
}

// UpsertMeta inserts or replaces metadata for a doc.
func UpsertMeta(meta *KBDocMeta) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	meta.UpdatedAt = time.Now()
	if meta.UploadedAt.IsZero() {
		meta.UploadedAt = meta.UpdatedAt
	}
	if meta.Version == 0 {
		meta.Version = 1
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"holdingcode": meta.HoldingCode, "ragflowdocid": meta.RagflowDocID}
	_, err = col.ReplaceOne(ctx, filter, meta, options.Replace().SetUpsert(true))
	return err
}

// GetMetaByShop returns all metadata rows for a shop, optionally filtered by branch.
// branchID="*" or "" → no branch filter (all branches).
func GetMetaByShop(holdingCode, branchID string) ([]KBDocMeta, error) {
	col, err := getCollection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"holdingcode": holdingCode}
	if branchID != "" && branchID != "*" {
		filter["branchid"] = bson.M{"$in": []string{branchID, "*"}}
	}
	cur, err := col.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)
	var out []KBDocMeta
	if err := cur.All(ctx, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// FindByFilename — used by update/delete handlers (Flutter UI keys by filename).
func FindByFilename(holdingCode, filename string) (*KBDocMeta, error) {
	col, err := getCollection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var meta KBDocMeta
	err = col.FindOne(ctx, bson.M{"holdingcode": holdingCode, "filename": filename}).Decode(&meta)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

// UpdateFields applies a partial update to one doc by (holdingcode, filename).
func UpdateFields(holdingCode, filename string, set bson.M) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	set["updatedat"] = time.Now()
	_, err = col.UpdateOne(ctx,
		bson.M{"holdingcode": holdingCode, "filename": filename},
		bson.M{"$set": set},
	)
	return err
}

// DeleteMeta removes a row by (holdingcode, ragflowdocid).
func DeleteMeta(holdingCode, ragflowDocID string) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = col.DeleteOne(ctx, bson.M{"holdingcode": holdingCode, "ragflowdocid": ragflowDocID})
	return err
}
