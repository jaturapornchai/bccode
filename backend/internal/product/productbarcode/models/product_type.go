package models

import "smlcloudplatform/internal/models"

type ProductType struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type ProductTypeMessageQueueRequest struct {
	models.ShopIdentity `bson:"inline"`
	ProductType
}

func (doc ProductType) ToProductType() ProductType {
	temp := &ProductType{}
	temp.GuidFixed = doc.GuidFixed
	temp.Code = doc.Code
	temp.Names = doc.Names
	return doc
}
