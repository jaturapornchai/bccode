package models

import (
	"strings"

	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productCollectionName = "products"

type Product struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string                 `json:"code" bson:"code"`
	Names                    *[]models.NameX        `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	GroupCode                string                 `json:"groupcode" bson:"groupcode"`
	GroupNames               *[]models.NameX        `json:"groupnames" bson:"groupnames"`
	ManufacturerGUID         string                 `json:"manufacturerguid" bson:"manufacturerguid"`
	ManufacturerCode         string                 `json:"manufacturercode" bson:"manufacturercode"`
	ManufacturerNames        *[]models.NameX        `json:"manufacturernames" bson:"manufacturernames"`
	Dimensions               []ProductDimension     `json:"dimensions" bson:"dimensions"`
	VatType                  int8                   `json:"vattype" bson:"vattype"`
	Barcodes                 []Barcodes             `json:"barcodes,omitempty" bson:"-"`
	ItemType                 int8                   `json:"itemtype" bson:"itemtype"`
	UnitGuid                 string                 `json:"unitguid" bson:"unitguid"`
	UnitCode                 string                 `json:"unitcode" bson:"unitcode" validate:"required,max=100"`
	UnitNames                *[]models.NameX        `json:"unitnames" bson:"unitnames"`
	GroupsuboneGuid          string                 `json:"groupsuboneguid" bson:"groupsuboneguid"`
	GroupsuboneCode          string                 `json:"groupsubonecode" bson:"groupsubonecode"`
	GroupsuboneNames         *[]models.NameX        `json:"groupsubonenames" bson:"groupsubonenames"`
	GroupsubtwoGuid          string                 `json:"groupsubtwoguid" bson:"groupsubtwoguid"`
	GroupsubtwoCode          string                 `json:"groupsubtwocode" bson:"groupsubtwocode"`
	GroupsubtwoNames         *[]models.NameX        `json:"groupsubtwonames" bson:"groupsubtwonames"`
	BrandGuid                string                 `json:"brandguid" bson:"brandguid"`
	BrandCode                string                 `json:"brandcode" bson:"brandcode"`
	BrandNames               *[]models.NameX        `json:"brandnames" bson:"brandnames"`
	DesignGuid               string                 `json:"designguid" bson:"designguid"`
	DesignCode               string                 `json:"designcode" bson:"designcode"`
	DesignNames              *[]models.NameX        `json:"designnames" bson:"designnames"`
	ModelGuid                string                 `json:"modelguid" bson:"modelguid"`
	ModelCode                string                 `json:"modelcode" bson:"modelcode"`
	ModelNames               *[]models.NameX        `json:"modelnames" bson:"modelnames"`
	PatternGuid              string                 `json:"patternguid" bson:"patternguid"`
	PatternCode              string                 `json:"patterncode" bson:"patterncode"`
	PatternNames             *[]models.NameX        `json:"patternnames" bson:"patternnames"`
	GradeGuid                string                 `json:"gradeguid" bson:"gradeguid"`
	GradeCode                string                 `json:"gradecode" bson:"gradecode"`
	GradeNames               *[]models.NameX        `json:"gradenames" bson:"gradenames"`
	CategoryGuid             string                 `json:"categoryguid" bson:"categoryguid"`
	CategoryCode             string                 `json:"categorycode" bson:"categorycode"`
	CategoryNames            *[]models.NameX        `json:"categorynames" bson:"categorynames"`
	ClassGuid                string                 `json:"classguid" bson:"classguid"`
	ClassCode                string                 `json:"classcode" bson:"classcode"`
	ClassNames               *[]models.NameX        `json:"classnames" bson:"classnames"`
	MaterialType             int8                   `json:"materialtype" bson:"materialtype"`
	TaxType                  int8                   `json:"taxtype" bson:"taxtype"`
	Manufacturers            *[]ProductManufacturer `json:"manufacturers" bson:"manufacturers"`
	Suppliers                *[]ProductSupplier     `json:"suppliers" bson:"suppliers"`

	// Core Product Properties Moved from ProductBarcode
	ImageURI             string                        `json:"imageuri" bson:"imageuri"`
	Images               *[]ProductImage               `json:"images" bson:"images"`
	Videos               *[]ProductVideo               `json:"videos" bson:"videos"`
	UseImageOrColor      bool                          `json:"useimageorcolor" bson:"useimageorcolor"`
	ColorSelect          string                        `json:"colorselect" bson:"colorselect"`
	ColorSelectHex       string                        `json:"colorselecthex" bson:"colorselecthex"`
	IsSumPoint           bool                          `json:"issumpoint" bson:"issumpoint"`
	IsALaCarte           bool                          `json:"isalacarte" bson:"isalacarte"`
	IsSplitUnitPrint     bool                          `json:"issplitunitprint" bson:"issplitunitprint"`
	IsOnlyStaff          bool                          `json:"isonlystaff" bson:"isonlystaff"`
	FoodType             int                           `json:"foodtype" bson:"foodtype"`
	IsStockForRestaurant bool                          `json:"isstockforrestaurant" bson:"isstockforrestaurant"`
	Restaurant           ProductRestaurant             `json:"restaurant" bson:"restaurant"`
	OrderTypes           *[]ProductOrderType           `json:"ordertypes" bson:"ordertypes"`
	Options              *[]ProductOption              `json:"options" bson:"options"`
	IsAlert              bool                          `json:"isalert" bson:"isalert"`
	AlertDescription     string                        `json:"alertdescription" bson:"alertdescription" validate:"max=1500"`
	Description          string                        `json:"description" bson:"description" validate:"max=1500"`
	TimeForSales         *[]ProductTimeForSale         `json:"timeforsales" bson:"timeforsales"`
	BusinessTypes        *[]ProductBarcodeBusinessType `json:"businesstypes" bson:"businesstypes"`
	IgnoreBranches       *[]ProductBarcodeBranch       `json:"ignorebranches" bson:"ignorebranches"`

	// Units and BOM properties moved from ProductBarcode to Product
	Condition        bool                    `json:"condition" bson:"condition"`
	DivideValue      int64                   `json:"dividevalue" bson:"dividevalue"`
	StandValue       int64                   `json:"standvalue" bson:"standvalue"`
	UnitConversions  []ProductUnitConversion `json:"unitconversions" bson:"unitconversions"`
	IsUseSubBarcodes bool                    `json:"isusesubbarcodes" bson:"isusesubbarcodes"`
	RefBarcodes      *[]RefProductBarcode    `json:"refbarcodes" bson:"refbarcodes"`
	BOM              *[]BOMProductBarcode    `json:"bom" bson:"bom"`

	// Marketplace & Logistics
	PackageWeight       float64                  `json:"packageweight" bson:"packageweight"`
	PackageLength       float64                  `json:"packagelength" bson:"packagelength"`
	PackageWidth        float64                  `json:"packagewidth" bson:"packagewidth"`
	PackageHeight       float64                  `json:"packageheight" bson:"packageheight"`
	MarketplaceProducts *[]MarketplaceProductMap `json:"marketplaceproducts" bson:"marketplaceproducts"`

	// ชั้นลงขาย (ช่องทางขายออนไลน์) — omitempty จำเป็น เพราะการบันทึกฝั่งบัญชี $set ทั้ง struct จะได้ไม่ลบค่าเหล่านี้ทิ้ง
	ImageURIThumb string          `json:"imageurithumb,omitempty" bson:"imageurithumb,omitempty"`
	Listing       *ProductListing `json:"listing,omitempty" bson:"listing,omitempty"`

	// Stock properties
	OrderPoint   float64 `json:"orderpoint" bson:"orderpoint"`
	MinPoint     float64 `json:"minpoint" bson:"minpoint"`
	MaxPoint     float64 `json:"maxpoint" bson:"maxpoint"`
	Qty          float64 `json:"qty" bson:"qty"`
	StockBarcode string  `json:"stockbarcode" bson:"stockbarcode"`
}

type ProductUnitConversion struct {
	UnitCode    string          `json:"unitcode" bson:"unitcode"`
	UnitNames   *[]models.NameX `json:"unitnames" bson:"unitnames"`
	DivideValue int64           `json:"dividevalue" bson:"dividevalue"`
	StandValue  int64           `json:"standvalue" bson:"standvalue"`
}

func (product Product) UnitRatio(unitCode string) (int64, int64, bool) {
	unit, ok := product.UnitDefinition(unitCode)
	if !ok {
		return 0, 0, false
	}
	return unit.DivideValue, unit.StandValue, true
}

func (product Product) UnitDefinition(unitCode string) (ProductUnitConversion, bool) {
	unitCode = strings.TrimSpace(unitCode)
	if unitCode != "" && strings.EqualFold(unitCode, product.UnitCode) {
		return ProductUnitConversion{
			UnitCode:    product.UnitCode,
			UnitNames:   product.UnitNames,
			DivideValue: 1,
			StandValue:  1,
		}, true
	}
	for _, unit := range product.UnitConversions {
		if strings.EqualFold(unitCode, unit.UnitCode) {
			return unit, true
		}
	}
	return ProductUnitConversion{}, false
}

type RefProductBarcode struct {
	GuidFixed     string          `json:"guidfixed" bson:"guidfixed"`
	ItemCode      string          `json:"itemcode" bson:"itemcode"`
	Names         *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode  string          `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode       string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition     bool            `json:"condition" bson:"condition"`
	DivideValue   int64           `json:"dividevalue" bson:"dividevalue"`
	StandValue    int64           `json:"standvalue" bson:"standvalue"`
	Qty           float64         `json:"qty" bson:"qty"`

	// Marketplace & SKU Logistics
	SellerSKU              string               `json:"sellersku" bson:"sellersku"`
	SkuPackageWeight       float64              `json:"skupackageweight" bson:"skupackageweight"`
	SkuPackageLength       float64              `json:"skupackagelength" bson:"skupackagelength"`
	SkuPackageWidth        float64              `json:"skupackagewidth" bson:"skupackagewidth"`
	SkuPackageHeight       float64              `json:"skupackageheight" bson:"skupackageheight"`
	MarketplaceSKUMappings *[]MarketplaceSKUMap `json:"marketplaceskumappings" bson:"marketplaceskumappings"`
}

type BOMProductBarcode struct {
	BarcodeGuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	ItemCode         string          `json:"itemcode" bson:"itemcode"`
	Level            int             `json:"level" bson:"level"`
	Names            *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode     string          `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames    *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	Condition        bool            `json:"condition" bson:"condition"`
	DivideValue      int64           `json:"dividevalue" bson:"dividevalue"`
	StandValue       int64           `json:"standvalue" bson:"standvalue"`
	Qty              float64         `json:"qty" bson:"qty"`
}

type ProductManufacturer struct {
	GuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Code      string          `json:"code" bson:"code"`
	Names     *[]models.NameX `json:"names" bson:"names"`
}

type ProductSupplier struct {
	GuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Code      string          `json:"code" bson:"code"`
	Names     *[]models.NameX `json:"names" bson:"names"`
}

type Barcodes struct {
	GuidFixed     string          `json:"guidfixed" gorm:"-"`
	ItemUnitCode  string          `json:"itemunitcode" gorm:"-"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames" gorm:"-"`
	Barcode       string          `json:"barcode" gorm:"-"`
	ImageURI      string          `json:"imageuri" gorm:"-"`
	Images        *[]ProductImage `json:"images" gorm:"-"`
	Videos        *[]ProductVideo `json:"videos" gorm:"-"`
	Description   string          `json:"description" gorm:"-"`
	Prices        *[]ProductPrice `json:"prices" gorm:"-"`
	Condition     bool            `json:"condition" gorm:"-"`
	DivideValue   int64           `json:"dividevalue" gorm:"-"`
	StandValue    int64           `json:"standvalue" gorm:"-"`
	Qty           float64         `json:"qty" gorm:"-"`
}

type ProductPrice struct {
	KeyNumber int     `json:"keynumber" gorm:"-"`
	Price     float64 `json:"price" gorm:"-"`
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

type ProductInfo struct {
	models.DocIdentity `bson:"inline"`
	Product            `bson:"inline"`
}

func (ProductInfo) CollectionName() string {
	return productCollectionName
}

type ProductData struct {
	models.HoldingCodeentity `bson:"inline"`
	BusinessCode             string `json:"businesscode" bson:"businesscode"`
	ProductInfo              `bson:"inline"`
}

type ProductDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductData        `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ProductDoc) CollectionName() string {
	return productCollectionName
}

type ProductItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (ProductItemGuid) CollectionName() string {
	return productCollectionName
}

type ProductActivity struct {
	ProductData         `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductActivity) CollectionName() string {
	return productCollectionName
}

type ProductDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductDeleteActivity) CollectionName() string {
	return productCollectionName
}

type ProductImage struct {
	XOrder   int    `json:"xorder" bson:"xorder"`
	URI      string `json:"uri" bson:"uri"`
	URIThumb string `json:"urithumb,omitempty" bson:"urithumb,omitempty"` // รูปย่อ (ชั้นลงขาย)
}

type ProductVideo struct {
	XOrder      int    `json:"xorder" bson:"xorder"`
	URI         string `json:"uri" bson:"uri"`
	PosterURI   string `json:"posteruri" bson:"posteruri"`
	DurationSec int    `json:"durationsec,omitempty" bson:"durationsec,omitempty"` // ความยาววิดีโอ (วินาที)
	SizeBytes   int64  `json:"sizebytes,omitempty" bson:"sizebytes,omitempty"`     // ขนาดไฟล์ (ไบต์)
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

type ProductRestaurant struct {
	IsForRestaurant       bool `json:"isforrestaurant" bson:"isforrestaurant"`
	IsForTakeAway         bool `json:"isfortakeaway" bson:"isfortakeaway"`
	IsForDelivery         bool `json:"isfordelivery" bson:"isfordelivery"`
	IsForCustomer         bool `json:"isforcustomer" bson:"isforcustomer"`
	IsForCustomerPreOrder bool `json:"isforcustomerpreorder" bson:"isforcustomerpreorder"`
}

type ProductOrderType struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
	Price              float64         `json:"price" bson:"price"`
}

type ProductChoice struct {
	GUID            string          `json:"guid" bson:"guid"`
	RefBarcode      string          `json:"refbarcode" bson:"refbarcode"`
	RefProductCode  string          `json:"refproductcode" bson:"refproductcode"`
	RefBarcodeNames *[]models.NameX `json:"refbarcodenames" bson:"refbarcodenames"`
	Names           *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	RefUnitCode     string          `json:"refunitcode" bson:"refunitcode"`
	RefUnitNames    *[]models.NameX `json:"refunitnames" bson:"refunitnames" validate:"required,min=1,unique=Code,dive"`
	Price           *string         `json:"price" bson:"price"`
	Qty             float64         `json:"qty" bson:"qty"`
	ImageURI        string          `json:"imageuri" bson:"imageuri"`
	IsStock         bool            `json:"isstock" bson:"isstock"`
	IsDefault       bool            `json:"isdefault" bson:"isdefault"`
	VatCal          int8            `json:"vatcal" bson:"vatcal"`
}

type ProductOption struct {
	GUID       string           `json:"guid" bson:"guid"`
	ChoiceType int8             `json:"choicetype" bson:"choicetype"`
	MaxSelect  uint16           `json:"maxselect" bson:"maxselect" validate:"min=0,max=60000"`
	MinSelect  uint16           `json:"minselect" bson:"minselect" validate:"min=0,max=60000"`
	Names      *[]models.NameX  `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Choices    *[]ProductChoice `json:"choices" bson:"choices"`
}

type ProductTimeForSale struct {
	DaysOfWeek []int8 `json:"daysofweek" bson:"daysofweek"`
	FromDate   string `json:"fromdate" bson:"fromdate"`
	ToDate     string `json:"todate" bson:"todate"`
	FromTime   string `json:"fromtime" bson:"fromtime"`
	ToTime     string `json:"totime" bson:"totime"`
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
