package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeCollectionName = "productBarcodes"

type ProductBarcodeBase struct {
	ItemCode         string          `json:"itemcode" bson:"itemcode"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	GroupGuid        string          `json:"groupguid" bson:"groupguid"`
	GroupCode        string          `json:"group_code" bson:"group_code"`
	GroupNames       *[]models.NameX `json:"group_names" bson:"group_names"`
	GroupsuboneGuid  string          `json:"groupsuboneguid" bson:"groupsuboneguid"`
	GroupsuboneCode  string          `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupsuboneNames *[]models.NameX `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupsubtwoGuid  string          `json:"groupsubtwoguid" bson:"groupsubtwoguid"`
	GroupsubtwoCode  string          `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupsubtwoNames *[]models.NameX `json:"groupsubtwonames" bson:"groupsubtwonames"`
	BrandGuid        string          `json:"brandguid" bson:"brandguid"`
	BrandCode        string          `json:"brand_code" bson:"brand_code"`
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
	CategoryGuid     string          `json:"category_guid" bson:"category_guid"`
	CategoryCode     string          `json:"categorycode" bson:"categorycode"`
	CategoryNames    *[]models.NameX `json:"category_names" bson:"category_names"`
	ClassGuid        string          `json:"classguid" bson:"classguid"`
	ClassCode        string          `json:"classcode" bson:"classcode"`
	ClassNames       *[]models.NameX `json:"classnames" bson:"classnames"`
	OrderPoint       float64         `json:"orderpoint" bson:"orderpoint"`
	MinPoint         float64         `json:"minpoint" bson:"minpoint"`
	MaxPoint         float64         `json:"maxpoint" bson:"maxpoint"`
	RefGuidFixed     string          `json:"refguidfixed" bson:"refguidfixed"`

	Names           *[]models.NameX  `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	XSorts          *[]models.XSort  `json:"xsorts" bson:"xsorts" validate:"unique=Code,dive"`
	ItemGuid        string           `json:"item_guid" bson:"item_guid"`
	ItemUnitGuid    string           `json:"itemunitguid" bson:"itemunitguid"`
	ItemUnitCode    string           `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames   *[]models.NameX  `json:"itemunitnames" bson:"itemunitnames"`
	ItemUnitSize    float64          `json:"itemunitsize" bson:"itemunitsize"`
	Prices          *[]ProductPrice  `json:"prices" bson:"prices"`
	ImageURI        string           `json:"imageuri" bson:"imageuri"`
	Options         *[]ProductOption `json:"options" bson:"options"`
	Images          *[]ProductImage  `json:"images" bson:"images"`
	UseImageOrColor bool             `json:"useimageorcolor" bson:"useimageorcolor"`
	ColorSelect     string           `json:"colorselect" bson:"colorselect"`
	ColorSelectHex  string           `json:"colorselecthex" bson:"colorselecthex"`

	Condition        bool    `json:"condition" bson:"condition"`
	DivideValue      float64 `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64 `json:"standvalue" bson:"standvalue"`
	IsUseSubBarcodes bool    `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	IsMainBarcode    bool    `json:"is_main_barcode" bson:"is_main_barcode"`

	// Marketplace & Logistics
	PackageWeight       float64                  `json:"package_weight" bson:"package_weight"`
	PackageLength       float64                  `json:"package_length" bson:"package_length"`
	PackageWidth        float64                  `json:"package_width" bson:"package_width"`
	PackageHeight       float64                  `json:"package_height" bson:"package_height"`
	MarketplaceProducts *[]MarketplaceProductMap `json:"marketplace_products" bson:"marketplace_products"`

	ItemType     int8   `json:"item_type" bson:"item_type"`
	MaterialType int8   `json:"materialtype" bson:"materialtype"`
	TaxType      int8   `json:"tax_type" bson:"tax_type"`
	VatType      int8   `json:"vat_type" bson:"vat_type"`
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
	GuidFixed     string          `json:"guid_fixed" bson:"guid_fixed"`
	Names         *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode  string          `json:"item_unit_code" bson:"item_unit_code"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode       string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition     bool            `json:"condition" bson:"condition"`
	DivideValue   float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue    float64         `json:"standvalue" bson:"standvalue"`
	Qty           float64         `json:"qty" bson:"qty"`

	// Marketplace & SKU Logistics
	SellerSKU              string               `json:"seller_sku" bson:"seller_sku"`
	SkuPackageWeight       float64              `json:"sku_package_weight" bson:"sku_package_weight"`
	SkuPackageLength       float64              `json:"sku_package_length" bson:"sku_package_length"`
	SkuPackageWidth        float64              `json:"sku_package_width" bson:"sku_package_width"`
	SkuPackageHeight       float64              `json:"sku_package_height" bson:"sku_package_height"`
	MarketplaceSKUMappings *[]MarketplaceSKUMap `json:"marketplace_sku_mappings" bson:"marketplace_sku_mappings"`
}

type ProductBarcodeBOMVersion struct {
	GuidFixed string               `json:"guid_fixed" bson:"guid_fixed"`
	StartDate time.Time            `json:"start_date" bson:"start_date"`
	EndDate   *time.Time           `json:"end_date" bson:"end_date"`
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

type MarketplaceMediaAsset struct {
	Kind           string `json:"kind" bson:"kind"` // main, gallery, sku, size_chart, video, description
	URI            string `json:"uri" bson:"uri"`
	ExternalID     string `json:"external_id" bson:"external_id"`
	ExternalURL    string `json:"external_url" bson:"external_url"`
	OptionCode     string `json:"option_code" bson:"option_code"`
	OptionValue    string `json:"option_value" bson:"option_value"`
	SortOrder      int    `json:"sort_order" bson:"sort_order"`
	AltText        string `json:"alt_text" bson:"alt_text"`
	UseCase        string `json:"use_case" bson:"use_case"`
	MimeType       string `json:"mime_type" bson:"mime_type"`
	Width          int    `json:"width" bson:"width"`
	Height         int    `json:"height" bson:"height"`
	LastImportedAt string `json:"last_imported_at" bson:"last_imported_at"`
}

type MarketplaceAttributeValue struct {
	ValueID       string `json:"value_id" bson:"value_id"`
	ValueCode     string `json:"value_code" bson:"value_code"`
	ValueText     string `json:"value_text" bson:"value_text"`
	DisplayText   string `json:"display_text" bson:"display_text"`
	UnitCode      string `json:"unit_code" bson:"unit_code"`
	SortOrder     int    `json:"sort_order" bson:"sort_order"`
	IsCustomValue bool   `json:"is_custom_value" bson:"is_custom_value"`
}

type MarketplaceAttribute struct {
	AttributeID   string                       `json:"attribute_id" bson:"attribute_id"`
	AttributeCode string                       `json:"attribute_code" bson:"attribute_code"`
	AttributeName string                       `json:"attribute_name" bson:"attribute_name"`
	InputType     string                       `json:"input_type" bson:"input_type"` // text, single_select, multi_select, number, date, boolean
	Scope         string                       `json:"scope" bson:"scope"`           // product, sku, package, compliance
	IsRequired    bool                         `json:"is_required" bson:"is_required"`
	IsSaleProp    bool                         `json:"is_sale_prop" bson:"is_sale_prop"`
	IsCustom      bool                         `json:"is_custom" bson:"is_custom"`
	Values        *[]MarketplaceAttributeValue `json:"values" bson:"values"`
}

type MarketplaceSpecificationGroup struct {
	GroupCode  string                  `json:"group_code" bson:"group_code"`
	GroupName  string                  `json:"group_name" bson:"group_name"`
	Attributes *[]MarketplaceAttribute `json:"attributes" bson:"attributes"`
}

type MarketplacePayloadExample struct {
	Direction string `json:"direction" bson:"direction"` // import, export, stock_update, price_update, order_import
	UseCase   string `json:"use_case" bson:"use_case"`
	Payload   any    `json:"payload" bson:"payload"`
}

type ProductPrice struct {
	KeyNumber int     `json:"key_number" bson:"key_number"`
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
	HoldingCode              string `json:"holding_code" gorm:"column:holding_code;primaryKey"`
	models.PartitionIdentity `gorm:"embedded;"`
	Barcode                  string              `json:"barcode" gorm:"column:barcode;primaryKey"`
	Names                    JSONB               `json:"names"  gorm:"column:names;type:jsonb" `
	UnitCode                 string              `json:"item_unit_code" gorm:"column:unitcode"`
	UnitNames                JSONB               `json:"itemunitnames" gorm:"column:unitnames;type:jsonb"`
	BalanceQty               float64             `json:"balance_qty" gorm:"column:balance_qty"`
	MainBarcodeRef           string              `json:"mainbarcoderef" gorm:"column:mainbarcoderef"`
	StandValue               float64             `json:"standvalue" gorm:"column:standvalue"`
	DivideValue              float64             `json:"dividevalue" gorm:"column:dividevalue"`
	BalanceAmount            float64             `json:"balanceamount" gorm:"column:balanceamount"`
	AverageCost              float64             `json:"averagecost" gorm:"column:averagecost"`
	ItemCode                 string              `json:"itemcode" gorm:"column:itemcode"`
	ItemType                 int8                `json:"item_type" gorm:"column:itemtype"`
	MaterialType             int8                `json:"materialtype" gorm:"column:materialtype"`
	BrandCode                string              `json:"brand_code" gorm:"column:brand_code"`
	BrandName                JSONB               `json:"brandnames" gorm:"column:brandnames;type:jsonb"`
	CategoryCode             string              `json:"categorycode" gorm:"column:categorycode"`
	CategoryName             JSONB               `json:"category_names" gorm:"column:category_names;type:jsonb"`
	ClassCode                string              `json:"classcode" gorm:"column:classcode"`
	ClassNames               JSONB               `json:"classnames" gorm:"column:classnames;type:jsonb"`
	DesignCode               string              `json:"designcode" gorm:"column:designcode"`
	DesignNames              JSONB               `json:"designnames" gorm:"column:designnames;type:jsonb"`
	GradeCode                string              `json:"gradecode" gorm:"column:gradecode"`
	GradeNames               JSONB               `json:"gradenames" gorm:"column:gradenames;type:jsonb"`
	GroupCode                string              `json:"group_code" gorm:"column:group_code"`
	GroupNames               JSONB               `json:"group_names" gorm:"column:group_names;type:jsonb"`
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
	AccountID   string `json:"account_id" bson:"account_id"`     // seller/shop account on the platform
	HoldingCode string `json:"holding_code" bson:"holding_code"` // legacy field, kept for backward-compat
	// Item identity on the marketplace
	MarketItemID  string `json:"market_item_id" bson:"market_item_id"`   // Shopee item_id / Lazada item_id / TikTok product_id
	MarketModelID string `json:"market_model_id" bson:"market_model_id"` // variation/model/sku id for this listing
	ItemURL       string `json:"item_url" bson:"item_url"`
	// SKU identity
	SellerSKU string `json:"seller_sku" bson:"seller_sku"` // our SKU pushed to the platform
	ShopSKU   string `json:"shop_sku" bson:"shop_sku"`     // platform-generated SKU (Lazada ShopSku)
	GTIN      string `json:"gtin" bson:"gtin"`             // EAN/UPC/GTIN (TikTok mandatory for some categories)
	// Catalog mapping (each platform has its own category/brand tree)
	CategoryID   string `json:"category_id" bson:"category_id"`
	CategoryName string `json:"category_name" bson:"category_name"`
	BrandID      string `json:"brand_id" bson:"brand_id"`
	// Media and specifications are internal import/sync structures. Normal product UI should display
	// merged product images/specs, not expose marketplace origin labels.
	MediaAssets         *[]MarketplaceMediaAsset         `json:"media_assets" bson:"media_assets"`
	SpecificationGroups *[]MarketplaceSpecificationGroup `json:"specification_groups" bson:"specification_groups"`
	RawAttributes       *[]MarketplaceAttribute          `json:"raw_attributes" bson:"raw_attributes"`
	PayloadExamples     *[]MarketplacePayloadExample     `json:"payload_examples" bson:"payload_examples"`
	// Price & stock (current value on platform + sync intent)
	Currency      string  `json:"currency" bson:"currency"`
	CustomPrice   float64 `json:"custom_price" bson:"custom_price"`     // price we intend to push
	PlatformPrice float64 `json:"platform_price" bson:"platform_price"` // actual price on platform (inbound)
	PlatformStock int     `json:"platform_stock" bson:"platform_stock"` // actual stock on platform (inbound)
	SyncStock     bool    `json:"sync_stock" bson:"sync_stock"`
	SyncPrice     bool    `json:"sync_price" bson:"sync_price"`
	// Listing & sync status
	Status        string `json:"status" bson:"status"` // LIVE / UNLIST / REVIEWING / REJECTED / DELETED
	RejectReason  string `json:"reject_reason" bson:"reject_reason"`
	DaysToShip    int    `json:"days_to_ship" bson:"days_to_ship"`
	IsPreOrder    bool   `json:"is_pre_order" bson:"is_pre_order"`
	SyncEnabled   bool   `json:"sync_enabled" bson:"sync_enabled"`
	SyncStatus    string `json:"sync_status" bson:"sync_status"` // legacy field, kept for backward-compat
	LastSyncAt    string `json:"last_sync_at" bson:"last_sync_at"`
	LastSyncError string `json:"last_sync_error" bson:"last_sync_error"`
}

// MarketplaceSKUMap maps a sub/variation barcode (RefProductBarcode) to a
// marketplace SKU. Same unified field set as MarketplaceProductMap minus the
// item-level catalog fields, so sub-barcode variations sync consistently.
type MarketplaceDimensionStock struct {
	DimensionKey      string  `json:"dimension_key" bson:"dimension_key"`
	DimensionName     string  `json:"dimension_name" bson:"dimension_name"`
	MarketDimensionID string  `json:"market_dimension_id" bson:"market_dimension_id"`
	AvailableQty      float64 `json:"available_qty" bson:"available_qty"`
	ReservedQty       float64 `json:"reserved_qty" bson:"reserved_qty"`
	InboundQty        float64 `json:"inbound_qty" bson:"inbound_qty"`
	OversellBufferQty float64 `json:"oversell_buffer_qty" bson:"oversell_buffer_qty"`
	LastPlatformStock float64 `json:"last_platform_stock" bson:"last_platform_stock"`
	LastSyncedAt      string  `json:"last_synced_at" bson:"last_synced_at"`
	LastSyncStatus    string  `json:"last_sync_status" bson:"last_sync_status"`
	LastSyncError     string  `json:"last_sync_error" bson:"last_sync_error"`
}

type MarketplaceSKUMap struct {
	Platform                   string                       `json:"platform" bson:"platform"`
	AccountID                  string                       `json:"account_id" bson:"account_id"`
	HoldingCode                string                       `json:"holding_code" bson:"holding_code"`
	MarketItemID               string                       `json:"market_item_id" bson:"market_item_id"`
	MarketModelID              string                       `json:"market_model_id" bson:"market_model_id"`
	SellerSKU                  string                       `json:"seller_sku" bson:"seller_sku"`
	ShopSKU                    string                       `json:"shop_sku" bson:"shop_sku"`
	GTIN                       string                       `json:"gtin" bson:"gtin"`
	MediaAssets                *[]MarketplaceMediaAsset     `json:"media_assets" bson:"media_assets"`
	RawAttributes              *[]MarketplaceAttribute      `json:"raw_attributes" bson:"raw_attributes"`
	Currency                   string                       `json:"currency" bson:"currency"`
	SyncStock                  bool                         `json:"sync_stock" bson:"sync_stock"`
	SyncPrice                  bool                         `json:"sync_price" bson:"sync_price"`
	CustomPrice                float64                      `json:"custom_price" bson:"custom_price"`
	PlatformPrice              float64                      `json:"platform_price" bson:"platform_price"`
	PlatformStock              int                          `json:"platform_stock" bson:"platform_stock"`
	MarketplaceDimensionStocks *[]MarketplaceDimensionStock `json:"marketplace_dimension_stocks" bson:"marketplace_dimension_stocks"`
	Status                     string                       `json:"status" bson:"status"`
	SyncEnabled                bool                         `json:"sync_enabled" bson:"sync_enabled"`
	LastSyncAt                 string                       `json:"last_sync_at" bson:"last_sync_at"`
}
