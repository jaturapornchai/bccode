package shopadmin

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShopDoc struct {
	ID         primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	GuidFixed  string             `json:"shopid" bson:"guidfixed"`
	Name1      string             `json:"name1" bson:"name1"`
	Names      []models.NameX     `json:"names" bson:"names"`
	Telephone  string             `json:"telephone" bson:"telephone"`
	BranchCode string             `json:"branchcode" bson:"branchcode"`
	CreatedBy  string             `json:"createdby" bson:"createdby"`
}

func (ShopDoc) CollectionName() string {
	return "shops"
}

type ShopUserDoc struct {
	ID             primitive.ObjectID `json:"-" bson:"_id,omitempty"`
	ShopId         string             `json:"shopid" bson:"shopid"`
	Username       string             `json:"username" bson:"username"`
	Role           int                `json:"role" bson:"role"`
	LastAccessedAt time.Time          `json:"lastaccessedat" bson:"lastaccessedat"`
}

func (ShopUserDoc) CollectionName() string {
	return "shopUsers"
}
