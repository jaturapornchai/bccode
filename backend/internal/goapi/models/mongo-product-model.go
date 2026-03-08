package models

type MongoProductBarcodeModel struct {
	ShopId           string                               `json:"shopid" bson:"shopid"`
	GuidFixed        string                               `json:"guidfixed" bson:"guidfixed"`
	ItemCode         string                               `json:"itemcode" bson:"itemcode"`
	Barcode          string                               `json:"barcode" bson:"barcode"`
	Names            []LanguageModel                      `json:"names" bson:"names"`
	GroupCode        string                               `json:"groupcode" bson:"groupcode"`
	GroupNames       []LanguageModel                      `json:"groupnames" bson:"groupnames"`
	ItemUnitCode     string                               `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames    []LanguageModel                      `json:"itemunitnames" bson:"itemunitnames"`
	Prices           []PriceModel                         `json:"prices" bson:"prices"`
	DivideValue      float64                              `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64                              `json:"standvalue" bson:"standvalue"`
	RefBarCodes      []ProcessMongoBarcodeRefBarcodeModel `json:"refbarcodes" bson:"refbarcodes"`
	ItemType         int                                  `json:"itemtype" bson:"itemtype"`
	IsUseSubBarcodes bool                                 `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	ImageUri         string                               `json:"imageuri" bson:"imageuri"`
	BrandCode        string                               `json:"brandcode" bson:"brandcode"`
	BrandNames       []LanguageModel                      `json:"brandnames" bson:"brandnames"`
	CategoryCode     string                               `json:"categorycode" bson:"categorycode"`
	CategoryNames    []LanguageModel                      `json:"categorynames" bson:"categorynames"`
	ClassCode        string                               `json:"classcode" bson:"classcode"`
	ClassNames       []LanguageModel                      `json:"classnames" bson:"classnames"`
	DesignCode       string                               `json:"designcode" bson:"designcode"`
	DesignNames      []LanguageModel                      `json:"designnames" bson:"designnames"`
	GradeCode        string                               `json:"gradecode" bson:"gradecode"`
	GradeNames       []LanguageModel                      `json:"gradenames" bson:"gradenames"`
	ModelCode        string                               `json:"modelcode" bson:"modelcode"`
	ModelNames       []LanguageModel                      `json:"modelnames" bson:"modelnames"`
	PatternCode      string                               `json:"patterncode" bson:"patterncode"`
	PatternNames     []LanguageModel                      `json:"patternnames" bson:"patternnames"`
	GroupSubOneCode  string                               `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupSubOneNames []LanguageModel                      `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupSubTwoCode  string                               `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupSubTwoNames []LanguageModel                      `json:"groupsubtwonames" bson:"groupsubtwonames"`
}

type PriceModel struct {
	KeyNumber int     `json:"keynumber" bson:"keynumber"`
	Price     float64 `json:"price" bson:"price"`
}
