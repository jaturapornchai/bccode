package knowledgebase

// Knowledge Base document metadata — stored in MongoDB.
//
// RAGFlow tracks doc id/name/size/chunks/status, but BC needs additional
// per-document fields (branch, on/off, schedule). We keep them in a Mongo
// collection keyed by (holding_code, ragflow_doc_id).
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

const kbMetaCollection = "kbDocumentMetadata"

// KBDocMeta — bccaccount-specific metadata that lives outside RAGFlow
type KBDocMeta struct {
	HoldingCode   string    `bson:"holding_code" json:"holding_code"`
	RagflowDocID  string    `bson:"ragflow_doc_id" json:"ragflow_doc_id"`
	Filename      string    `bson:"file_name" json:"file_name"`
	BranchID      string    `bson:"branch_id" json:"branch_id"`
	ContentType   string    `bson:"content_type" json:"content_type"`
	Size          int64     `bson:"size" json:"size"`
	UploadedBy    string    `bson:"uploaded_by" json:"uploaded_by"`
	Description   string    `bson:"description" json:"description"`
	Tags          []string  `bson:"tags" json:"tags"`
	Status        bool      `bson:"status" json:"status"`  // enabled?
	AllDay        bool      `bson:"allday" json:"all_day"` // active 24/7?
	StartDateTime *string   `bson:"start_date_time,omitempty" json:"start_date_time,omitempty"`
	EndDateTime   *string   `bson:"end_date_time,omitempty" json:"end_date_time,omitempty"`
	Version       int       `bson:"version" json:"version"`
	UploadedAt    time.Time `bson:"uploaded_at" json:"uploaded_at"`
	UpdatedAt     time.Time `bson:"updated_at" json:"updated_at"`
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
				{Key: "holding_code", Value: 1},
				{Key: "ragflow_doc_id", Value: 1},
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
	filter := bson.M{"holding_code": meta.HoldingCode, "ragflow_doc_id": meta.RagflowDocID}
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
	filter := bson.M{"holding_code": holdingCode}
	if branchID != "" && branchID != "*" {
		filter["branch_id"] = bson.M{"$in": []string{branchID, "*"}}
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
	err = col.FindOne(ctx, bson.M{"holding_code": holdingCode, "file_name": filename}).Decode(&meta)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

// UpdateFields applies a partial update to one doc by (holding_code, filename).
func UpdateFields(holdingCode, filename string, set bson.M) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	set["updated_at"] = time.Now()
	_, err = col.UpdateOne(ctx,
		bson.M{"holding_code": holdingCode, "file_name": filename},
		bson.M{"$set": set},
	)
	return err
}

// DeleteMeta removes a row by (holding_code, ragflow_doc_id).
func DeleteMeta(holdingCode, ragflowDocID string) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = col.DeleteOne(ctx, bson.M{"holding_code": holdingCode, "ragflow_doc_id": ragflowDocID})
	return err
}
