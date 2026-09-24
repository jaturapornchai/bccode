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
	Code string `json:"code"`
	Name string `json:"name"`
}

type Identity struct {
	ID           string    `json:"id"`
	HoldingCode  string    `json:"holdingcode"`
	BusinessCode string    `json:"businesscode"`
	Version      int64     `json:"version"`
	CreatedAt    time.Time `json:"createdat"`
	CreatedBy    string    `json:"createdby"`
	UpdatedAt    time.Time `json:"updatedat"`
	UpdatedBy    string    `json:"updatedby"`
	IsDeleted    bool      `json:"isdeleted"`
}

type AssetType struct {
	Identity
	TypeCode                 string `json:"typecode"`
	Names                    []Name `json:"names"`
	DefaultUsefulLifeYears   int    `json:"defaultusefullifeyears"`
	DefaultDeprecPercent     Amount `json:"defaultdeprecpercent"`
	AssetAccountCode         string `json:"assetaccountcode"`
	AccumDeprecAccountCode   string `json:"accumdeprecaccountcode"`
	DeprecExpenseAccountCode string `json:"deprecexpenseaccountcode"`
	IsActive                 bool   `json:"isactive"`
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
	Identity
	AssetCode                string `json:"assetcode"`
	Names                    []Name `json:"names"`
	AssetTypeCode            string `json:"assettypecode"`
	BranchCode               string `json:"branchcode"`
	DepartmentCode           string `json:"departmentcode"`
	LocationCode             string `json:"locationcode"`
	PurchaseDate             string `json:"purchasedate"` // YYYY-MM-DD
	StartCalcDate            string `json:"startcalcdate"`
	Cost                     Amount `json:"cost"`
	ScrapValue               Amount `json:"scrapvalue"`
	UsefulLifeYears          int    `json:"usefullifeyears"`
	DeprecPercent            Amount `json:"deprecpercent"`
	Method                   string `json:"method"`           // straight_line, sum_of_years, declining
	FirstYearPercent         Amount `json:"firstyearpercent"` // e.g. 40% tax initial allowance
	BeginAccumDeprec         Amount `json:"beginaccumdeprec"`
	AssetAccountCode         string `json:"assetaccountcode"`
	AccumDeprecAccountCode   string `json:"accumdeprecaccountcode"`
	DeprecExpenseAccountCode string `json:"deprecexpenseaccountcode"`
	Status                   string `json:"status"` // active, disposed, written_off
	SerialNumber             string `json:"serialnumber"`
	Brand                    string `json:"brand"`
	Model                    string `json:"model"`
	SupplierCode             string `json:"suppliercode"`
	IsTaxDeductible          bool   `json:"istaxdeductible"`
	// Passenger car / bus with ≤10 seats: tax depreciation only on cost up to the statutory
	// cap (docs/kms/21-thai-tax-form-references.md §13). Set per asset, not by type code,
	// because the law exempts some cars (car-rental stock, R&D prototypes).
	PassengerCarTaxCap bool   `json:"passengercartaxcap"`
	Notes              string `json:"notes"`
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
	Identity
	AssetCode    string     `json:"assetcode"`
	FiscalYear   string     `json:"fiscalyear"`
	Period       int        `json:"period"` // 1-12
	StartDate    string     `json:"startdate"`
	StopDate     string     `json:"stopdate"`
	Days         int        `json:"days"`
	PeriodDeprec Amount     `json:"perioddeprec"`
	AccumDeprec  Amount     `json:"accumdeprec"`
	NetBookValue Amount     `json:"netbookvalue"`
	IsPosted     bool       `json:"isposted"`
	JournalDocNo string     `json:"journaldocno"`
	PostedAt     *time.Time `json:"postedat,omitempty"`
}

func (DepreciationScheduleItem) CollectionName() string { return "asset_depreciations" }

type AssetDisposal struct {
	Identity
	DocNo                  string `json:"docno"`
	AssetCode              string `json:"assetcode"`
	DisposalDate           string `json:"disposaldate"`
	DisposalType           string `json:"disposaltype"` // sale, write_off, scrap
	SalePrice              Amount `json:"saleprice"`
	VatAmount              Amount `json:"vatamount"`
	AccumDeprecAtDisposal  Amount `json:"accumdeprecatdisposal"`
	NetBookValueAtDisposal Amount `json:"netbookvalueatdisposal"`
	GainLoss               Amount `json:"gainloss"`              // SalePrice - NetBookValue
	SettlementAccountCode  string `json:"settlementaccountcode"` // Cash or AR
	GainLossAccountCode    string `json:"gainlossaccountcode"`
	VatAccountCode         string `json:"vataccountcode"` // output VAT; required when vatamount > 0
	JournalDocNo           string `json:"journaldocno"`   // optional input; blank = <general book>-DISP-<assetcode>
	Reason                 string `json:"reason"`
}

func (AssetDisposal) CollectionName() string { return "asset_disposals" }

type Command struct {
	Resource   string         `json:"resource"` // "assets", "types", "depreciations", "disposals", "processes"
	Action     string         `json:"action"`   // "create", "update", "delete", "calculate", "post-gl", "reverse-gl", "dispose"
	RequestID  string         `json:"requestid"`
	ID         string         `json:"id,omitempty"`
	Version    int64          `json:"version,omitempty"`
	Date       string         `json:"date,omitempty"`
	FiscalYear string         `json:"fiscalyear,omitempty"`
	Period     int            `json:"period,omitempty"`
	AssetCode  string         `json:"assetcode,omitempty"`
	Reason     string         `json:"reason,omitempty"`
	DocNo      string         `json:"docno,omitempty"`
	BranchCode string         `json:"branchcode,omitempty"` // post-gl voucher branch; blank = session branch
	Asset      *Asset         `json:"asset,omitempty"`
	AssetType  *AssetType     `json:"assettype,omitempty"`
	Disposal   *AssetDisposal `json:"disposal,omitempty"`
}

func CleanString(s string) string {
	return strings.TrimSpace(s)
}
