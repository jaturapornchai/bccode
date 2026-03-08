package models

import (
	"encoding/json"
	"time"
)

type ProcessStockMovementStruct struct {
	ItemCode string                             `json:"itemCode" bson:"itemCode"`
	Name     string                             `json:"name" bson:"name"`
	UnitCode string                             `json:"unitCode" bson:"unitCode"`
	UnitName string                             `json:"unitName" bson:"unitName"`
	Details  []ProcessStockMovementDetailStruct `json:"details" bson:"details"`
}

type ProcessStockMovementDetailStruct struct {
	IsExtra       bool      `json:"isExtra" bson:"isExtra"`
	DocDateTime   time.Time `json:"docDateTime" bson:"docDateTime"`
	DocNo         string    `json:"docNo" bson:"docNo"`
	ItemCode      string    `json:"itemCode" bson:"itemCode"`
	TransFlag     int       `json:"transFlag" bson:"transFlag"`
	UnitCode      string    `json:"unitCode" bson:"unitCode"`
	WhCode        string    `json:"whCode" bson:"whCode"`
	LocationCode  string    `json:"locationCode" bson:"locationCode"`
	TotalQty      float64   `json:"totalQty" bson:"totalQty"`
	Price         float64   `json:"price" bson:"price"`
	UnitStand     float64   `json:"unitStand" bson:"unitStand"`
	UnitDivide    float64   `json:"unitDivide" bson:"unitDivide"`
	AverageCost   float64   `json:"averageCost" bson:"averageCost"`
	CalcAmount    float64   `json:"calcAmount" bson:"calcAmount"`
	BalanceAmount float64   `json:"balanceAmount" bson:"balanceAmount"`
	BalanceQty    float64   `json:"balanceQty" bson:"balanceQty"`
	UnitCost      float64   `json:"unitCost" bson:"unitCost"`
	DocRef        string    `json:"docRef" bson:"docRef"`
}

type ProcessStockCostDetailStruct struct {
	ShopID          string
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
	ShopID         string                    `json:"shop_id"`
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
	IsSelected bool                 `json:"isSelected"`
	Locations  []LocationItemStruct `json:"locations,omitempty"`
}

type LocationItemStruct struct {
	Code       string `json:"code"`
	Title      string `json:"title"`
	IsSelected bool   `json:"isSelected"`
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
	SourceShopID string `json:"source_shop_id"`
	TargetShopID string `json:"target_shop_id"`
}
