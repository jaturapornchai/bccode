package models

type PurchaseStatusStruct struct {
	DocNo string                          `json:"doc_no"`
	IsClosed bool                            `json:"is_closed"`
	IsComparedSuccess int                             `json:"is_compared_success"` // 0=ยังไม่ได้รับสินค้า, 1=รับครบ, 2=รับไม่ครบ, 3=รับเกิน
	Details []PurchaseStatusDocDetailStruct `json:"details"`
}

type PurchaseStatusDocDetailStruct struct {
	DocNo string                               `json:"doc_no" bson:"doc_no"`
	TransFlag int                                  `json:"trans_flag" bson:"trans_flag"`
	CalcFlag float64                              `json:"calc_flag" bson:"calc_flag"`
	CalcSeq int                                  `json:"calc_seq" bson:"calc_seq"`
	LineNumber int                                  `json:"line_number" bson:"line_number"`
	ItemCode string                               `json:"item_code" bson:"item_code"`
	UnitCode string                               `json:"unit_code" bson:"unit_code"`
	WhCode string                               `json:"wh_code" bson:"wh_code"`
	LocationCode string                               `json:"location_code" bson:"location_code"`
	ToWhCode string                               `json:"to_wh_code" bson:"to_wh_code"`
	ToLocationCode string                               `json:"to_location_code" bson:"to_location_code"`
	TotalQty float64                              `json:"total_qty" bson:"total_qty"`
	Price float64                              `json:"price" bson:"price"`
	PriceExcludeVat float64                              `json:"price_exclude_vat" bson:"price_exclude_vat"`
	ReceivedQty float64                              `json:"receivedqty"`  // จำนวนที่รับมาแล้ว
	PendingQty float64                              `json:"pendingqty"`   // จำนวนที่รอดำเนินการ
	ProductName string                               `json:"product_name"` // ชื่อสินค้า
	UnitName string                               `json:"unit_name"`    // ชื่อหน่วยนับ
	DocRefer []PurchaseStatusDocDetailReferStruct `json:"doc_refer"`    // รายการอ้างอิงการรับสินค้า
}

type PurchaseStatusDocDetailReferStruct struct {
	DocNo string  `json:"doc_no" bson:"doc_no"`
	TransFlag int     `json:"trans_flag" bson:"trans_flag"`
	ReceivedQty float64 `json:"receivedqty"` // จำนวนที่รับ
}
