package generalledger

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

type Change struct {
	Kind    string `json:"kind" bson:"kind"`
	ID      string `json:"id" bson:"id"`
	Code    string `json:"code" bson:"code"`
	Payload string `json:"payload" bson:"payload"`
}

// One command, its audit and its projection intent are committed together.
// Payloads are immutable decimal-string JSON; only delivery metadata may change.
type Event struct {
	ID           string    `json:"id" bson:"_id"`
	HoldingCode  string    `json:"holdingcode" bson:"holdingcode"`
	BusinessCode string    `json:"businesscode" bson:"businesscode"`
	Sequence     int64     `json:"sequence" bson:"sequence"`
	RequestID    string    `json:"requestid" bson:"requestid"`
	RequestHash  string    `json:"requesthash" bson:"requesthash"`
	Action       string    `json:"action" bson:"action"`
	Actor        string    `json:"actor" bson:"actor"`
	Reason       string    `json:"reason" bson:"reason"`
	OccurredAt   time.Time `json:"occurredat" bson:"occurredat"`
	Changes      []Change  `json:"changes" bson:"changes"`
	Delivered    bool      `json:"delivered" bson:"delivered"`
}

func (Event) CollectionName() string { return "gl_events" }

type Page struct {
	Sequence int64             `json:"sequence"`
	Items    []json.RawMessage `json:"items"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

type ReportQuery struct {
	From           string
	To             string
	FiscalYear     string
	AccountCode    string
	BranchCode     string
	DepartmentCode string
	ProjectCode    string
	BookCode       string
	BudgetCode     string
	Page           int
	Limit          int
}

type ReportColumn struct {
	Key    string `json:"key"`
	Label  string `json:"label"`
	Amount bool   `json:"amount,omitempty"`
}
type Report struct {
	Sequence  int64               `json:"sequence"`
	Columns   []ReportColumn      `json:"columns"`
	Rows      []map[string]string `json:"rows"`
	Totals    map[string]string   `json:"totals"`
	TotalRows int64               `json:"totalrows"`
	Warnings  []string            `json:"warnings"`
	AsOf      string              `json:"asof"`
}

type Projection interface {
	Project(context.Context, Event) error
	Version(context.Context, Scope) (int64, error)
	List(context.Context, Scope, string, string, int, int, ListFilter) (Page, error)
	Get(context.Context, Scope, string, string) (json.RawMessage, error)
	Report(context.Context, Scope, string, ReportQuery) (Report, error)
}

var ErrProjectionPending = errors.New("บันทึกแล้ว กำลังเตรียมข้อมูลรายงาน กรุณารอสักครู่แล้วโหลดใหม่")

var ErrNotFound = errors.New("ไม่พบข้อมูลบัญชีที่ต้องการ")

type ListFilter struct {
	BookCode string
	Kind     string
	Status   string
}
