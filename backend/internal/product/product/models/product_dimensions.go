package models

type ProductDimensionPg struct {
	HoldingCode   string `json:"holdingcode" gorm:"column:holdingcode;primaryKey;default:''"`
	ProductGuid   string `json:"productguid" gorm:"column:productguid;primaryKey"`
	DimensionGuid string `json:"dimensionguid" gorm:"column:dimensionguid;primaryKey"`
}

func (ProductDimensionPg) TableName() string {
	return "product_dimensions"
}
