package models

import (
	"encoding/json"
	"time"
)

type ProcessStockMovementStruct struct {
	ItemCode string                             `json:"itemcode" bson:"itemcode"`
	Name     string                             `json:"name" bson:"name"`
	UnitCode string                             `json:"unitcode" bson:"unitcode"`
	UnitName string                             `json:"unitname" bson:"unitname"`
	Details  []ProcessStockMovementDetailStruct `json:"details" bson:"details"`
}

type ProcessStockMovementDetailStruct struct {
	IsExtra       bool      `json:"isextra" bson:"isextra"`
	DocDateTime   time.Time `json:"docdatetime" bson:"docdatetime"`
	DocNo         string    `json:"docno" bson:"docno"`
	ItemCode      string    `json:"itemcode" bson:"itemcode"`
	TransFlag     int       `json:"transflag" bson:"transflag"`
	UnitCode      string    `json:"unitcode" bson:"unitcode"`
	WhCode        string    `json:"whcode" bson:"whcode"`
	LocationCode  string    `json:"locationcode" bson:"locationcode"`
	TotalQty      float64   `json:"totalqty" bson:"totalqty"`
	Price         float64   `json:"price" bson:"price"`
	UnitStand     float64   `json:"unitstand" bson:"unitstand"`
	UnitDivide    float64   `json:"unitdivide" bson:"unitdivide"`
	AverageCost   float64   `json:"averagecost" bson:"averagecost"`
	CalcAmount    float64   `json:"calcamount" bson:"calcamount"`
	BalanceAmount float64   `json:"balanceamount" bson:"balanceamount"`
	BalanceQty    float64   `json:"balanceqty" bson:"balanceqty"`
	UnitCost      float64   `json:"unitcost" bson:"unitcost"`
	DocRef        string    `json:"docref" bson:"docref"`
}

type ProcessStockCostDetailStruct struct {
	HoldingCode     string
	DocDateTime     time.Time
	DocNo           string
	LineNumber      int
	TransFlag       int
	ItemCode        string
	Barcode         string
	BarcodeMain     string
	UnitCode        string
	WhCode          string
	LocationCode    string
	Qty             float64
	Price           float64
	PriceExcludeVat float64
	UnitStand       float64
	UnitDivide      float64
	UnitCost        float64
	AverageCost     float64
	SumAmount       float64
	CalcAmount      float64
	BalanceAmount   float64
	BalanceQty      float64
	Guid            string
	DocRef          string
}

type ProcessStockLotStruct struct {
	DocDateTime   time.Time
	LotNumber     string
	DocNo         string
	TransFlag     int
	ItemCode      string
	UnitCode      string
	WhCode        string
	LocationCode  string
	Qty           float64
	Price         float64
	UnitStand     float64
	UnitDivide    float64
	Cost          float64
	BalanceAmount float64
	BalanceQty    float64
	GuidRef       string
}

// กำหนด struct เพื่อเก็บข้อมูลแต่ละ row
type ProductBalanceStruct struct {
	ItemCode      string
	ItemName      string
	WhCode        string
	LocationCode  string
	UnitCode      string
	UnitName      string
	BalanceQty    float64
	BalanceAmount float64
	BalanceWord   string
	IsAutoPacking uint64
}

// auto packing
type ProductBarcodePackingStruct struct {
	UnitName             string
	BarcodeRefUnitStand  float64
	BarcodeRefUnitDivide float64
}

type ProductBalanceByCodeStruct struct {
	ItemCode      string                                `json:"itemcode" bson:"itemcode"`
	ItemName      string                                `json:"itemname" bson:"itemname"`
	BarcodeList   string                                `json:"barcodelist" bson:"barcodelist"`
	UnitCode      string                                `json:"unitcode" bson:"unitcode"`
	UnitName      string                                `json:"unitname" bson:"unitname"`
	BalanceQty    float64                               `json:"balanceqty" bson:"balanceqty"`
	AverageCost   float64                               `json:"averagecost" bson:"averagecost"`
	BalanceAmount float64                               `json:"balanceamount" bson:"balanceamount"`
	BalanceWord   string                                `json:"balanceword" bson:"balanceword"`
	IsAutoPacking uint64                                `json:"isautopacking" bson:"isautopacking"`
	WareHouses    []ProductBalanceByCodeWareHouseStruct `json:"warehouses" bson:"warehouses"`
}

type ProductBalanceByCodeWareHouseStruct struct {
	WareHouseCode string                               `json:"warehousecode" bson:"warehousecode"`
	BalanceQty    float64                              `json:"balanceqty" bson:"balanceqty"`
	AverageCost   float64                              `json:"averagecost" bson:"averagecost"`
	BalanceWord   string                               `json:"balanceword" bson:"balanceword"`
	BalanceAmount float64                              `json:"balanceamount" bson:"balanceamount"`
	Locations     []ProductBalanceByCodeLocationStruct `json:"locations" bson:"locations"`
}

type ProductBalanceByCodeLocationStruct struct {
	LocationCode string  `json:"locationcode" bson:"locationcode"`
	BalanceQty   float64 `json:"balanceqty" bson:"balanceqty"`
	BalanceWord  string  `json:"balanceword" bson:"balanceword"`
}

type ProductBalanceByCodeWareHouseGetStruct struct {
	ItemCode      string
	WareHouseCode string
	BalanceQty    float64
	AverageCost   float64
	BalanceAmount float64
}

type ProductBalanceByCodeLocationGetStruct struct {
	ItemCode      string
	WareHouseCode string
	LocationCode  string
	BalanceQty    float64
}

type PayLoadCommandStruct struct {
	HoldingCode    string                    `json:"holdingcode"`
	CommandID      string                    `json:"commandid"`
	DocNumberList  []string                  `json:"docnumberlist"`
	Condition      string                    `json:"condition"`
	BalanceOnly    string                    `json:"balanceonly"`
	MovementOnly   string                    `json:"movementonly"`
	ItemCodeList   json.RawMessage           `json:"itemcodelist"`
	BarcodeList    string                    `json:"barcodelist"`
	WarehouseList  []WarehouseListItemStruct `json:"warehouselist"`
	FromDate       string                    `json:"fromdate"`
	FinalDate      string                    `json:"finaldate"`
	TimezoneCode   string                    `json:"timezonecode"`
	LanguageCode   string                    `json:"languagecode"`
	Guid           string                    `json:"guid,omitempty"`
	PointQty       *int                      `json:"pointqty"`
	PointAmount    *int                      `json:"pointamount"`
	PointCost      *int                      `json:"pointcost"`
	DeleteFirst    *bool                     `json:"deletefirst"`
	CreateDatabase *bool                     `json:"createdatabase"` // true = drop+create ใหม่ (default), false = calc only
}

type WarehouseListItemStruct struct {
	Code       string               `json:"code"`
	Title      string               `json:"title"`
	IsSelected bool                 `json:"isselected"`
	Locations  []LocationItemStruct `json:"locations,omitempty"`
}

type LocationItemStruct struct {
	Code       string `json:"code"`
	Title      string `json:"title"`
	IsSelected bool   `json:"isselected"`
}

type ProductBalanceByWareHouseAndBarcodeStruct struct {
	WareHouseCode string  `json:"whcode" bson:"whcode"`
	ItemCode      string  `json:"itemcode" bson:"itemcode"`
	ItemName      string  `json:"itemname" bson:"itemname"`
	BarcodeList   string  `json:"barcodelist" bson:"barcodelist"`
	UnitCode      string  `json:"unitcode" bson:"unitcode"`
	UnitName      string  `json:"unitname" bson:"unitname"`
	BalanceQty    float64 `json:"balanceqty" bson:"balanceqty"`
	BalanceWord   string  `json:"balanceword" bson:"balanceword"`
	IsAutoPacking uint64  `json:"isautopacking" bson:"isautopacking"`
}

type ProductBalanceByWareHouseAndLocationAndBarcodeStruct struct {
	WareHouseCode string  `json:"whcode" bson:"whcode"`
	LocationCode  string  `json:"locationcode" bson:"locationcode"`
	ItemCode      string  `json:"itemcode" bson:"itemcode"`
	ItemName      string  `json:"itemname" bson:"itemname"`
	BarcodeList   string  `json:"barcodelist" bson:"barcodelist"`
	UnitCode      string  `json:"unitcode" bson:"unitcode"`
	UnitName      string  `json:"unitname" bson:"unitname"`
	BalanceQty    float64 `json:"balanceqty" bson:"balanceqty"`
	BalanceWord   string  `json:"balanceword" bson:"balanceword"`
	IsAutoPacking uint64  `json:"isautopacking" bson:"isautopacking"`
}

type ProductDocRefStruct struct {
	ItemCode   string  `json:"itemcode" bson:"itemcode"`
	Barcode    string  `json:"barcode" bson:"barcode"`
	UnitCode   string  `json:"unitcode" bson:"unitcode"`
	UnitStand  float64 `json:"unitstand" bson:"unitstand"`
	UnitDivide float64 `json:"unitdivide" bson:"unitdivide"`
}

type StockTransactionStruct struct {
	Name  string
	Flags []int
}

type PayLoadCopyMongoStruct struct {
	SourceHoldingCode string `json:"sourceholdingcode"`
	TargetHoldingCode string `json:"targetholdingcode"`
	SourceEnvironment string `json:"sourceenvironment"`
	TargetEnvironment string `json:"targetenvironment"`
}
