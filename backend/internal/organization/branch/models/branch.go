package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const branchCollectionName = "organizationbranches"

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
	CompanyRegistrationNo    string                  `json:"companyregistrationno" bson:"companyregistrationno"`
	IsVatRegistered          bool                    `json:"isvatregistered" bson:"isvatregistered"`
	PaymentRounding          PaymentRoundingSettings `json:"paymentrounding" bson:"paymentrounding"`
	PointConfig              PointConfig             `json:"pointconfig" bson:"pointconfig"`
	IsMainShop               bool                    `json:"ismainshop" bson:"ismainshop"`
	MainHoldingCode          string                  `json:"mainholdingcode" bson:"mainholdingcode"`
	ProductCenterType        int8                    `json:"productcentertype" bson:"productcentertype"`
	DebtorCenterType         int8                    `json:"debtorcentertype" bson:"debtorcentertype"`
	CouponUseType            int8                    `json:"couponusetype" bson:"couponusetype"`
	BaseCurrency             string                  `json:"basecurrency" bson:"basecurrency"`
	Language                 string                  `json:"language" bson:"language"`
	Timezone                 string                  `json:"timezone" bson:"timezone"`
	DateFormat               string                  `json:"dateformat" bson:"dateformat"`
	YearType                 string                  `json:"yeartype" bson:"yeartype"`
	TimezoneLabel            string                  `json:"timezonelabel" bson:"timezonelabel"`
	TimezoneOffset           string                  `json:"timezoneoffset" bson:"timezoneoffset"`
	DecimalQuantity          int8                    `json:"decimalquantity" bson:"decimalquantity"`
	DecimalPrice             int8                    `json:"decimalprice" bson:"decimalprice"`
	DecimalDocument          int8                    `json:"decimaldocument" bson:"decimaldocument"`
	IsRestaurant             bool                    `json:"isrestaurant" bson:"isrestaurant"`
	IsTire                   bool                    `json:"istire" bson:"istire"`
	IsAgriculture            bool                    `json:"isagriculture" bson:"isagriculture"`
	IsPharmacy               bool                    `json:"ispharmacy" bson:"ispharmacy"`
	IsRetail                 bool                    `json:"isretail" bson:"isretail"`
	IsService                bool                    `json:"isservice" bson:"isservice"`
	IsWholesale              bool                    `json:"iswholesale" bson:"iswholesale"`
	IsManufacturing          bool                    `json:"ismanufacturing" bson:"ismanufacturing"`
	IsImportExport           bool                    `json:"isimportexport" bson:"isimportexport"`
	IsContractor             bool                    `json:"iscontractor" bson:"iscontractor"`
	IsRental                 bool                    `json:"isrental" bson:"isrental"`
	IsEcommerce              bool                    `json:"isecommerce" bson:"isecommerce"`
	IsLogistics              bool                    `json:"islogistics" bson:"islogistics"`
	IsEducation              bool                    `json:"iseducation" bson:"iseducation"`
	IsHotel                  bool                    `json:"ishotel" bson:"ishotel"`
	IsBeauty                 bool                    `json:"isbeauty" bson:"isbeauty"`
	IsGoldShop               bool                    `json:"isgoldshop" bson:"isgoldshop"`
	IsAccountingFirm         bool                    `json:"isaccountingfirm" bson:"isaccountingfirm"`
	IsConstruction           bool                    `json:"isconstruction" bson:"isconstruction"`
	IsElectronics            bool                    `json:"iselectronics" bson:"iselectronics"`
	IsMobileShop             bool                    `json:"ismobileshop" bson:"ismobileshop"`
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
	TaxID               string  `json:"taxid" bson:"taxid"`
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
	CountryCode     string         `json:"countrycode" bson:"countrycode"`
	ProvinceCode    string         `json:"provincecode" bson:"provincecode"`
	DistrictCode    string         `json:"districtcode" bson:"districtcode"`
	SubDistrictCode string         `json:"subdistrictcode" bson:"subdistrictcode"`
	ZipCode         string         `json:"zipcode" bson:"zipcode"`
	PhoneNumber     string         `json:"phonenumber" bson:"phonenumber"`
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
	models.HoldingCodeentity `bson:"inline"`
	BranchInfo               `bson:"inline"`
}

type BranchDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"id,omitempty"`
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
	// GuidFixed string         `json:"guidfixed"`
	Code  string         `json:"code"`
	Names []models.NameX `json:"names"`
}

type BusinessType struct {
	GuidFixed string         `json:"guidfixed"`
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
	StartDate   time.Time `json:"startdate" bson:"startdate"`
	EndDate     time.Time `json:"enddate" bson:"enddate"`
	PayPerPoint float64   `json:"payperpoint" bson:"payperpoint"`
	PointValue  float64   `json:"pointvalue" bson:"pointvalue"`
}

type PointSpecialRule struct {
	StartDate       time.Time `json:"startdate" bson:"startdate"`
	EndDate         time.Time `json:"enddate" bson:"enddate"`
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
