package models

type ProductDimensionPg struct {
	ShopID        string `json:"shopid" gorm:"column:shopid;primaryKey;default:''"`
	ProductGuid   string `json:"product_guid" gorm:"column:product_guid;primaryKey"`
	DimensionGuid string `json:"dimension_guid" gorm:"column:dimension_guid;primaryKey"`
}

func (ProductDimensionPg) TableName() string {
	return "product_dimensions"
}
