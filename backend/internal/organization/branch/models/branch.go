package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const branchCollectionName = "organizationBranches"

type Branch struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string                  `json:"code" bson:"code"`
	CompanyNames             *[]models.NameX         `json:"companynames" bson:"companynames"`
	MachineType              int                     `json:"machinetype" bson:"machinetype"`
	Names                    *[]models.NameX         `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Departments              *[]Department           `json:"departments" bson:"departments"`
	BusinessTypes            *[]string               `json:"businesstypes" bson:"businesstypes"`
	ImageURI                 string                  `json:"imageuri" bson:"imageuri"`
	ImageURIs                []string                `json:"imageuris" bson:"imageuris"`
	LogoURI                  string                  `json:"logouri" bson:"logouri"`
	Languages                *[]string               `json:"languages" bson:"languages"`
	Contact                  Contact                 `json:"contact" bson:"contact"`
	POS                      BranchPOS               `json:"pos" bson:"pos"`
	BusinessType             BranchBusinessType      `json:"businesstype" bson:"businesstype"`
	CompanyRegistrationNo    string                  `json:"company_registration_no" bson:"company_registration_no"`
	IsVatRegistered          bool                    `json:"is_vat_registered" bson:"is_vat_registered"`
	PaymentRounding          PaymentRoundingSettings `json:"paymentrounding" bson:"paymentrounding"`
	PointConfig              PointConfig             `json:"pointconfig" bson:"pointconfig"`
	IsMainShop               bool                    `json:"ismainshop" bson:"ismainshop"`
	MainShopId               string                  `json:"main_shop_id" bson:"main_shop_id"`
	ProductCenterType        int8                    `json:"productcentertype" bson:"productcentertype"`
	DebtorCenterType         int8                    `json:"debtorcentertype" bson:"debtorcentertype"`
	CouponUseType            int8                    `json:"couponusetype" bson:"couponusetype"`
	BaseCurrency             string                  `json:"base_currency" bson:"base_currency"`
	Language                 string                  `json:"language" bson:"language"`
	Timezone                 string                  `json:"timezone" bson:"timezone"`
	DateFormat               string                  `json:"date_format" bson:"date_format"`
	YearType                 string                  `json:"year_type" bson:"year_type"`
	TimezoneLabel            string                  `json:"timezone_label" bson:"timezone_label"`
	TimezoneOffset           string                  `json:"timezone_offset" bson:"timezone_offset"`
	DecimalQuantity          int8                    `json:"decimal_quantity" bson:"decimal_quantity"`
	DecimalPrice             int8                    `json:"decimal_price" bson:"decimal_price"`
	DecimalDocument          int8                    `json:"decimal_document" bson:"decimal_document"`
	IsRestaurant             bool                    `json:"is_restaurant" bson:"is_restaurant"`
	IsTire                   bool                    `json:"is_tire" bson:"is_tire"`
	IsAgriculture            bool                    `json:"is_agriculture" bson:"is_agriculture"`
	IsPharmacy               bool                    `json:"is_pharmacy" bson:"is_pharmacy"`
	IsRetail                 bool                    `json:"is_retail" bson:"is_retail"`
	IsService                bool                    `json:"is_service" bson:"is_service"`
	IsWholesale              bool                    `json:"is_wholesale" bson:"is_wholesale"`
	IsManufacturing          bool                    `json:"is_manufacturing" bson:"is_manufacturing"`
	IsImportExport           bool                    `json:"is_import_export" bson:"is_import_export"`
	IsContractor             bool                    `json:"is_contractor" bson:"is_contractor"`
	IsRental                 bool                    `json:"is_rental" bson:"is_rental"`
	IsEcommerce              bool                    `json:"is_ecommerce" bson:"is_ecommerce"`
	IsLogistics              bool                    `json:"is_logistics" bson:"is_logistics"`
	IsEducation              bool                    `json:"is_education" bson:"is_education"`
	IsHotel                  bool                    `json:"is_hotel" bson:"is_hotel"`
	IsBeauty                 bool                    `json:"is_beauty" bson:"is_beauty"`
	IsGoldShop               bool                    `json:"is_gold_shop" bson:"is_gold_shop"`
	IsAccountingFirm         bool                    `json:"is_accounting_firm" bson:"is_accounting_firm"`
	IsConstruction           bool                    `json:"is_construction" bson:"is_construction"`
	IsElectronics            bool                    `json:"is_electronics" bson:"is_electronics"`
	IsMobileShop             bool                    `json:"is_mobile_shop" bson:"is_mobile_shop"`
}

type PaymentRoundingRule struct {
	LowerBound float64 `json:"lowerbound" bson:"lowerbound"`
	UpperBound float64 `json:"upperbound" bson:"upperbound"`
	RoundTo    float64 `json:"roundto" bson:"roundto"`
}

type PaymentMethodRounding struct {
	Enabled bool                  `json:"enabled" bson:"enabled"`
	Rules   []PaymentRoundingRule `json:"rules" bson:"rules"`
}

type PaymentRoundingSettings struct {
	Cash         PaymentMethodRounding `json:"cash" bson:"cash"`
	BankTransfer PaymentMethodRounding `json:"banktransfer" bson:"banktransfer"`
	CreditCard   PaymentMethodRounding `json:"creditcard" bson:"creditcard"`
	Cheque       PaymentMethodRounding `json:"cheque" bson:"cheque"`
	Coupon       PaymentMethodRounding `json:"coupon" bson:"coupon"`
	QRCode       PaymentMethodRounding `json:"qrcode" bson:"qrcode"`
	Delivery     PaymentMethodRounding `json:"delivery" bson:"delivery"`
}

type BranchBusinessType struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type BranchPOS struct {
	TaxID               string  `json:"tax_id" bson:"tax_id"`
	IsBom               bool    `json:"isbom" bson:"isbom"`
	VatRate             float64 `json:"vatrate" bson:"vatrate"`
	VatTypeSale         int8    `json:"vattypesale" bson:"vattypesale"`
	VatTypePurchase     int8    `json:"vattypepurchase" bson:"vattypepurchase"`
	InquiryTypeSale     int8    `json:"inquirytypesale" bson:"inquirytypesale"`
	InquiryTypePurchase int8    `json:"inquirytypepurchase" bson:"inquirytypepurchase"`
	HeaderReceiptPOS    string  `json:"headerreceiptpos" bson:"headerreceiptpos"`
	FooterReceiptPOS    string  `json:"footerreceiptpos" bson:"footerreceiptpos"`
}

type Contact struct {
	Address         []models.NameX `json:"address" bson:"addressx"`
	CountryCode     string         `json:"country_code" bson:"country_code"`
	ProvinceCode    string         `json:"province_code" bson:"province_code"`
	DistrictCode    string         `json:"district_code" bson:"district_code"`
	SubDistrictCode string         `json:"sub_district_code" bson:"sub_district_code"`
	ZipCode         string         `json:"zip_code" bson:"zip_code"`
	PhoneNumber     string         `json:"phone_number" bson:"phone_number"`
	Latitude        float64        `json:"latitude" bson:"latitude"`
	Longitude       float64        `json:"longitude" bson:"longitude"`
}

type BranchInfo struct {
	models.DocIdentity `bson:"inline"`
	Branch             `bson:"inline"`
}

func (BranchInfo) CollectionName() string {
	return branchCollectionName
}

type BranchData struct {
	models.ShopIdentity `bson:"inline"`
	BranchInfo          `bson:"inline"`
}

type BranchDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	BranchData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (BranchDoc) CollectionName() string {
	return branchCollectionName
}

type BranchItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (BranchItemGuid) CollectionName() string {
	return branchCollectionName
}

type BranchActivity struct {
	BranchData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BranchActivity) CollectionName() string {
	return branchCollectionName
}

type BranchDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (BranchDeleteActivity) CollectionName() string {
	return branchCollectionName
}

type BranchInfoResponse struct {
	BranchInfo
	// Departments   []Department   `json:"departments" bson:"departments"`
	BusinessTypes []BusinessType `json:"businesstypes" bson:"businesstypes"`
}

type Department struct {
	// GuidFixed string         `json:"guid_fixed"`
	Code  string         `json:"code"`
	Names []models.NameX `json:"names"`
}

type BusinessType struct {
	GuidFixed string         `json:"guid_fixed"`
	Code      string         `json:"code"`
	Names     []models.NameX `json:"names"`
}

type PointConfig struct {
	GeneralRules   []PointGeneralRule `json:"generalrules" bson:"generalrules" default:"[]"`
	SpecialRules   []PointSpecialRule `json:"specialrules" bson:"specialrules" default:"[]"`
	PointUsageType int8               `json:"pointusagetype" bson:"pointusagetype" default:"1"` // 1 = discount, 2 = cash
}

// Point Usage Type Constants
const (
	PointTypeDiscount = 1 // Points used as discount (ใช้พอยท์เป็นส่วนลด)
	PointTypeCash     = 2 // Points used as cash (ใช้พอยท์เป็นเงินสด)
)

type PointGeneralRule struct {
	StartDate   time.Time `json:"start_date" bson:"start_date"`
	EndDate     time.Time `json:"end_date" bson:"end_date"`
	PayPerPoint float64   `json:"payperpoint" bson:"payperpoint"`
	PointValue  float64   `json:"pointvalue" bson:"pointvalue"`
}

type PointSpecialRule struct {
	StartDate       time.Time `json:"start_date" bson:"start_date"`
	EndDate         time.Time `json:"end_date" bson:"end_date"`
	Multiplier      float64   `json:"multiplier" bson:"multiplier"`
	Sunday          bool      `json:"sunday" bson:"sunday"`
	Monday          bool      `json:"monday" bson:"monday"`
	Tuesday         bool      `json:"tuesday" bson:"tuesday"`
	Wednesday       bool      `json:"wednesday" bson:"wednesday"`
	Thursday        bool      `json:"thursday" bson:"thursday"`
	Friday          bool      `json:"friday" bson:"friday"`
	Saturday        bool      `json:"saturday" bson:"saturday"`
	MaxPointPerBill float64   `json:"maxpointperbill" bson:"maxpointperbill"`
}
