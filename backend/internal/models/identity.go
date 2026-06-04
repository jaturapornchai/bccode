package models

type Identity struct {
	HoldingCode string `json:"holding_code" bson:"holding_code" gorm:"column:holding_code;primaryKey"`
	GuidFixed   string `json:"guid_fixed" bson:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
}

type HoldingCodeentity struct {
	HoldingCode string `json:"holding_code" bson:"holding_code" gorm:"column:holding_code;primaryKey"`
}

type DocIdentity struct {
	GuidFixed string `json:"guid_fixed" bson:"guid_fixed" gorm:"column:guid_fixed;primaryKey" `
}

type PartitionIdentity struct {
	ParID string `json:"-"  gorm:"column:parid"`
}
