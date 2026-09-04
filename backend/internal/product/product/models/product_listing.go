package models

import (
	"encoding/json"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsontype"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ListingPrice ราคาแบบทศนิยมเที่ยงตรง (Decimal128) รับ JSON ได้ทั้งข้อความ "199.00" และตัวเลข 199.00
// ตอบกลับเป็นข้อความเสมอเพื่อไม่ให้ทศนิยมเพี้ยน และเก็บใน MongoDB เป็น Decimal128
type ListingPrice struct {
	primitive.Decimal128
}

func (p *ListingPrice) UnmarshalJSON(data []byte) error {
	raw := strings.TrimSpace(string(data))
	if raw == "" || raw == "null" {
		*p = ListingPrice{}
		return nil
	}
	if strings.HasPrefix(raw, "\"") {
		var text string
		if err := json.Unmarshal(data, &text); err != nil {
			return err
		}
		raw = strings.TrimSpace(text)
	}
	d, err := primitive.ParseDecimal128(raw)
	if err != nil {
		return err
	}
	p.Decimal128 = d
	return nil
}

func (p ListingPrice) MarshalJSON() ([]byte, error) {
	return json.Marshal(p.Decimal128.String())
}

func (p ListingPrice) MarshalBSONValue() (bsontype.Type, []byte, error) {
	return bson.MarshalValue(p.Decimal128)
}

func (p *ListingPrice) UnmarshalBSONValue(t bsontype.Type, data []byte) error {
	return bson.UnmarshalValue(t, data, &p.Decimal128)
}

// สภาพสินค้าในชั้นลงขาย (ช่องทางขายออนไลน์)
const (
	ListingConditionNew  = "NEW"
	ListingConditionUsed = "USED"
)

// ProductListing ข้อมูลชั้นลงขายของสินค้า — แยกจากข้อมูลบัญชี แก้ได้ผ่าน API v2 เท่านั้น
type ProductListing struct {
	Title         string                       `json:"title" bson:"title"`             // ชื่อสินค้าสำหรับลงขาย
	Description   string                       `json:"description" bson:"description"` // รายละเอียดสำหรับลงขาย
	Condition     string                       `json:"condition" bson:"condition"`     // NEW | USED
	Preorder      *ProductListingPreorder      `json:"preorder,omitempty" bson:"preorder,omitempty"`
	PurchaseLimit *ProductListingPurchaseLimit `json:"purchaselimit,omitempty" bson:"purchaselimit,omitempty"`
	Wholesale     []ProductListingWholesale    `json:"wholesale" bson:"wholesale"`
	SizeChart     *ProductListingSizeChart     `json:"sizechart,omitempty" bson:"sizechart,omitempty"`
	Tiers         []ProductListingTier         `json:"tiers" bson:"tiers"`
}

// ProductListingPreorder การสั่งจองล่วงหน้า
type ProductListingPreorder struct {
	IsPreorder bool `json:"ispreorder" bson:"ispreorder"`
	DaysToShip int  `json:"daystoship" bson:"daystoship"` // จำนวนวันเตรียมจัดส่ง
}

// ProductListingPurchaseLimit จำกัดจำนวนซื้อต่อคำสั่งซื้อ
type ProductListingPurchaseLimit struct {
	Min int `json:"min" bson:"min"`
	Max int `json:"max" bson:"max"`
}

// ProductListingWholesale ราคาขายส่งตามช่วงจำนวน — unitprice เป็น Decimal128 (JSON เป็นข้อความ เช่น "179.00")
type ProductListingWholesale struct {
	MinCount  int          `json:"mincount" bson:"mincount"`
	MaxCount  int          `json:"maxcount" bson:"maxcount"`
	UnitPrice ListingPrice `json:"unitprice" bson:"unitprice"`
}

// ProductListingSizeChart ตารางไซส์
type ProductListingSizeChart struct {
	URI      string `json:"uri" bson:"uri"`
	URIThumb string `json:"urithumb" bson:"urithumb"`
}

// ProductListingTier ชั้นตัวเลือกสินค้า (เช่น สี, ขนาด)
type ProductListingTier struct {
	XOrder  int                        `json:"xorder" bson:"xorder"`
	Name    string                     `json:"name" bson:"name"`
	Options []ProductListingTierOption `json:"options" bson:"options"`
}

// ProductListingTierOption ตัวเลือกในแต่ละชั้น
type ProductListingTierOption struct {
	XOrder        int    `json:"xorder" bson:"xorder"`
	Name          string `json:"name" bson:"name"`
	ImageURI      string `json:"imageuri" bson:"imageuri"`
	ImageURIThumb string `json:"imageurithumb" bson:"imageurithumb"`
}
