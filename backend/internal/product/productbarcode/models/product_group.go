package models

import "smlcloudplatform/internal/models"

type ProductGroup struct {
	GuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Code      string          `json:"code" bson:"code"`
	Names     *[]models.NameX `json:"names" bson:"names"`
}

type ProductGroupMessageQueueRequest struct {
	models.HoldingCodeentity `bson:"inline"`
	ProductGroup
}

func (doc ProductGroup) ToProductGroup() ProductGroup {
	return doc
}
