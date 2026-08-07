package models

type ProcessMongoBarcodeModel struct {
	HoldingCode   string                               `json:"holdingcode" bson:"holdingcode"`
	BusinessCode  string                               `json:"businesscode" bson:"businesscode"`
	ItemCode      string                               `json:"itemcode" bson:"itemcode"`
	Barcode       string                               `json:"barcode" bson:"barcode"`
	ItemType      int                                  `json:"itemtype" bson:"itemtype"`
	MaterialType  int                                  `json:"materialtype" bson:"materialtype"`
	Names         []LanguageModel                      `json:"names" bson:"names"`
	GroupCode     string                               `json:"groupcode" bson:"groupcode"`
	GroupNames    []LanguageModel                      `json:"groupnames" bson:"groupnames"`
	ItemUnitCode  string                               `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames []LanguageModel                      `json:"itemunitnames" bson:"itemunitnames"`
	Prices        []ProcessMongoBarcodePriceModel      `json:"prices" bson:"prices"`
	RefBarCodes   []ProcessMongoBarcodeRefBarcodeModel `json:"refbarcodes" bson:"refbarcodes"`
	ImageUri      string                               `json:"imageuri" bson:"imageuri"`
}

type ProcessMongoBarcodeRefBarcodeModel struct {
	Barcode      string  `json:"barcode" bson:"barcode"`
	ItemUnitCode string  `json:"itemunitcode" bson:"itemunitcode"`
	UnitStand    float64 `json:"standvalue" bson:"standvalue"`
	UnitDivide   float64 `json:"dividevalue" bson:"dividevalue"`
}

type ProcessMongoBarcodePriceModel struct {
	KeyNumber int     `json:"keynumber" bson:"keynumber"`
	Price     float64 `json:"price" bson:"price"`
}
