package models

import (
	"math"
	"time"
)

type StockBalanceImportMeta struct {
	TotalItem   uint64  `json:"totalitem" ch:"totalitem"`
	TotalAmount float64 `json:"total_amount" ch:"total_amount"`
}

type StockBalanceImportRaw struct {
	Barcode       string  `json:"barcode" ch:"barcode"`
	Name          string  `json:"name" ch:"name"`
	UnitCode      string  `json:"unitcode" ch:"unitcode"`
	WarehouseCode string  `json:"warehousecode" ch:"warehousecode"`
	ShelfCode     string  `json:"shelfcode" ch:"shelfcode"`
	Qty           float64 `json:"qty" ch:"qty"`
	Price         float64 `json:"price" ch:"price"`
	SumAmount     float64 `json:"sum_amount" ch:"sum_amount"`
	IsNotExist    bool    `json:"isnotexist" ch:"isnotexist"`
}

type StockBalanceImport struct {
	TaskID    string  `json:"taskid" ch:"taskid"`
	RowNumber float64 `json:"rownumber" ch:"rownumber"`
	StockBalanceImportRaw
}

type StockBalanceImportInfo struct {
	GUIDFixed   string `json:"guid_fixed" ch:"guid_fixed"`
	HoldingCode string `json:"holding_code" ch:"holding_code"`
	StockBalanceImport
}

type StockBalanceImportDoc struct {
	StockBalanceImportInfo
	CreatedAt time.Time `json:"created_at" ch:"created_at"`
	CreatedBy string    `json:"createdby" ch:"createdby"`
}

func (StockBalanceImportDoc) TableName() string {
	return "stockbalanceimport"
}

type TaskStatus int8

const (
	TaskStatusPending TaskStatus = iota
	TaskStatusProcessing
	TaskStatusDone
	TaskStatusError
	TaskStatusSaveSucceded
	TaskStatusSaveFailed
	TaskStatusNotFound
)

type PaginationData struct {
	Total     int64 `json:"total"`
	Page      int64 `json:"page"`
	PerPage   int64 `json:"per_page"`
	Prev      int64 `json:"prev"`
	Next      int64 `json:"next"`
	TotalPage int64 `json:"total_page"`
}

func (p *PaginationData) Build() {
	totalPage := math.Ceil(float64(p.Total) / float64(p.PerPage))
	p.TotalPage = int64(totalPage)

	if p.Page == 0 {
		p.Page = 1
	}

	if p.Page > 1 {
		p.Prev = p.Page - 1
	}

	if p.Page < p.TotalPage {
		p.Next = p.Page + 1
	}
}
