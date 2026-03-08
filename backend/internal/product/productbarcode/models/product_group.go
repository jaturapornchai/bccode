package models

import "smlcloudplatform/internal/models"

type ProductGroup struct {
	Code  string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names"`
}

type ProductGroupMessageQueueRequest struct {
	models.ShopIdentity `bson:"inline"`
	ProductGroup
}

func (doc ProductGroup) ToProductGroup() ProductGroup {
	temp := &ProductGroup{}
	temp.Code = doc.Code
	temp.Names = doc.Names
	return doc
}
