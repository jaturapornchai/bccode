package models

import (
	"encoding/json"
	"time"
)

type ProcessStockMovementStruct struct {
	ItemCode string                             `json:"item_code" bson:"item_code"`
	Name     string                             `json:"name" bson:"name"`
	UnitCode string                             `json:"unit_code" bson:"unit_code"`
	UnitName string                             `json:"unit_name" bson:"unit_name"`
	Details  []ProcessStockMovementDetailStruct `json:"details" bson:"details"`
}

type ProcessStockMovementDetailStruct struct {
	IsExtra       bool      `json:"is_extra" bson:"is_extra"`
	DocDateTime   time.Time `json:"doc_date_time" bson:"doc_date_time"`
	DocNo         string    `json:"doc_no" bson:"doc_no"`
	ItemCode      string    `json:"item_code" bson:"item_code"`
	TransFlag     int       `json:"trans_flag" bson:"trans_flag"`
	UnitCode      string    `json:"unit_code" bson:"unit_code"`
	WhCode        string    `json:"wh_code" bson:"wh_code"`
	LocationCode  string    `json:"location_code" bson:"location_code"`
	TotalQty      float64   `json:"total_qty" bson:"total_qty"`
	Price         float64   `json:"price" bson:"price"`
	UnitStand     float64   `json:"unit_stand" bson:"unit_stand"`
	UnitDivide    float64   `json:"unit_divide" bson:"unit_divide"`
	AverageCost   float64   `json:"average_cost" bson:"average_cost"`
	CalcAmount    float64   `json:"calc_amount" bson:"calc_amount"`
	BalanceAmount float64   `json:"balance_amount" bson:"balance_amount"`
	BalanceQty    float64   `json:"balance_qty" bson:"balance_qty"`
	UnitCost      float64   `json:"unit_cost" bson:"unit_cost"`
	DocRef        string    `json:"doc_ref" bson:"doc_ref"`
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
	ItemCode      string                                `json:"item_code" bson:"item_code"`
	ItemName      string                                `json:"item_name" bson:"item_name"`
	BarcodeList   string                                `json:"barcode_list" bson:"barcode_list"`
	UnitCode      string                                `json:"unit_code" bson:"unit_code"`
	UnitName      string                                `json:"unit_name" bson:"unit_name"`
	BalanceQty    float64                               `json:"balance_qty" bson:"balance_qty"`
	AverageCost   float64                               `json:"average_cost" bson:"average_cost"`
	BalanceAmount float64                               `json:"balance_amount" bson:"balance_amount"`
	BalanceWord   string                                `json:"balance_word" bson:"balance_word"`
	IsAutoPacking uint64                                `json:"is_auto_packing" bson:"is_auto_packing"`
	WareHouses    []ProductBalanceByCodeWareHouseStruct `json:"warehouses" bson:"warehouses"`
}

type ProductBalanceByCodeWareHouseStruct struct {
	WareHouseCode string                               `json:"warehouse_code" bson:"warehouse_code"`
	BalanceQty    float64                              `json:"balance_qty" bson:"balance_qty"`
	AverageCost   float64                              `json:"average_cost" bson:"average_cost"`
	BalanceWord   string                               `json:"balance_word" bson:"balance_word"`
	BalanceAmount float64                              `json:"balance_amount" bson:"balance_amount"`
	Locations     []ProductBalanceByCodeLocationStruct `json:"locations" bson:"locations"`
}

type ProductBalanceByCodeLocationStruct struct {
	LocationCode string  `json:"location_code" bson:"location_code"`
	BalanceQty   float64 `json:"balance_qty" bson:"balance_qty"`
	BalanceWord  string  `json:"balance_word" bson:"balance_word"`
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
	HoldingCode    string                    `json:"holding_code"`
	CommandID      string                    `json:"command_id"`
	DocNumberList  []string                  `json:"doc_number_list"`
	Condition      string                    `json:"condition"`
	BalanceOnly    string                    `json:"balance_only"`
	MovementOnly   string                    `json:"movement_only"`
	ItemCodeList   json.RawMessage           `json:"item_code_list"`
	BarcodeList    string                    `json:"barcode_list"`
	WarehouseList  []WarehouseListItemStruct `json:"warehouse_list"`
	FromDate       string                    `json:"from_date"`
	FinalDate      string                    `json:"final_date"`
	TimezoneCode   string                    `json:"timezone_code"`
	LanguageCode   string                    `json:"language_code"`
	Guid           string                    `json:"guid,omitempty"`
	PointQty       *int                      `json:"point_qty"`
	PointAmount    *int                      `json:"point_amount"`
	PointCost      *int                      `json:"point_cost"`
	DeleteFirst    *bool                     `json:"delete_first"`
	CreateDatabase *bool                     `json:"create_database"` // true = drop+create ใหม่ (default), false = calc only
}

type WarehouseListItemStruct struct {
	Code       string               `json:"code"`
	Title      string               `json:"title"`
	IsSelected bool                 `json:"is_selected"`
	Locations  []LocationItemStruct `json:"locations,omitempty"`
}

type LocationItemStruct struct {
	Code       string `json:"code"`
	Title      string `json:"title"`
	IsSelected bool   `json:"is_selected"`
}

type ProductBalanceByWareHouseAndBarcodeStruct struct {
	WareHouseCode string  `json:"whcode" bson:"whcode"`
	ItemCode      string  `json:"item_code" bson:"item_code"`
	ItemName      string  `json:"item_name" bson:"item_name"`
	BarcodeList   string  `json:"barcode_list" bson:"barcode_list"`
	UnitCode      string  `json:"unit_code" bson:"unit_code"`
	UnitName      string  `json:"unit_name" bson:"unit_name"`
	BalanceQty    float64 `json:"balance_qty" bson:"balance_qty"`
	BalanceWord   string  `json:"balance_word" bson:"balance_word"`
	IsAutoPacking uint64  `json:"is_auto_packing" bson:"is_auto_packing"`
}

type ProductBalanceByWareHouseAndLocationAndBarcodeStruct struct {
	WareHouseCode string  `json:"whcode" bson:"whcode"`
	LocationCode  string  `json:"location_code" bson:"location_code"`
	ItemCode      string  `json:"item_code" bson:"item_code"`
	ItemName      string  `json:"item_name" bson:"item_name"`
	BarcodeList   string  `json:"barcode_list" bson:"barcode_list"`
	UnitCode      string  `json:"unit_code" bson:"unit_code"`
	UnitName      string  `json:"unit_name" bson:"unit_name"`
	BalanceQty    float64 `json:"balance_qty" bson:"balance_qty"`
	BalanceWord   string  `json:"balance_word" bson:"balance_word"`
	IsAutoPacking uint64  `json:"is_auto_packing" bson:"is_auto_packing"`
}

type ProductDocRefStruct struct {
	ItemCode   string  `json:"itemcode" bson:"itemcode"`
	Barcode    string  `json:"barcode" bson:"barcode"`
	UnitCode   string  `json:"unit_code" bson:"unit_code"`
	UnitStand  float64 `json:"unit_stand" bson:"unit_stand"`
	UnitDivide float64 `json:"unit_divide" bson:"unit_divide"`
}

type StockTransactionStruct struct {
	Name  string
	Flags []int
}

type PayLoadCopyMongoStruct struct {
	SourceHoldingCode string `json:"source_holding_code"`
	TargetHoldingCode string `json:"target_holding_code"`
	SourceEnvironment string `json:"source_environment"`
	TargetEnvironment string `json:"target_environment"`
}
