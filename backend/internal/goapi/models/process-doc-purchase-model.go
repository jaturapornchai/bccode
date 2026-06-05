package models

type PurchaseStatusStruct struct {
	DocNo             string                          `json:"docno"`
	IsClosed          bool                            `json:"isclosed"`
	IsComparedSuccess int                             `json:"iscomparedsuccess"` // 0=ยังไม่ได้รับสินค้า, 1=รับครบ, 2=รับไม่ครบ, 3=รับเกิน
	Details           []PurchaseStatusDocDetailStruct `json:"details"`
}

type PurchaseStatusDocDetailStruct struct {
	DocNo           string                               `json:"docno" bson:"docno"`
	TransFlag       int                                  `json:"transflag" bson:"transflag"`
	CalcFlag        float64                              `json:"calcflag" bson:"calcflag"`
	CalcSeq         int                                  `json:"calcseq" bson:"calcseq"`
	LineNumber      int                                  `json:"linenumber" bson:"linenumber"`
	ItemCode        string                               `json:"itemcode" bson:"itemcode"`
	UnitCode        string                               `json:"unitcode" bson:"unitcode"`
	WhCode          string                               `json:"whcode" bson:"whcode"`
	LocationCode    string                               `json:"locationcode" bson:"locationcode"`
	ToWhCode        string                               `json:"towhcode" bson:"towhcode"`
	ToLocationCode  string                               `json:"tolocationcode" bson:"tolocationcode"`
	TotalQty        float64                              `json:"totalqty" bson:"totalqty"`
	Price           float64                              `json:"price" bson:"price"`
	PriceExcludeVat float64                              `json:"priceexcludevat" bson:"priceexcludevat"`
	ReceivedQty     float64                              `json:"receivedqty"` // จำนวนที่รับมาแล้ว
	PendingQty      float64                              `json:"pendingqty"`  // จำนวนที่รอดำเนินการ
	ProductName     string                               `json:"productname"` // ชื่อสินค้า
	UnitName        string                               `json:"unitname"`    // ชื่อหน่วยนับ
	DocRefer        []PurchaseStatusDocDetailReferStruct `json:"docrefer"`    // รายการอ้างอิงการรับสินค้า
}

type PurchaseStatusDocDetailReferStruct struct {
	DocNo       string  `json:"docno" bson:"docno"`
	TransFlag   int     `json:"transflag" bson:"transflag"`
	ReceivedQty float64 `json:"receivedqty"` // จำนวนที่รับ
}
