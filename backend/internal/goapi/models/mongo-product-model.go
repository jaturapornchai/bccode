package models

type MongoProductBarcodeModel struct {
	ShopId string                               `json:"shopid" bson:"shopid"`
	GuidFixed string                               `json:"guid_fixed" bson:"guid_fixed"`
	ItemCode string                               `json:"itemcode" bson:"itemcode"`
	Barcode string                               `json:"barcode" bson:"barcode"`
	Names []LanguageModel                      `json:"names" bson:"names"`
	GroupCode string                               `json:"group_code" bson:"group_code"`
	GroupNames []LanguageModel                      `json:"group_names" bson:"group_names"`
	ItemUnitCode string                               `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames []LanguageModel                      `json:"itemunitnames" bson:"itemunitnames"`
	Prices []PriceModel                         `json:"prices" bson:"prices"`
	DivideValue float64                              `json:"dividevalue" bson:"dividevalue"`
	StandValue float64                              `json:"standvalue" bson:"standvalue"`
	RefBarCodes []ProcessMongoBarcodeRefBarcodeModel `json:"refbarcodes" bson:"refbarcodes"`
	ItemType int                                  `json:"item_type" bson:"item_type"`
	IsUseSubBarcodes bool                                 `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	ImageUri string                               `json:"imageuri" bson:"imageuri"`
	BrandCode string                               `json:"brand_code" bson:"brand_code"`
	BrandNames []LanguageModel                      `json:"brandnames" bson:"brandnames"`
	CategoryCode string                               `json:"categorycode" bson:"categorycode"`
	CategoryNames []LanguageModel                      `json:"category_names" bson:"category_names"`
	ClassCode string                               `json:"classcode" bson:"classcode"`
	ClassNames []LanguageModel                      `json:"classnames" bson:"classnames"`
	DesignCode string                               `json:"designcode" bson:"designcode"`
	DesignNames []LanguageModel                      `json:"designnames" bson:"designnames"`
	GradeCode string                               `json:"gradecode" bson:"gradecode"`
	GradeNames []LanguageModel                      `json:"gradenames" bson:"gradenames"`
	ModelCode string                               `json:"modelcode" bson:"modelcode"`
	ModelNames []LanguageModel                      `json:"modelnames" bson:"modelnames"`
	PatternCode string                               `json:"patterncode" bson:"patterncode"`
	PatternNames []LanguageModel                      `json:"patternnames" bson:"patternnames"`
	GroupSubOneCode string                               `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupSubOneNames []LanguageModel                      `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupSubTwoCode string                               `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupSubTwoNames []LanguageModel                      `json:"groupsubtwonames" bson:"groupsubtwonames"`
}

type PriceModel struct {
	KeyNumber int     `json:"key_number" bson:"key_number"`
	Price float64 `json:"price" bson:"price"`
}
