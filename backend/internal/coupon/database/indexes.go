package database

import (
	"context"
	"log"
	"smlcloudplatform/internal/coupon/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// createNonUniqueIndex สร้าง non-unique index สำหรับ MongoDB
func createNonUniqueIndex(pst microservice.IPersisterMongo, model interface{}, indexName string, keys bson.D) error {
	db, err := pst.Exec(context.Background(), model)
	if err != nil {
		return err
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetName(indexName), // ไม่ตั้งค่า SetUnique(true)
	}

	_, err = db.Indexes().CreateOne(context.Background(), indexModel)
	return err
}

// dropIndexIfExists ลบ index หากมีอยู่แล้ว
func dropIndexIfExists(pst microservice.IPersisterMongo, model interface{}, indexName string) error {
	db, err := pst.Exec(context.Background(), model)
	if err != nil {
		return err
	}

	// ลองลบ index (ไม่ return error หากไม่มี index)
	_, _ = db.Indexes().DropOne(context.Background(), indexName)
	return nil
}

// CreateCouponReservationIndexes สร้าง indexes สำหรับ collection coupon_reservations
func CreateCouponReservationIndexes(pst microservice.IPersisterMongo) error {
	ctx := context.Background()

	// สร้างแบบจำลองเพื่อให้ได้ collection name
	model := &models.CouponReservation{}

	// ลบ index เก่าที่ผิดพลาด (unique index)
	problemIndexes := []string{
		"idx_shop_customer_status", // index เก่าที่ทำให้เกิด duplicate key error
	}

	for _, indexName := range problemIndexes {
		err := dropIndexIfExists(pst, model, indexName)
		if err != nil {
			log.Printf("Note: Could not drop index %s (might not exist): %v", indexName, err)
		} else {
			log.Printf("Dropped problematic index: %s", indexName)
		}
	}

	// สร้าง indexes โดยใช้ CreateIndex method (unique indexes)
	uniqueIndexes := []struct {
		name string
		keys bson.D
	}{
		{
			name: "idx_status_expires_at",
			keys: bson.D{
				{Key: "status", Value: 1},
				{Key: "expires_at", Value: 1},
			},
		},
		{
			name: "idx_shop_coupon_status_expires",
			keys: bson.D{
				{Key: "shopid", Value: 1},
				{Key: "coupon_id", Value: 1},
				{Key: "status", Value: 1},
				{Key: "expires_at", Value: 1},
			},
		},
		{
			name: "idx_transaction_id",
			keys: bson.D{
				{Key: "transaction_id", Value: 1},
			},
		},
	}

	// สร้าง unique indexes
	for _, index := range uniqueIndexes {
		indexName, err := pst.CreateIndex(ctx, model, index.name, index.keys)
		if err != nil {
			log.Printf("Error creating unique index %s: %v", index.name, err)
			continue
		}
		log.Printf("Created unique index: %s", indexName)
	}

	// สร้าง non-unique indexes
	nonUniqueIndexes := []struct {
		name string
		keys bson.D
	}{
		{
			name: "idx_shop_customer_nonunique",
			keys: bson.D{
				{Key: "shopid", Value: 1},
				{Key: "customer_id", Value: 1},
			},
		},
		{
			name: "idx_shop_customer_status_nonunique",
			keys: bson.D{
				{Key: "shopid", Value: 1},
				{Key: "customer_id", Value: 1},
				{Key: "status", Value: 1},
			},
		},
	}

	// สร้าง non-unique indexes
	for _, index := range nonUniqueIndexes {
		err := createNonUniqueIndex(pst, model, index.name, index.keys)
		if err != nil {
			log.Printf("Error creating non-unique index %s: %v", index.name, err)
			continue
		}
		log.Printf("Created non-unique index: %s", index.name)
	}

	return nil
}
