package models

type Identity struct {
	ShopID    string `json:"shopid" bson:"shopid" gorm:"column:shopid;primaryKey"`
	GuidFixed string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
}

type ShopIdentity struct {
	ShopID string `json:"shopid" bson:"shopid" gorm:"column:shopid;primaryKey"`
}

type DocIdentity struct {
	GuidFixed string `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey" `
}

type PartitionIdentity struct {
	ParID string `json:"-"  gorm:"column:parid"`
}
