package models

type ProductDimensionPg struct {
	HoldingCode   string `json:"holding_code" gorm:"column:holding_code;primaryKey;default:''"`
	ProductGuid   string `json:"product_guid" gorm:"column:product_guid;primaryKey"`
	DimensionGuid string `json:"dimension_guid" gorm:"column:dimension_guid;primaryKey"`
}

func (ProductDimensionPg) TableName() string {
	return "product_dimensions"
}
