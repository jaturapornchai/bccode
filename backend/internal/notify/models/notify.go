package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const notifyCollectionName = "notify"

type Notify struct {
	Type         string                 `json:"type" bson:"type"`
	Name         string                 `json:"name" bson:"name"`
	Options      map[string]interface{} `json:"options" bson:"options"`
	BranchEvents []NotifyBranchEvent    `json:"branchevents" bson:"branchevents"`
}

type NotifyBranchEvent struct {
	Branch           NotifyBranch `json:"branch" bson:"branch"`
	IsEnable         bool         `json:"isenable" bson:"isenable"`
	IsSaveBill       bool         `json:"issavebill" bson:"issavebill"`
	IsOutOfStock     bool         `json:"isoutofstock" bson:"isoutofstock"`
	IsNearOutOfStock bool         `json:"isnearoutofstock" bson:"isnearoutofstock"`
	IsPreorder       bool         `json:"ispreorder" bson:"ispreorder"`
}

type NotifyBranch struct {
	GuidFixed string         `json:"guidfixed" bson:"guidfixed"`
	Code      string         `json:"code" bson:"code"`
	Names     []models.NameX `json:"names" bson:"names"`
}

type NotifyInfo struct {
	models.DocIdentity `bson:"inline"`
	Token              string `json:"token" bson:"token"`
	Notify             `bson:"inline"`
}

func (NotifyInfo) CollectionName() string {
	return notifyCollectionName
}

type NotifyData struct {
	models.ShopIdentity `bson:"inline"`
	NotifyInfo          `bson:"inline"`
}

type NotifyDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	NotifyData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (NotifyDoc) CollectionName() string {
	return notifyCollectionName
}

type NotifyItemGuid struct {
	Token string `json:"token" bson:"token"`
}

func (NotifyItemGuid) CollectionName() string {
	return notifyCollectionName
}

type NotifyActivity struct {
	NotifyData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (NotifyActivity) CollectionName() string {
	return notifyCollectionName
}

type NotifyDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (NotifyDeleteActivity) CollectionName() string {
	return notifyCollectionName
}

type NotifyRequest struct {
	Token  string `json:"token" bson:"token"`
	Notify `bson:"inline"`
}
