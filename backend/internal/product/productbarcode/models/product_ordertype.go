package models

import "smlcloudplatform/internal/models"

type ProductOrderType struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	Price              float64         `json:"price" bson:"price"`
}

type ProductOrderTypeMessageQueueRequest struct {
	models.ShopIdentity `bson:"inline"`
	ProductOrderType
}

func (doc ProductOrderType) ToProductOrderType() ProductOrderType {
	temp := &ProductOrderType{}
	temp.Code = doc.Code
	temp.Names = doc.Names
	return doc
}
