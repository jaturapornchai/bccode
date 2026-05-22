package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeCollectionName = "productBarcodes"

type ProductBarcodeBase struct {
	ItemCode string          `json:"itemcode" bson:"itemcode"`
	Barcode string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	GroupGuid string          `json:"groupguid" bson:"groupguid"`
	GroupCode string          `json:"group_code" bson:"group_code"`
	GroupNames *[]models.NameX `json:"group_names" bson:"group_names"`
	GroupsuboneGuid string          `json:"groupsuboneguid" bson:"groupsuboneguid"`
	GroupsuboneCode string          `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupsuboneNames *[]models.NameX `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupsubtwoGuid string          `json:"groupsubtwoguid" bson:"groupsubtwoguid"`
	GroupsubtwoCode string          `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupsubtwoNames *[]models.NameX `json:"groupsubtwonames" bson:"groupsubtwonames"`
	BrandGuid string          `json:"brandguid" bson:"brandguid"`
	BrandCode string          `json:"brand_code" bson:"brand_code"`
	BrandNames *[]models.NameX `json:"brandnames" bson:"brandnames"`
	DesignGuid string          `json:"designguid" bson:"designguid"`
	DesignCode string          `json:"designcode" bson:"designcode"`
	DesignNames *[]models.NameX `json:"designnames" bson:"designnames"`
	ModelGuid string          `json:"modelguid" bson:"modelguid"`
	ModelCode string          `json:"modelcode" bson:"modelcode"`
	ModelNames *[]models.NameX `json:"modelnames" bson:"modelnames"`
	PatternGuid string          `json:"patternguid" bson:"patternguid"`
	PatternCode string          `json:"patterncode" bson:"patterncode"`
	PatternNames *[]models.NameX `json:"patternnames" bson:"patternnames"`
	GradeGuid string          `json:"gradeguid" bson:"gradeguid"`
	GradeCode string          `json:"gradecode" bson:"gradecode"`
	GradeNames *[]models.NameX `json:"gradenames" bson:"gradenames"`
	CategoryGuid string          `json:"category_guid" bson:"category_guid"`
	CategoryCode string          `json:"categorycode" bson:"categorycode"`
	CategoryNames *[]models.NameX `json:"category_names" bson:"category_names"`
	ClassGuid string          `json:"classguid" bson:"classguid"`
	ClassCode string          `json:"classcode" bson:"classcode"`
	ClassNames *[]models.NameX `json:"classnames" bson:"classnames"`
	OrderPoint float64         `json:"orderpoint" bson:"orderpoint"`
	MinPoint float64         `json:"minpoint" bson:"minpoint"`
	MaxPoint float64         `json:"maxpoint" bson:"maxpoint"`
	RefGuidFixed string          `json:"refguidfixed" bson:"refguidfixed"`

	Names *[]models.NameX  `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	XSorts *[]models.XSort  `json:"xsorts" bson:"xsorts" validate:"unique=Code,dive"`
	ItemGuid string           `json:"item_guid" bson:"item_guid"`
	ItemUnitGuid string           `json:"itemunitguid" bson:"itemunitguid"`
	ItemUnitCode string           `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames *[]models.NameX  `json:"itemunitnames" bson:"itemunitnames"`
	ItemUnitSize float64          `json:"itemunitsize" bson:"itemunitsize"`
	Prices *[]ProductPrice  `json:"prices" bson:"prices"`
	ImageURI string           `json:"imageuri" bson:"imageuri"`
	Options *[]ProductOption `json:"options" bson:"options"`
	Images *[]ProductImage  `json:"images" bson:"images"`
	UseImageOrColor bool             `json:"useimageorcolor" bson:"useimageorcolor"`
	ColorSelect string           `json:"colorselect" bson:"colorselect"`
	ColorSelectHex string           `json:"colorselecthex" bson:"colorselecthex"`

	Condition bool    `json:"condition" bson:"condition"`
	DivideValue float64 `json:"dividevalue" bson:"dividevalue"`
	StandValue float64 `json:"standvalue" bson:"standvalue"`
	IsUseSubBarcodes bool    `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	IsMainBarcode bool    `json:"is_main_barcode" bson:"is_main_barcode"`

	ItemType int8   `json:"item_type" bson:"item_type"`
	MaterialType int8   `json:"materialtype" bson:"materialtype"`
	TaxType int8   `json:"tax_type" bson:"tax_type"`
	VatType int8   `json:"vat_type" bson:"vat_type"`
	IsSumPoint bool   `json:"issumpoint" bson:"issumpoint"`
	MaxDiscount string `json:"maxdiscount" bson:"maxdiscount"`
	IsDividend bool   `json:"isdividend" bson:"isdividend"`

	FixedCost *[]FixedCost          `json:"fixedcost" bson:"fixedcost"`
	RefUnitNames *[]models.NameX       `json:"refunitnames" bson:"refunitnames"`
	StockBarcode string                `json:"stockbarcode" bson:"stockbarcode"`
	Qty float64               `json:"qty" bson:"qty"`
	RefDivideValue float64               `json:"refdividevalue" bson:"refdividevalue"`
	RefStandValue float64               `json:"refstandvalue" bson:"refstandvalue"`
	VatCal int                   `json:"vatcal" bson:"vatcal"`
	IsALaCarte bool                  `json:"isalacarte" bson:"isalacarte"`
	OrderTypes *[]ProductOrderType   `json:"ordertypes" bson:"ordertypes"`
	ProductType ProductType           `json:"producttype" bson:"producttype"`
	IsSplitUnitPrint bool                  `json:"issplitunitprint" bson:"issplitunitprint"`
	IsOnlyStaff bool                  `json:"isonlystaff" bson:"isonlystaff"`
	FoodType int                   `json:"foodtype" bson:"foodtype"`
	Discount string                `json:"discount" bson:"discount"`
	IsStockForRestaurant bool                  `json:"isstockforrestaurant" bson:"isstockforrestaurant"`
	ManufacturerGUID string                `json:"manufacturerguid" bson:"manufacturerguid"`
	ManufacturerCode string                `json:"manufacturercode" bson:"manufacturercode"`
	ManufacturerNames *[]models.NameX       `json:"manufacturernames" bson:"manufacturernames"`
	Dimensions []ProductDimension    `json:"dimensions" bson:"dimensions"`
	IsDiscountPointOfPurchase bool                  `json:"isdiscountpointofpurchase" bson:"isdiscountpointofpurchase"`
	Restaurant ProductRestaurant     `json:"restaurant" bson:"restaurant"`
	IsAlert bool                  `json:"isalert" bson:"isalert"`
	AlertDescription string                `json:"alertdescription" bson:"alertdescription" validate:"max=1500"`
	Description string                `json:"description" bson:"description" validate:"max=1500"`
	TimeForSales *[]ProductTimeForSale `json:"timeforsales" bson:"timeforsales"`
}

type ProductTimeForSale struct {
	DaysOfWeek []int8 `json:"daysofweek" bson:"daysofweek"`
	FromDate string `json:"fromdate" bson:"fromdate"`
	ToDate string `json:"todate" bson:"todate"`
	FromTime string `json:"fromtime" bson:"fromtime"`
	ToTime string `json:"totime" bson:"totime"`
}

type FixedCost struct {
	EffectDate string  `json:"effectdate" bson:"effectdate"`
	Amount float64 `json:"amount" bson:"amount"`
}

type ProductRestaurant struct {
	IsForRestaurant bool `json:"isforrestaurant" bson:"isforrestaurant"`             // ทานที่ร้าน
	IsForTakeAway bool `json:"isfortakeaway" bson:"isfortakeaway"`                 // สั่งกลับบ้าน
	IsForDelivery bool `json:"isfordelivery" bson:"isfordelivery"`                 // เดลิเวอรี่
	IsForCustomer bool `json:"isforcustomer" bson:"isforcustomer"`                 // สำหรับลูกค้าสามารถสั่งได้
	IsForCustomerPreOrder bool `json:"isforcustomerpreorder" bson:"isforcustomerpreorder"` // สำหรับลูกค้าสามารถสั่ง preorder
}

type ProductDimension struct {
	models.DocIdentity `bson:"inline"`
	Names *[]models.NameX      `json:"names" bson:"names"`
	IsDisabled bool                 `json:"isdisabled" bson:"isdisabled"`
	Item ProductDimensionItem `json:"item" bson:"item"`
}

type ProductDimensionItem struct {
	models.DocIdentity `bson:"inline"`
	Names *[]models.NameX `json:"names" bson:"names"`
	IsDisabled bool            `json:"isdisabled" bson:"isdisabled"`
}

type RefProductBarcode struct {
	GuidFixed string          `json:"guid_fixed" bson:"guid_fixed"`
	Names *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode string          `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition bool            `json:"condition" bson:"condition"`
	DivideValue float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue float64         `json:"standvalue" bson:"standvalue"`
	Qty float64         `json:"qty" bson:"qty"`
}

type ProductBarcode struct {
	models.PartitionIdentity `bson:"inline"`
	ProductBarcodeBase  `bson:"inline"`
	RefBarcodes *[]RefProductBarcode          `json:"refbarcodes" bson:"refbarcodes"`
	BOM *[]BOMProductBarcode          `json:"bom" bson:"bom"`
	BusinessTypes *[]ProductBarcodeBusinessType `json:"businesstypes" bson:"businesstypes" `
	IgnoreBranches *[]ProductBarcodeBranch       `json:"ignorebranches" bson:"ignorebranches"`
}

type ProductBarcodeBusinessType struct {
	models.DocIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names"`
	IsIgnore bool            `json:"isignore" bson:"isignore"`
}

type ProductBarcodeBranch struct {
	models.DocIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names"`
	IsIgnore bool            `json:"isignore" bson:"isignore"`
}

type ProductImage struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI string `json:"uri" bson:"uri"`
}

type ProductPrice struct {
	KeyNumber int     `json:"key_number" bson:"key_number"`
	Price float64 `json:"price" bson:"price"`
}

type ProductBarcodeInfo struct {
	models.DocIdentity  `bson:"inline"`
	models.ShopIdentity `bson:"inline"`
	ProductBarcode  `bson:"inline"`
}

func (ProductBarcodeInfo) CollectionName() string {
	return productBarcodeCollectionName
}

type ProductBarcodeData struct {
	models.ShopIdentity `bson:"inline"`
	ProductBarcodeInfo  `bson:"inline"`
}

type ProductBarcodeDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductBarcodeData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductBarcodeDoc) CollectionName() string {
	return productBarcodeCollectionName
}

func (doc ProductBarcodeDoc) ToRefBarcode() RefProductBarcode {
	return RefProductBarcode{
		GuidFixed:     doc.GuidFixed,
		Names:         doc.Names,
		ItemUnitCode:  doc.ItemUnitCode,
		ItemUnitNames: doc.ItemUnitNames,
		Barcode:       doc.Barcode,
	}
}

func (doc ProductBarcodeDoc) ToBOM() BOMProductBarcode {

	return BOMProductBarcode{
		BarcodeGuidFixed: doc.GuidFixed,
		Names:            doc.Names,
		ItemUnitCode:     doc.ItemUnitCode,
		ItemUnitNames:    doc.ItemUnitNames,
		Barcode:          doc.Barcode,
	}
}

type ProductBarcodeItemGuid struct {
	Barcode string `json:"barcode" bson:"barcode"`
}

func (ProductBarcodeItemGuid) CollectionName() string {
	return productBarcodeCollectionName
}

type ProductBarcodeActivity struct {
	ProductBarcodeData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductBarcodeActivity) CollectionName() string {
	return productBarcodeCollectionName
}

type ProductBarcodeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductBarcodeDeleteActivity) CollectionName() string {
	return productBarcodeCollectionName
}

type ProductBarcodeSearch struct {
	ICCode string   `json:"iccode" ch:"iccode"`
	Barcode string   `json:"barcode" ch:"barcode"`
	UnitCode string   `json:"unitcode" ch:"unitcode"`
	Price string   `json:"price" ch:"price"`
	Names []string `json:"names" ch:"names"`
}

func (ProductBarcodeSearch) TableName() string {
	return productBarcodeCollectionName
}

// //
// Names                    datatypes.JSON `json:"names"  gorm:"column:names;type:jsonb;default:'[]'" `
// Names                    *JSONB  `json:"names"  gorm:"column:names;type:jsonb" `
type ProductBarcodePg struct {
	ShopID string `json:"shopid" gorm:"column:shopid;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	Barcode string              `json:"barcode" gorm:"column:barcode;primaryKey"`
	Names JSONB               `json:"names"  gorm:"column:names;type:jsonb" `
	UnitCode string              `json:"item_unit_code" gorm:"column:unitcode"`
	UnitNames JSONB               `json:"itemunitnames" gorm:"column:unitnames;type:jsonb"`
	BalanceQty float64             `json:"balance_qty" gorm:"column:balance_qty"`
	MainBarcodeRef string              `json:"mainbarcoderef" gorm:"column:mainbarcoderef"`
	StandValue float64             `json:"standvalue" gorm:"column:standvalue"`
	DivideValue float64             `json:"dividevalue" gorm:"column:dividevalue"`
	BalanceAmount float64             `json:"balanceamount" gorm:"column:balanceamount"`
	AverageCost float64             `json:"averagecost" gorm:"column:averagecost"`
	ItemCode string              `json:"itemcode" gorm:"column:itemcode"`
	BrandCode string              `json:"brand_code" gorm:"column:brand_code"`
	BrandName JSONB               `json:"brandnames" gorm:"column:brandnames;type:jsonb"`
	CategoryCode string              `json:"categorycode" gorm:"column:categorycode"`
	CategoryName JSONB               `json:"category_names" gorm:"column:category_names;type:jsonb"`
	ClassCode string              `json:"classcode" gorm:"column:classcode"`
	ClassNames JSONB               `json:"classnames" gorm:"column:classnames;type:jsonb"`
	DesignCode string              `json:"designcode" gorm:"column:designcode"`
	DesignNames JSONB               `json:"designnames" gorm:"column:designnames;type:jsonb"`
	GradeCode string              `json:"gradecode" gorm:"column:gradecode"`
	GradeNames JSONB               `json:"gradenames" gorm:"column:gradenames;type:jsonb"`
	GroupCode string              `json:"group_code" gorm:"column:group_code"`
	GroupNames JSONB               `json:"group_names" gorm:"column:group_names;type:jsonb"`
	GroupSubOneCode string              `json:"groupsubonecode" gorm:"column:groupsubonecode"`
	GroupSubOneNames JSONB               `json:"groupsubonenames" gorm:"column:groupsubonenames;type:jsonb"`
	GroupSubTwoCode string              `json:"groupsubtwocode" gorm:"column:groupsubtwocode"`
	GroupSubTwoNames JSONB               `json:"groupsubtwonames" gorm:"column:groupsubtwonames;type:jsonb"`
	ModelCode string              `json:"modelcode" gorm:"column:modelcode"`
	ModelNames JSONB               `json:"modelnames" gorm:"column:modelnames;type:jsonb"`
	PatternCode string              `json:"patterncode" gorm:"column:patterncode"`
	PatternNames JSONB               `json:"patternnames" gorm:"column:patternnames;type:jsonb"`
	BOM BOMProductBarcodePg `json:"bom" gorm:"column:bom;type:jsonb"`
}

func (ProductBarcodePg) TableName() string {
	return "productbarcode"
}

type BOMProductBarcodePg []BOMProductBarcode

func (a BOMProductBarcodePg) Value() (driver.Value, error) {

	j, err := json.Marshal(a)
	return j, err
}

func (a *BOMProductBarcodePg) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}

type JSONB []models.NameX

// Value Marshal
func (a JSONB) Value() (driver.Value, error) {

	j, err := json.Marshal(a)
	return j, err
	//return json.Marshal(a)
}

// Scan Unmarshal
func (a *JSONB) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
