package accountadmin

import "go.mongodb.org/mongo-driver/bson/primitive"

type Account struct {
	ID       primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name     string             `json:"name" bson:"name"`
	UserName string             `json:"username" bson:"username"`
}

func (Account) CollectionName() string {
	return "users"
}
