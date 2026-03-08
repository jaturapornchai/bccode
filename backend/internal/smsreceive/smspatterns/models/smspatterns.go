package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const smspatternsCollectionName = "smsPatterns"

type SmsPatterns struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string `json:"code" bson:"code"`
	Name                     string `json:"name" bson:"name"`
	Address                  string `json:"address" bson:"address"`
	Pattern                  string `json:"pattern" bson:"pattern"`
}

type SmsPatternsInfo struct {
	models.DocIdentity `bson:"inline"`
	SmsPatterns        `bson:"inline"`
}

func (SmsPatternsInfo) CollectionName() string {
	return smspatternsCollectionName
}

type SmsPatternsData struct {
	SmsPatternsInfo `bson:"inline"`
}

type SmsPatternsDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SmsPatternsData    `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (SmsPatternsDoc) CollectionName() string {
	return smspatternsCollectionName
}

type SmsPatternsItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SmsPatternsItemGuid) CollectionName() string {
	return smspatternsCollectionName
}

type SmsPatternsActivity struct {
	SmsPatternsData     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SmsPatternsActivity) CollectionName() string {
	return smspatternsCollectionName
}

type SmsPatternsDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SmsPatternsDeleteActivity) CollectionName() string {
	return smspatternsCollectionName
}
