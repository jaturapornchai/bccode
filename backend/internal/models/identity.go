package models

type Identity struct {
	ShopID string `json:"shopid" bson:"shopid" gorm:"column:shopid;primaryKey"`
	GuidFixed string `json:"guid_fixed" bson:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
}

type ShopIdentity struct {
	ShopID string `json:"shopid" bson:"shopid" gorm:"column:shopid;primaryKey"`
}

type DocIdentity struct {
	GuidFixed string `json:"guid_fixed" bson:"guid_fixed" gorm:"column:guid_fixed;primaryKey" `
}

type PartitionIdentity struct {
	ParID string `json:"-"  gorm:"column:parid"`
}
