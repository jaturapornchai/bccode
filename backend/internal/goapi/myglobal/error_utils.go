package myglobal

import (
	"smlcloudplatform/internal/goapi/logger"

	"go.mongodb.org/mongo-driver/mongo"
)

// SafeMongoConnect เป็น wrapper ที่ handle error แต่ไม่ทำให้โปรแกรมหยุด
func SafeMongoConnect() *mongo.Client {
	client, err := MongoConnect()
	if err != nil {
		logger.Error("connecting to MongoDB: %v", err)
		return nil
	}
	return client
}

// SafeMongoConnectFast เป็น wrapper ที่ handle error แต่ไม่ทำให้โปรแกรมหยุด
func SafeMongoConnectFast() *mongo.Client {
	client, err := MongoConnectFast()
	if err != nil {
		logger.Error("connecting to MongoDB: %v", err)
		return nil
	}
	return client
}
