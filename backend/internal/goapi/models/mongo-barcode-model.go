package models

type ProcessMongoBarcodeModel struct {
	ShopID        string                               `json:"shopid" bson:"shopid"`
	ItemCode      string                               `json:"itemcode" bson:"itemcode"`
	Barcode       string                               `json:"barcode" bson:"barcode"`
	ItemType      int                                  `json:"item_type" bson:"item_type"`
	MaterialType  int                                  `json:"materialtype" bson:"materialtype"`
	Names         []LanguageModel                      `json:"names" bson:"names"`
	GroupCode     string                               `json:"group_code" bson:"group_code"`
	GroupNames    []LanguageModel                      `json:"group_names" bson:"group_names"`
	ItemUnitCode  string                               `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames []LanguageModel                      `json:"itemunitnames" bson:"itemunitnames"`
	Prices        []ProcessMongoBarcodePriceModel      `json:"prices" bson:"prices"`
	RefBarCodes   []ProcessMongoBarcodeRefBarcodeModel `json:"refbarcodes" bson:"refbarcodes"`
	ImageUri      string                               `json:"imageuri" bson:"imageuri"`
}

type ProcessMongoBarcodeRefBarcodeModel struct {
	Barcode      string  `json:"barcode" bson:"barcode"`
	ItemUnitCode string  `json:"item_unit_code" bson:"item_unit_code"`
	UnitStand    float64 `json:"standvalue" bson:"standvalue"`
	UnitDivide   float64 `json:"dividevalue" bson:"dividevalue"`
}

type ProcessMongoBarcodePriceModel struct {
	KeyNumber int     `json:"key_number" bson:"key_number"`
	Price     float64 `json:"price" bson:"price"`
}
