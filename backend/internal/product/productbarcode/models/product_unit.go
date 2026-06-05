package models

import "smlcloudplatform/internal/models"

type ProductUnit struct {
	UnitCode string          `json:"unitcode" bson:"unitcode"`
	Names    *[]models.NameX `json:"names" bson:"names"`
}

type ProductUnitMessageQueueRequest struct {
	models.HoldingCodeentity `bson:"inline"`
	ProductUnit
}

func (doc ProductUnit) ToProductUnit() ProductUnit {
	temp := &ProductUnit{}
	temp.UnitCode = doc.UnitCode
	temp.Names = doc.Names
	return doc
}
