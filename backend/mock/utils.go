package mock

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func MockObjectID() primitive.ObjectID {
	idx, _ := primitive.ObjectIDFromHex("62f9cb12c76fd9e83ac1b2ff")
	return idx
}

func MockHashPassword(password string) (string, error) {
	return password, nil
}

func MockCheckPasswordHash(password string, hash string) bool {
	return password == hash
}

func MockGUID() string {
	return "MOCKGUID001"
}

func MockTime() time.Time {
	timeVal, _ := time.Parse("2006-01-02 15:04:05", "2022-08-30 00:00:00")
	return timeVal
}
