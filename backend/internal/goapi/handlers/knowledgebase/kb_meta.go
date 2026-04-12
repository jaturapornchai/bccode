package knowledgebase

// Knowledge Base document metadata — stored in MongoDB.
//
// RAGFlow tracks doc id/name/size/chunks/status, but BC needs additional
// per-document fields (branch, on/off, schedule). We keep them in a Mongo
// collection keyed by (shop_id, ragflow_doc_id).
//
// On every list call we fetch RAGFlow's docs THEN merge with this metadata.

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	serviceConfig "smlcloudplatform/internal/goapi/config"
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
	ShopID        string    `bson:"shopid" json:"shop_id"`
	RagflowDocID  string    `bson:"ragflowdocid" json:"ragflow_doc_id"`
	Filename      string    `bson:"filename" json:"filename"`
	BranchID      string    `bson:"branchid" json:"branch_id"`
	ContentType   string    `bson:"contenttype" json:"content_type"`
	Size          int64     `bson:"size" json:"size"`
	UploadedBy    string    `bson:"uploadedby" json:"uploaded_by"`
	Description   string    `bson:"description" json:"description"`
	Tags          []string  `bson:"tags" json:"tags"`
	Status        bool      `bson:"status" json:"status"`     // enabled?
	AllDay        bool      `bson:"allday" json:"all_day"`    // active 24/7?
	StartDateTime *string   `bson:"startdatetime,omitempty" json:"start_date_time,omitempty"`
	EndDateTime   *string   `bson:"enddatetime,omitempty" json:"end_date_time,omitempty"`
	Version       int       `bson:"version" json:"version"`
	UploadedAt    time.Time `bson:"uploadedat" json:"uploaded_at"`
	UpdatedAt     time.Time `bson:"updatedat" json:"updated_at"`
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
				{Key: "shopid", Value: 1},
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
	filter := bson.M{"shopid": meta.ShopID, "ragflowdocid": meta.RagflowDocID}
	_, err = col.ReplaceOne(ctx, filter, meta, options.Replace().SetUpsert(true))
	return err
}

// GetMetaByShop returns all metadata rows for a shop, optionally filtered by branch.
// branchID="*" or "" → no branch filter (all branches).
func GetMetaByShop(shopID, branchID string) ([]KBDocMeta, error) {
	col, err := getCollection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	filter := bson.M{"shopid": shopID}
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
func FindByFilename(shopID, filename string) (*KBDocMeta, error) {
	col, err := getCollection()
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	var meta KBDocMeta
	err = col.FindOne(ctx, bson.M{"shopid": shopID, "filename": filename}).Decode(&meta)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}
	return &meta, nil
}

// UpdateFields applies a partial update to one doc by (shop_id, filename).
func UpdateFields(shopID, filename string, set bson.M) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	set["updatedat"] = time.Now()
	_, err = col.UpdateOne(ctx,
		bson.M{"shopid": shopID, "filename": filename},
		bson.M{"$set": set},
	)
	return err
}

// DeleteMeta removes a row by (shop_id, ragflow_doc_id).
func DeleteMeta(shopID, ragflowDocID string) error {
	col, err := getCollection()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err = col.DeleteOne(ctx, bson.M{"shopid": shopID, "ragflowdocid": ragflowDocID})
	return err
}
