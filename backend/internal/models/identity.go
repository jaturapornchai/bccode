package models

type Identity struct {
	HoldingCode string `json:"holdingcode" bson:"holdingcode" gorm:"column:holdingcode;primaryKey"`
	GuidFixed   string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
}

type HoldingCodeentity struct {
	HoldingCode string `json:"holdingcode" bson:"holdingcode" gorm:"column:holdingcode;primaryKey"`
}

type DocIdentity struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey" `
}

type PartitionIdentity struct {
	ParID string `json:"-"  gorm:"column:parid"`
}
