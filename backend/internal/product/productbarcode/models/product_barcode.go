package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeCollectionName = "productbarcodes"

type ProductBarcodeBase struct {
	ItemCode         string          `json:"itemcode" bson:"itemcode"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	GroupGuid        string          `json:"groupguid" bson:"groupguid"`
	GroupCode        string          `json:"groupcode" bson:"groupcode"`
	GroupNames       *[]models.NameX `json:"groupnames" bson:"groupnames"`
	GroupsuboneGuid  string          `json:"groupsuboneguid" bson:"groupsuboneguid"`
	GroupsuboneCode  string          `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupsuboneNames *[]models.NameX `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupsubtwoGuid  string          `json:"groupsubtwoguid" bson:"groupsubtwoguid"`
	GroupsubtwoCode  string          `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupsubtwoNames *[]models.NameX `json:"groupsubtwonames" bson:"groupsubtwonames"`
	BrandGuid        string          `json:"brandguid" bson:"brandguid"`
	BrandCode        string          `json:"brandcode" bson:"brandcode"`
	BrandNames       *[]models.NameX `json:"brandnames" bson:"brandnames"`
	DesignGuid       string          `json:"designguid" bson:"designguid"`
	DesignCode       string          `json:"designcode" bson:"designcode"`
	DesignNames      *[]models.NameX `json:"designnames" bson:"designnames"`
	ModelGuid        string          `json:"modelguid" bson:"modelguid"`
	ModelCode        string          `json:"modelcode" bson:"modelcode"`
	ModelNames       *[]models.NameX `json:"modelnames" bson:"modelnames"`
	PatternGuid      string          `json:"patternguid" bson:"patternguid"`
	PatternCode      string          `json:"patterncode" bson:"patterncode"`
	PatternNames     *[]models.NameX `json:"patternnames" bson:"patternnames"`
	GradeGuid        string          `json:"gradeguid" bson:"gradeguid"`
	GradeCode        string          `json:"gradecode" bson:"gradecode"`
	GradeNames       *[]models.NameX `json:"gradenames" bson:"gradenames"`
	CategoryGuid     string          `json:"categoryguid" bson:"categoryguid"`
	CategoryCode     string          `json:"categorycode" bson:"categorycode"`
	CategoryNames    *[]models.NameX `json:"categorynames" bson:"categorynames"`
	ClassGuid        string          `json:"classguid" bson:"classguid"`
	ClassCode        string          `json:"classcode" bson:"classcode"`
	ClassNames       *[]models.NameX `json:"classnames" bson:"classnames"`
	OrderPoint       float64         `json:"orderpoint" bson:"orderpoint"`
	MinPoint         float64         `json:"minpoint" bson:"minpoint"`
	MaxPoint         float64         `json:"maxpoint" bson:"maxpoint"`
	RefGuidFixed     string          `json:"refguidfixed" bson:"refguidfixed"`

	Names           *[]models.NameX  `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	XSorts          *[]models.XSort  `json:"xsorts" bson:"xsorts" validate:"unique=Code,dive"`
	ItemGuid        string           `json:"itemguid" bson:"itemguid"`
	ItemUnitGuid    string           `json:"itemunitguid" bson:"itemunitguid"`
	ItemUnitCode    string           `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames   *[]models.NameX  `json:"itemunitnames" bson:"itemunitnames"`
	ItemUnitSize    float64          `json:"itemunitsize" bson:"itemunitsize"`
	Prices          *[]ProductPrice  `json:"prices" bson:"prices"`
	ImageURI        string           `json:"imageuri" bson:"imageuri"`
	Options         *[]ProductOption `json:"options" bson:"options"`
	Images          *[]ProductImage  `json:"images" bson:"images"`
	Videos          *[]ProductVideo  `json:"videos" bson:"videos"`
	UseImageOrColor bool             `json:"useimageorcolor" bson:"useimageorcolor"`
	ColorSelect     string           `json:"colorselect" bson:"colorselect"`
	ColorSelectHex  string           `json:"colorselecthex" bson:"colorselecthex"`

	Condition        bool    `json:"condition" bson:"condition"`
	DivideValue      float64 `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64 `json:"standvalue" bson:"standvalue"`
	IsUseSubBarcodes bool    `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	IsMainBarcode    bool    `json:"ismainbarcode" bson:"ismainbarcode"`

	// Marketplace & Logistics
	PackageWeight       float64                  `json:"packageweight" bson:"packageweight"`
	PackageLength       float64                  `json:"packagelength" bson:"packagelength"`
	PackageWidth        float64                  `json:"packagewidth" bson:"packagewidth"`
	PackageHeight       float64                  `json:"packageheight" bson:"packageheight"`
	MarketplaceProducts *[]MarketplaceProductMap `json:"marketplaceproducts" bson:"marketplaceproducts"`

	ItemType     int8   `json:"itemtype" bson:"itemtype"`
	MaterialType int8   `json:"materialtype" bson:"materialtype"`
	TaxType      int8   `json:"taxtype" bson:"taxtype"`
	VatType      int8   `json:"vattype" bson:"vattype"`
	IsSumPoint   bool   `json:"issumpoint" bson:"issumpoint"`
	MaxDiscount  string `json:"maxdiscount" bson:"maxdiscount"`
	IsDividend   bool   `json:"isdividend" bson:"isdividend"`

	FixedCost                 *[]FixedCost                  `json:"fixedcost" bson:"fixedcost"`
	RefUnitNames              *[]models.NameX               `json:"refunitnames" bson:"refunitnames"`
	StockBarcode              string                        `json:"stockbarcode" bson:"stockbarcode"`
	Qty                       float64                       `json:"qty" bson:"qty"`
	RefDivideValue            float64                       `json:"refdividevalue" bson:"refdividevalue"`
	RefStandValue             float64                       `json:"refstandvalue" bson:"refstandvalue"`
	VatCal                    int                           `json:"vatcal" bson:"vatcal"`
	IsALaCarte                bool                          `json:"isalacarte" bson:"isalacarte"`
	OrderTypes                *[]ProductOrderType           `json:"ordertypes" bson:"ordertypes"`
	ProductType               ProductType                   `json:"producttype" bson:"producttype"`
	IsSplitUnitPrint          bool                          `json:"issplitunitprint" bson:"issplitunitprint"`
	IsOnlyStaff               bool                          `json:"isonlystaff" bson:"isonlystaff"`
	FoodType                  int                           `json:"foodtype" bson:"foodtype"`
	Discount                  string                        `json:"discount" bson:"discount"`
	IsStockForRestaurant      bool                          `json:"isstockforrestaurant" bson:"isstockforrestaurant"`
	ManufacturerGUID          string                        `json:"manufacturerguid" bson:"manufacturerguid"`
	ManufacturerCode          string                        `json:"manufacturercode" bson:"manufacturercode"`
	ManufacturerNames         *[]models.NameX               `json:"manufacturernames" bson:"manufacturernames"`
	Manufacturers             *[]ProductBarcodeManufacturer `json:"manufacturers" bson:"manufacturers"`
	Suppliers                 *[]ProductBarcodeSupplier     `json:"suppliers" bson:"suppliers"`
	Dimensions                []ProductDimension            `json:"dimensions" bson:"dimensions"`
	IsDiscountPointOfPurchase bool                          `json:"isdiscountpointofpurchase" bson:"isdiscountpointofpurchase"`
	Restaurant                ProductRestaurant             `json:"restaurant" bson:"restaurant"`
	IsAlert                   bool                          `json:"isalert" bson:"isalert"`
	AlertDescription          string                        `json:"alertdescription" bson:"alertdescription" validate:"max=1500"`
	Description               string                        `json:"description" bson:"description" validate:"max=1500"`
	TimeForSales              *[]ProductTimeForSale         `json:"timeforsales" bson:"timeforsales"`
}

type ProductTimeForSale struct {
	DaysOfWeek []int8 `json:"daysofweek" bson:"daysofweek"`
	FromDate   string `json:"fromdate" bson:"fromdate"`
	ToDate     string `json:"todate" bson:"todate"`
	FromTime   string `json:"fromtime" bson:"fromtime"`
	ToTime     string `json:"totime" bson:"totime"`
}

type FixedCost struct {
	EffectDate string  `json:"effectdate" bson:"effectdate"`
	Amount     float64 `json:"amount" bson:"amount"`
}

type ProductRestaurant struct {
	IsForRestaurant       bool `json:"isforrestaurant" bson:"isforrestaurant"`             // ทานที่ร้าน
	IsForTakeAway         bool `json:"isfortakeaway" bson:"isfortakeaway"`                 // สั่งกลับบ้าน
	IsForDelivery         bool `json:"isfordelivery" bson:"isfordelivery"`                 // เดลิเวอรี่
	IsForCustomer         bool `json:"isforcustomer" bson:"isforcustomer"`                 // สำหรับลูกค้าสามารถสั่งได้
	IsForCustomerPreOrder bool `json:"isforcustomerpreorder" bson:"isforcustomerpreorder"` // สำหรับลูกค้าสามารถสั่ง preorder
}

type ProductDimension struct {
	models.DocIdentity `bson:"inline"`
	Names              *[]models.NameX      `json:"names" bson:"names"`
	IsDisabled         bool                 `json:"isdisabled" bson:"isdisabled"`
	Item               ProductDimensionItem `json:"item" bson:"item"`
}

type ProductDimensionItem struct {
	models.DocIdentity `bson:"inline"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	IsDisabled         bool            `json:"isdisabled" bson:"isdisabled"`
}

type RefProductBarcode struct {
	GuidFixed     string          `json:"guidfixed" bson:"guidfixed"`
	ItemCode      string          `json:"itemcode" bson:"itemcode"`
	Names         *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode  string          `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode       string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition     bool            `json:"condition" bson:"condition"`
	DivideValue   float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue    float64         `json:"standvalue" bson:"standvalue"`
	Qty           float64         `json:"qty" bson:"qty"`

	// Marketplace & SKU Logistics
	SellerSKU              string               `json:"sellersku" bson:"sellersku"`
	SkuPackageWeight       float64              `json:"skupackageweight" bson:"skupackageweight"`
	SkuPackageLength       float64              `json:"skupackagelength" bson:"skupackagelength"`
	SkuPackageWidth        float64              `json:"skupackagewidth" bson:"skupackagewidth"`
	SkuPackageHeight       float64              `json:"skupackageheight" bson:"skupackageheight"`
	MarketplaceSKUMappings *[]MarketplaceSKUMap `json:"marketplaceskumappings" bson:"marketplaceskumappings"`
}

type ProductBarcodeBOMVersion struct {
	GuidFixed string               `json:"guidfixed" bson:"guidfixed"`
	StartDate time.Time            `json:"startdate" bson:"startdate"`
	EndDate   *time.Time           `json:"enddate" bson:"enddate"`
	BOM       *[]BOMProductBarcode `json:"bom" bson:"bom"`
}

type ProductBarcode struct {
	models.PartitionIdentity `bson:"inline"`
	ProductBarcodeBase       `bson:"inline"`
	RefBarcodes              *[]RefProductBarcode          `json:"refbarcodes" bson:"refbarcodes"`
	BOM                      *[]BOMProductBarcode          `json:"bom" bson:"bom"`
	BOMs                     *[]ProductBarcodeBOMVersion   `json:"boms" bson:"boms"`
	BusinessTypes            *[]ProductBarcodeBusinessType `json:"businesstypes" bson:"businesstypes" `
	IgnoreBranches           *[]ProductBarcodeBranch       `json:"ignorebranches" bson:"ignorebranches"`
}

type ProductBarcodeBusinessType struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	IsIgnore           bool            `json:"isignore" bson:"isignore"`
}

type ProductBarcodeBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	IsIgnore           bool            `json:"isignore" bson:"isignore"`
}

type ProductBarcodeManufacturer struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type ProductBarcodeSupplier struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type ProductImage struct {
	XOrder int    `json:"xorder" bson:"xorder"`
	URI    string `json:"uri" bson:"uri"`
}

type ProductVideo struct {
	XOrder    int    `json:"xorder" bson:"xorder"`
	URI       string `json:"uri" bson:"uri"`
	PosterURI string `json:"posteruri" bson:"posteruri"`
}

type MarketplaceMediaAsset struct {
	Kind           string `json:"kind" bson:"kind"` // main, gallery, sku, size_chart, video, description
	URI            string `json:"uri" bson:"uri"`
	ExternalID     string `json:"externalid" bson:"externalid"`
	ExternalURL    string `json:"externalurl" bson:"externalurl"`
	OptionCode     string `json:"optioncode" bson:"optioncode"`
	OptionValue    string `json:"optionvalue" bson:"optionvalue"`
	SortOrder      int    `json:"sortorder" bson:"sortorder"`
	AltText        string `json:"alttext" bson:"alttext"`
	UseCase        string `json:"usecase" bson:"usecase"`
	MimeType       string `json:"mimetype" bson:"mimetype"`
	Width          int    `json:"width" bson:"width"`
	Height         int    `json:"height" bson:"height"`
	LastImportedAt string `json:"lastimportedat" bson:"lastimportedat"`
}

type MarketplaceAttributeValue struct {
	ValueID       string `json:"valueid" bson:"valueid"`
	ValueCode     string `json:"valuecode" bson:"valuecode"`
	ValueText     string `json:"valuetext" bson:"valuetext"`
	DisplayText   string `json:"displaytext" bson:"displaytext"`
	UnitCode      string `json:"unitcode" bson:"unitcode"`
	SortOrder     int    `json:"sortorder" bson:"sortorder"`
	IsCustomValue bool   `json:"iscustomvalue" bson:"iscustomvalue"`
}

type MarketplaceAttribute struct {
	AttributeID   string                       `json:"attributeid" bson:"attributeid"`
	AttributeCode string                       `json:"attributecode" bson:"attributecode"`
	AttributeName string                       `json:"attributename" bson:"attributename"`
	InputType     string                       `json:"inputtype" bson:"inputtype"` // text, single_select, multiselect, number, date, boolean
	Scope         string                       `json:"scope" bson:"scope"`         // product, sku, package, compliance
	IsRequired    bool                         `json:"isrequired" bson:"isrequired"`
	IsSaleProp    bool                         `json:"issaleprop" bson:"issaleprop"`
	IsCustom      bool                         `json:"iscustom" bson:"iscustom"`
	Values        *[]MarketplaceAttributeValue `json:"values" bson:"values"`
}

type MarketplaceSpecificationGroup struct {
	GroupCode  string                  `json:"groupcode" bson:"groupcode"`
	GroupName  string                  `json:"groupname" bson:"groupname"`
	Attributes *[]MarketplaceAttribute `json:"attributes" bson:"attributes"`
}

type MarketplacePayloadExample struct {
	Direction string `json:"direction" bson:"direction"` // import, export, stock_update, price_update, order_import
	UseCase   string `json:"usecase" bson:"usecase"`
	Payload   any    `json:"payload" bson:"payload"`
}

type ProductPrice struct {
	KeyNumber int     `json:"keynumber" bson:"keynumber"`
	Price     float64 `json:"price" bson:"price"`
}

type ProductBarcodeInfo struct {
	models.DocIdentity       `bson:"inline"`
	models.HoldingCodeentity `bson:"inline"`
	ProductBarcode           `bson:"inline"`
}

func (ProductBarcodeInfo) CollectionName() string {
	return productBarcodeCollectionName
}

type ProductBarcodeData struct {
	models.HoldingCodeentity `bson:"inline"`
	BusinessCode             string `json:"businesscode" bson:"businesscode"`
	ProductBarcodeInfo       `bson:"inline"`
}

type ProductBarcodeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductBarcodeData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductBarcodeDoc) CollectionName() string {
	return productBarcodeCollectionName
}

func (doc ProductBarcodeDoc) ToRefBarcode() RefProductBarcode {
	return RefProductBarcode{
		GuidFixed:     doc.GuidFixed,
		ItemCode:      doc.ItemCode,
		Names:         doc.Names,
		ItemUnitCode:  doc.ItemUnitCode,
		ItemUnitNames: doc.ItemUnitNames,
		Barcode:       doc.Barcode,
	}
}

func (doc ProductBarcodeDoc) ToBOM() BOMProductBarcode {

	return BOMProductBarcode{
		BarcodeGuidFixed: doc.GuidFixed,
		ItemCode:         doc.ItemCode,
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
	ICCode   string   `json:"iccode" ch:"iccode"`
	Barcode  string   `json:"barcode" ch:"barcode"`
	UnitCode string   `json:"unitcode" ch:"unitcode"`
	Price    string   `json:"price" ch:"price"`
	Names    []string `json:"names" ch:"names"`
}

func (ProductBarcodeSearch) TableName() string {
	return productBarcodeCollectionName
}

// //
// Names                    datatypes.JSON `json:"names"  gorm:"column:names;type:jsonb;default:'[]'" `
// Names                    *JSONB  `json:"names"  gorm:"column:names;type:jsonb" `
type ProductBarcodePg struct {
	HoldingCode              string `json:"holdingcode" gorm:"column:holding_code;primaryKey"`
	BusinessCode             string `json:"businesscode" gorm:"column:businesscode;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	Barcode                  string              `json:"barcode" gorm:"column:barcode;primaryKey"`
	Names                    JSONB               `json:"names"  gorm:"column:names;type:jsonb" `
	UnitCode                 string              `json:"itemunitcode" gorm:"column:unitcode"`
	UnitNames                JSONB               `json:"itemunitnames" gorm:"column:unitnames;type:jsonb"`
	BalanceQty               float64             `json:"balanceqty" gorm:"column:balanceqty"`
	MainBarcodeRef           string              `json:"mainbarcoderef" gorm:"column:mainbarcoderef"`
	StandValue               float64             `json:"standvalue" gorm:"column:standvalue"`
	DivideValue              float64             `json:"dividevalue" gorm:"column:dividevalue"`
	BalanceAmount            float64             `json:"balanceamount" gorm:"column:balanceamount"`
	AverageCost              float64             `json:"averagecost" gorm:"column:averagecost"`
	ItemCode                 string              `json:"itemcode" gorm:"column:itemcode"`
	ItemType                 int8                `json:"itemtype" gorm:"column:itemtype"`
	MaterialType             int8                `json:"materialtype" gorm:"column:materialtype"`
	BrandCode                string              `json:"brandcode" gorm:"column:brandcode"`
	BrandName                JSONB               `json:"brandnames" gorm:"column:brandnames;type:jsonb"`
	CategoryCode             string              `json:"categorycode" gorm:"column:categorycode"`
	CategoryName             JSONB               `json:"categorynames" gorm:"column:categorynames;type:jsonb"`
	ClassCode                string              `json:"classcode" gorm:"column:classcode"`
	ClassNames               JSONB               `json:"classnames" gorm:"column:classnames;type:jsonb"`
	DesignCode               string              `json:"designcode" gorm:"column:designcode"`
	DesignNames              JSONB               `json:"designnames" gorm:"column:designnames;type:jsonb"`
	GradeCode                string              `json:"gradecode" gorm:"column:gradecode"`
	GradeNames               JSONB               `json:"gradenames" gorm:"column:gradenames;type:jsonb"`
	GroupCode                string              `json:"groupcode" gorm:"column:groupcode"`
	GroupNames               JSONB               `json:"groupnames" gorm:"column:groupnames;type:jsonb"`
	GroupSubOneCode          string              `json:"groupsubonecode" gorm:"column:groupsubonecode"`
	GroupSubOneNames         JSONB               `json:"groupsubonenames" gorm:"column:groupsubonenames;type:jsonb"`
	GroupSubTwoCode          string              `json:"groupsubtwocode" gorm:"column:groupsubtwocode"`
	GroupSubTwoNames         JSONB               `json:"groupsubtwonames" gorm:"column:groupsubtwonames;type:jsonb"`
	ModelCode                string              `json:"modelcode" gorm:"column:modelcode"`
	ModelNames               JSONB               `json:"modelnames" gorm:"column:modelnames;type:jsonb"`
	PatternCode              string              `json:"patterncode" gorm:"column:patterncode"`
	PatternNames             JSONB               `json:"patternnames" gorm:"column:patternnames;type:jsonb"`
	BOM                      BOMProductBarcodePg `json:"bom" gorm:"column:bom;type:jsonb"`
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

// MarketplaceProductMap maps one internal product/barcode to one marketplace
// listing on a single platform (Shopee / Lazada / AliExpress / TikTok). One product can hold
// many entries — one per platform+account. Unified shape so all platforms
// share the same fields ("คล้ายๆกัน"); blank fields are simply unused per platform.
// Platform values: "shopee" | "lazada" | "aliexpress" | "tiktok".
type MarketplaceProductMap struct {
	// Platform identity
	Platform    string `json:"platform" bson:"platform"`
	AccountID   string `json:"accountid" bson:"accountid"`     // seller/shop account on the platform
	HoldingCode string `json:"holdingcode" bson:"holdingcode"` // legacy field, kept for backward-compat
	// Item identity on the marketplace
	MarketItemID  string `json:"marketitemid" bson:"marketitemid"`   // Shopee item_id / Lazada item_id / TikTok product_id
	MarketModelID string `json:"marketmodelid" bson:"marketmodelid"` // variation/model/sku id for this listing
	ItemURL       string `json:"itemurl" bson:"itemurl"`
	// SKU identity
	SellerSKU string `json:"sellersku" bson:"sellersku"` // our SKU pushed to the platform
	ShopSKU   string `json:"shopsku" bson:"shopsku"`     // platform-generated SKU (Lazada ShopSku)
	GTIN      string `json:"gtin" bson:"gtin"`           // EAN/UPC/GTIN (TikTok mandatory for some categories)
	// Catalog mapping (each platform has its own category/brand tree)
	CategoryID   string `json:"categoryid" bson:"categoryid"`
	CategoryName string `json:"categoryname" bson:"categoryname"`
	BrandID      string `json:"brandid" bson:"brandid"`
	// Media and specifications are internal import/sync structures. Normal product UI should display
	// merged product images/specs, not expose marketplace origin labels.
	MediaAssets         *[]MarketplaceMediaAsset         `json:"mediaassets" bson:"mediaassets"`
	SpecificationGroups *[]MarketplaceSpecificationGroup `json:"specificationgroups" bson:"specificationgroups"`
	RawAttributes       *[]MarketplaceAttribute          `json:"rawattributes" bson:"rawattributes"`
	PayloadExamples     *[]MarketplacePayloadExample     `json:"payloadexamples" bson:"payloadexamples"`
	// Price & stock (current value on platform + sync intent)
	Currency      string  `json:"currency" bson:"currency"`
	CustomPrice   float64 `json:"customprice" bson:"customprice"`     // price we intend to push
	PlatformPrice float64 `json:"platformprice" bson:"platformprice"` // actual price on platform (inbound)
	PlatformStock int     `json:"platformstock" bson:"platformstock"` // actual stock on platform (inbound)
	SyncStock     bool    `json:"syncstock" bson:"syncstock"`
	SyncPrice     bool    `json:"syncprice" bson:"syncprice"`
	// Listing & sync status
	Status        string `json:"status" bson:"status"` // LIVE / UNLIST / REVIEWING / REJECTED / DELETED
	RejectReason  string `json:"rejectreason" bson:"rejectreason"`
	DaysToShip    int    `json:"daystoship" bson:"daystoship"`
	IsPreOrder    bool   `json:"ispreorder" bson:"ispreorder"`
	SyncEnabled   bool   `json:"syncenabled" bson:"syncenabled"`
	SyncStatus    string `json:"syncstatus" bson:"syncstatus"` // legacy field, kept for backward-compat
	LastSyncAt    string `json:"lastsyncat" bson:"lastsyncat"`
	LastSyncError string `json:"lastsyncerror" bson:"lastsyncerror"`
}

// MarketplaceSKUMap maps a sub/variation barcode (RefProductBarcode) to a
// marketplace SKU. Same unified field set as MarketplaceProductMap minus the
// item-level catalog fields, so sub-barcode variations sync consistently.
type MarketplaceDimensionStock struct {
	DimensionKey      string  `json:"dimensionkey" bson:"dimensionkey"`
	DimensionName     string  `json:"dimensionname" bson:"dimensionname"`
	MarketDimensionID string  `json:"marketdimensionid" bson:"marketdimensionid"`
	AvailableQty      float64 `json:"availableqty" bson:"availableqty"`
	ReservedQty       float64 `json:"reservedqty" bson:"reservedqty"`
	InboundQty        float64 `json:"inboundqty" bson:"inboundqty"`
	OversellBufferQty float64 `json:"oversellbufferqty" bson:"oversellbufferqty"`
	LastPlatformStock float64 `json:"lastplatformstock" bson:"lastplatformstock"`
	LastSyncedAt      string  `json:"lastsyncedat" bson:"lastsyncedat"`
	LastSyncStatus    string  `json:"lastsyncstatus" bson:"lastsyncstatus"`
	LastSyncError     string  `json:"lastsyncerror" bson:"lastsyncerror"`
}

type MarketplaceSKUMap struct {
	Platform                   string                       `json:"platform" bson:"platform"`
	AccountID                  string                       `json:"accountid" bson:"accountid"`
	HoldingCode                string                       `json:"holdingcode" bson:"holdingcode"`
	MarketItemID               string                       `json:"marketitemid" bson:"marketitemid"`
	MarketModelID              string                       `json:"marketmodelid" bson:"marketmodelid"`
	SellerSKU                  string                       `json:"sellersku" bson:"sellersku"`
	ShopSKU                    string                       `json:"shopsku" bson:"shopsku"`
	GTIN                       string                       `json:"gtin" bson:"gtin"`
	MediaAssets                *[]MarketplaceMediaAsset     `json:"mediaassets" bson:"mediaassets"`
	RawAttributes              *[]MarketplaceAttribute      `json:"rawattributes" bson:"rawattributes"`
	Currency                   string                       `json:"currency" bson:"currency"`
	SyncStock                  bool                         `json:"syncstock" bson:"syncstock"`
	SyncPrice                  bool                         `json:"syncprice" bson:"syncprice"`
	CustomPrice                float64                      `json:"customprice" bson:"customprice"`
	PlatformPrice              float64                      `json:"platformprice" bson:"platformprice"`
	PlatformStock              int                          `json:"platformstock" bson:"platformstock"`
	MarketplaceDimensionStocks *[]MarketplaceDimensionStock `json:"marketplacedimensionstocks" bson:"marketplacedimensionstocks"`
	Status                     string                       `json:"status" bson:"status"`
	SyncEnabled                bool                         `json:"syncenabled" bson:"syncenabled"`
	LastSyncAt                 string                       `json:"lastsyncat" bson:"lastsyncat"`
}
