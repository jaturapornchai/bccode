package models

import (
	"encoding/json"
	"time"
)

type ProcessStockMovementStruct struct {
	ItemCode string                             `json:"itemcode"`
	Name     string                             `json:"name"`
	UnitCode string                             `json:"unitcode"`
	UnitName string                             `json:"unitname"`
	Details  []ProcessStockMovementDetailStruct `json:"details"`
}

type ProcessStockMovementDetailStruct struct {
	IsExtra       bool      `json:"isextra"`
	DocDateTime   time.Time `json:"docdatetime"`
	DocNo         string    `json:"docno"`
	ItemCode      string    `json:"itemcode"`
	TransFlag     int       `json:"transflag"`
	UnitCode      string    `json:"unitcode"`
	WhCode        string    `json:"whcode"`
	LocationCode  string    `json:"locationcode"`
	TotalQty      float64   `json:"totalqty"`
	Price         float64   `json:"price"`
	UnitStand     float64   `json:"unitstand"`
	UnitDivide    float64   `json:"unitdivide"`
	AverageCost   float64   `json:"averagecost"`
	CalcAmount    float64   `json:"calcamount"`
	BalanceAmount float64   `json:"balanceamount"`
	BalanceQty    float64   `json:"balanceqty"`
	UnitCost      float64   `json:"unitcost"`
	DocRef        string    `json:"docref"`
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
	ItemCode      string                                `json:"itemcode"`
	ItemName      string                                `json:"itemname"`
	BarcodeList   string                                `json:"barcodelist"`
	UnitCode      string                                `json:"unitcode"`
	UnitName      string                                `json:"unitname"`
	BalanceQty    float64                               `json:"balanceqty"`
	AverageCost   float64                               `json:"averagecost"`
	BalanceAmount float64                               `json:"balanceamount"`
	BalanceWord   string                                `json:"balanceword"`
	IsAutoPacking uint64                                `json:"isautopacking"`
	WareHouses    []ProductBalanceByCodeWareHouseStruct `json:"warehouses"`
}

type ProductBalanceByCodeWareHouseStruct struct {
	WareHouseCode string                               `json:"warehousecode"`
	BalanceQty    float64                              `json:"balanceqty"`
	AverageCost   float64                              `json:"averagecost"`
	BalanceWord   string                               `json:"balanceword"`
	BalanceAmount float64                              `json:"balanceamount"`
	Locations     []ProductBalanceByCodeLocationStruct `json:"locations"`
}

type ProductBalanceByCodeLocationStruct struct {
	LocationCode string  `json:"locationcode"`
	BalanceQty   float64 `json:"balanceqty"`
	BalanceWord  string  `json:"balanceword"`
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
	WareHouseCode string  `json:"whcode"`
	ItemCode      string  `json:"itemcode"`
	ItemName      string  `json:"itemname"`
	BarcodeList   string  `json:"barcodelist"`
	UnitCode      string  `json:"unitcode"`
	UnitName      string  `json:"unitname"`
	BalanceQty    float64 `json:"balanceqty"`
	BalanceWord   string  `json:"balanceword"`
	IsAutoPacking uint64  `json:"isautopacking"`
}

type ProductBalanceByWareHouseAndLocationAndBarcodeStruct struct {
	WareHouseCode string  `json:"whcode"`
	LocationCode  string  `json:"locationcode"`
	ItemCode      string  `json:"itemcode"`
	ItemName      string  `json:"itemname"`
	BarcodeList   string  `json:"barcodelist"`
	UnitCode      string  `json:"unitcode"`
	UnitName      string  `json:"unitname"`
	BalanceQty    float64 `json:"balanceqty"`
	BalanceWord   string  `json:"balanceword"`
	IsAutoPacking uint64  `json:"isautopacking"`
}

type ProductDocRefStruct struct {
	ItemCode   string  `json:"itemcode"`
	Barcode    string  `json:"barcode"`
	UnitCode   string  `json:"unitcode"`
	UnitStand  float64 `json:"unitstand"`
	UnitDivide float64 `json:"unitdivide"`
}

type StockTransactionStruct struct {
	Name  string
	Flags []int
}
