package fixedasset

import (
	"strings"
	"time"
)

type Scope struct {
	Holding string
	Company string
	Branch  string
	Actor   string
}

type Name struct {
	Code string `json:"code" bson:"code"`
	Name string `json:"name" bson:"name"`
}

type Identity struct {
	ID           string    `json:"id" bson:"_id"`
	HoldingCode  string    `json:"holdingcode" bson:"holdingcode"`
	BusinessCode string    `json:"businesscode" bson:"businesscode"`
	Version      int64     `json:"version" bson:"__v"`
	CreatedAt    time.Time `json:"createdat" bson:"createdat"`
	CreatedBy    string    `json:"createdby" bson:"createdby"`
	UpdatedAt    time.Time `json:"updatedat" bson:"updatedat"`
	UpdatedBy    string    `json:"updatedby" bson:"updatedby"`
	IsDeleted    bool      `json:"isdeleted" bson:"isdeleted"`
}

type AssetType struct {
	Identity                 `bson:",inline"`
	TypeCode                 string `json:"typecode" bson:"typecode"`
	Names                    []Name `json:"names" bson:"names"`
	DefaultUsefulLifeYears   int    `json:"defaultusefullifeyears" bson:"defaultusefullifeyears"`
	DefaultDeprecPercent     Amount `json:"defaultdeprecpercent" bson:"defaultdeprecpercent"`
	AssetAccountCode         string `json:"assetaccountcode" bson:"assetaccountcode"`
	AccumDeprecAccountCode   string `json:"accumdeprecaccountcode" bson:"accumdeprecaccountcode"`
	DeprecExpenseAccountCode string `json:"deprecexpenseaccountcode" bson:"deprecexpenseaccountcode"`
	IsActive                 bool   `json:"isactive" bson:"isactive"`
}

func (AssetType) CollectionName() string { return "asset_types" }

func (at AssetType) ThaiName() string {
	for _, n := range at.Names {
		if n.Code == "th" {
			return n.Name
		}
	}
	return at.TypeCode
}

type Asset struct {
	Identity                 `bson:",inline"`
	AssetCode                string `json:"assetcode" bson:"assetcode"`
	Names                    []Name `json:"names" bson:"names"`
	AssetTypeCode            string `json:"assettypecode" bson:"assettypecode"`
	BranchCode               string `json:"branchcode" bson:"branchcode"`
	DepartmentCode           string `json:"departmentcode" bson:"departmentcode"`
	LocationCode             string `json:"locationcode" bson:"locationcode"`
	PurchaseDate             string `json:"purchasedate" bson:"purchasedate"` // YYYY-MM-DD
	StartCalcDate            string `json:"startcalcdate" bson:"startcalcdate"`
	Cost                     Amount `json:"cost" bson:"cost"`
	ScrapValue               Amount `json:"scrapvalue" bson:"scrapvalue"`
	UsefulLifeYears          int    `json:"usefullifeyears" bson:"usefullifeyears"`
	DeprecPercent            Amount `json:"deprecpercent" bson:"deprecpercent"`
	Method                   string `json:"method" bson:"method"` // straight_line, sum_of_years, declining
	FirstYearPercent         Amount `json:"firstyearpercent" bson:"firstyearpercent"` // e.g. 40% tax initial allowance
	BeginAccumDeprec         Amount `json:"beginaccumdeprec" bson:"beginaccumdeprec"`
	AssetAccountCode         string `json:"assetaccountcode" bson:"assetaccountcode"`
	AccumDeprecAccountCode   string `json:"accumdeprecaccountcode" bson:"accumdeprecaccountcode"`
	DeprecExpenseAccountCode string `json:"deprecexpenseaccountcode" bson:"deprecexpenseaccountcode"`
	Status                   string `json:"status" bson:"status"` // active, disposed, written_off
	SerialNumber             string `json:"serialnumber" bson:"serialnumber"`
	Brand                    string `json:"brand" bson:"brand"`
	Model                    string `json:"model" bson:"model"`
	SupplierCode             string `json:"suppliercode" bson:"suppliercode"`
	IsTaxDeductible          bool   `json:"istaxdeductible" bson:"istaxdeductible"`
	Notes                    string `json:"notes" bson:"notes"`
}

func (Asset) CollectionName() string { return "fixed_assets" }

func (a Asset) ThaiName() string {
	for _, n := range a.Names {
		if n.Code == "th" {
			return n.Name
		}
	}
	return a.AssetCode
}

type DepreciationScheduleItem struct {
	Identity     `bson:",inline"`
	AssetCode    string     `json:"assetcode" bson:"assetcode"`
	FiscalYear   string     `json:"fiscalyear" bson:"fiscalyear"`
	Period       int        `json:"period" bson:"period"` // 1-12
	StartDate    string     `json:"startdate" bson:"startdate"`
	StopDate     string     `json:"stopdate" bson:"stopdate"`
	Days         int        `json:"days" bson:"days"`
	PeriodDeprec Amount     `json:"perioddeprec" bson:"perioddeprec"`
	AccumDeprec  Amount     `json:"accumdeprec" bson:"accumdeprec"`
	NetBookValue Amount     `json:"netbookvalue" bson:"netbookvalue"`
	IsPosted     bool       `json:"isposted" bson:"isposted"`
	JournalDocNo string     `json:"journaldocno" bson:"journaldocno"`
	PostedAt     *time.Time `json:"postedat,omitempty" bson:"postedat,omitempty"`
}

func (DepreciationScheduleItem) CollectionName() string { return "asset_depreciations" }

type AssetDisposal struct {
	Identity               `bson:",inline"`
	DocNo                  string `json:"docno" bson:"docno"`
	AssetCode              string `json:"assetcode" bson:"assetcode"`
	DisposalDate           string `json:"disposaldate" bson:"disposaldate"`
	DisposalType           string `json:"disposaltype" bson:"disposaltype"` // sale, write_off, scrap
	SalePrice              Amount `json:"saleprice" bson:"saleprice"`
	VatAmount              Amount `json:"vatamount" bson:"vatamount"`
	AccumDeprecAtDisposal  Amount `json:"accumdeprecatdisposal" bson:"accumdeprecatdisposal"`
	NetBookValueAtDisposal Amount `json:"netbookvalueatdisposal" bson:"netbookvalueatdisposal"`
	GainLoss               Amount `json:"gainloss" bson:"gainloss"` // SalePrice - NetBookValue
	SettlementAccountCode  string `json:"settlementaccountcode" bson:"settlementaccountcode"` // Cash or AR
	GainLossAccountCode    string `json:"gainlossaccountcode" bson:"gainlossaccountcode"`
	JournalDocNo           string `json:"journaldocno" bson:"journaldocno"`
	Reason                 string `json:"reason" bson:"reason"`
}

func (AssetDisposal) CollectionName() string { return "asset_disposals" }

type Command struct {
	Resource         string          `json:"resource"` // "assets", "types", "depreciations", "disposals", "processes"
	Action           string          `json:"action"`   // "create", "update", "delete", "calculate", "post-gl", "reverse-gl", "dispose"
	RequestID        string          `json:"requestid"`
	ID               string          `json:"id,omitempty"`
	Version          int64           `json:"version,omitempty"`
	Date             string          `json:"date,omitempty"`
	FiscalYear       string          `json:"fiscalyear,omitempty"`
	Period           int             `json:"period,omitempty"`
	AssetCode        string          `json:"assetcode,omitempty"`
	Reason           string          `json:"reason,omitempty"`
	DocNo            string          `json:"docno,omitempty"`
	Asset            *Asset          `json:"asset,omitempty"`
	AssetType        *AssetType      `json:"assettype,omitempty"`
	Disposal         *AssetDisposal  `json:"disposal,omitempty"`
}

func CleanString(s string) string {
	return strings.TrimSpace(s)
}
