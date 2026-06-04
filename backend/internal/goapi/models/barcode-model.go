package models

type BarcodeModel struct {
	HoldingCode          string
	ItemCode             string
	Barcode              string
	BarcodeRef           string
	Name0                string
	Name1                string
	Name2                string
	Name3                string
	Name4                string
	Name5                string
	UnitCode             string
	UnitName             string
	GroupCode            string
	GroupNames           string
	Price1               float64
	PriceRetail          float64
	UnitStand            float64
	UnitDivide           float64
	BarcodeRefUnitStand  float64
	BarcodeRefUnitDivide float64
	IsStock              int
	ItemType             int
	MaterialType         int
	Checksum             string
	ImageUri             string
}

func NewBarcodeModel() BarcodeModel {
	return BarcodeModel{
		HoldingCode:          "",
		ItemCode:             "",
		Barcode:              "",
		BarcodeRef:           "",
		Name0:                "",
		Name1:                "",
		Name2:                "",
		Name3:                "",
		Name4:                "",
		Name5:                "",
		UnitCode:             "",
		UnitName:             "",
		GroupCode:            "",
		GroupNames:           "",
		Price1:               0,
		PriceRetail:          0,
		UnitStand:            0,
		UnitDivide:           0,
		BarcodeRefUnitStand:  0,
		BarcodeRefUnitDivide: 0,
		IsStock:              0,
		ItemType:             0,
		MaterialType:         0,
		Checksum:             "",
		ImageUri:             "",
	}
}

type BarcodeRefModel struct {
	HoldingCode string
	Barcode     string
	BarcodeRef  string
	ItemCode    string
	UnitCode    string
	StandValue  float64
	DivideValue float64
}
