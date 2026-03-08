package models

type CountStatus struct {
	Status int8 `json:"status"`
	Count  int  `json:"count"`
}
type DocumentImageTaskChangeMessage struct {
	ShopID           string        `json:"shopid"`
	TaskGUID         string        `json:"taskguid"`
	Count            int           `json:"count"`
	BillCount        float64       `json:"billcount"`
	ReferenceCount   float64       `json:"referencecount"`
	ReferenceBalance float64       `json:"referencebalance"`
	CountStatus      []CountStatus `json:"countstatus"`
	// Event    TaskChangeEvent `json:"event"`
}

type DocumentImageTaskRejectMessage struct {
	ShopID   string `json:"shopid"`
	TaskGUID string `json:"taskguid"`
	Count    int    `json:"count"`
	// Event    TaskRejectEvent `json:"event"`
}

type TaskChangeEvent int8

const (
	TaskChangePlus TaskChangeEvent = iota
	TaskChangeMinus
)

type TaskRejectEvent int8

const (
	TaskRejectPlus TaskRejectEvent = iota
	TaskRejectMinus
)
