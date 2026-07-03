package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const productBarcodeBOMCollectionName = "productbarcodeboms"

type BOMProductBarcode struct {
	BarcodeGuidFixed string          `json:"guidfixed" bson:"guidfixed"`
	Level            int             `json:"level" bson:"level"`
	Names            *[]models.NameX `json:"names" bson:"names"`
	ItemUnitCode     string          `json:"itemunitcode" bson:"itemunitcode"`
	ItemUnitNames    *[]models.NameX `json:"itemunitnames" bson:"itemunitnames"`
	Barcode          string          `json:"barcode" bson:"barcode" validate:"required,min=1"`
	RefType          string          `json:"reftype" bson:"reftype,omitempty"`
	Condition        bool            `json:"condition" bson:"condition"`
	DivideValue      float64         `json:"dividevalue" bson:"dividevalue"`
	StandValue       float64         `json:"standvalue" bson:"standvalue"`
	Qty              float64         `json:"qty" bson:"qty"`
	YieldPercent     float64         `json:"yieldpercent" bson:"yieldpercent,omitempty"`
	AverageCost      float64         `json:"averagecost" bson:"averagecost,omitempty"`
	Price            float64         `json:"price" bson:"price,omitempty"`
	MaterialType     int8            `json:"materialtype" bson:"materialtype,omitempty"`
	// SubRecipeOutputQty is only meaningful for reftype=="recipe": the sub-recipe's OWN batch yield
	// (its root Qty), resolved live on read so the frontend can normalize cost per unit. Zero/unset
	// for "product" items.
	SubRecipeOutputQty float64 `json:"subrecipeoutputqty" bson:"subrecipeoutputqty,omitempty"`
	// CostMode, StandardCost, LaborCost, OverheadCost, ScrapPercent, and FinishedGoodBarcode are only
	// meaningful on the recipe ROOT (the top-level ProductBarcodeBOMViewInfo), not on ingredient
	// lines — they live on this shared struct only because BOMProductBarcode is reused for both the
	// root and every line, same precedent as SubRecipeOutputQty above.
	//
	// CostMode is "standard" or "current". Empty defaults to "current" (today's live-calculated
	// rollup) — no migration needed under the Disposable Database Rule, empty-string is just handled
	// as "current" in code.
	//
	// NOTE: these 5 fields deliberately do NOT use `,omitempty` (unlike SubRecipeOutputQty above) —
	// the repository's Update() does a bson.Marshal + `$set` of the whole document (see
	// PersisterMongo.UpdateOne/toDoc), and `omitempty` silently drops zero-valued fields from that
	// $set, making it impossible to ever reset LaborCost/OverheadCost/ScrapPercent/StandardCost back
	// to 0 or CostMode back to "" via a normal save once they were previously nonzero. Found via the
	// "ขายส้มตำ 10 จาน" verification flow (set values, then tried to clear them back to 0 — the old
	// value silently stuck around in MongoDB even though the save reported success).
	CostMode     string  `json:"costmode" bson:"costmode"`
	StandardCost float64 `json:"standardcost" bson:"standardcost"`
	LaborCost    float64 `json:"laborcost" bson:"laborcost"`
	OverheadCost float64 `json:"overheadcost" bson:"overheadcost"`
	ScrapPercent float64 `json:"scrappercent" bson:"scrappercent"`
	// FinishedGoodBarcode links the recipe to the actual sellable product/barcode it produces
	// (future stock-in/COGS integration). Optional/nullable — same no-omitempty reasoning as above,
	// so clearing it back to "" actually persists.
	FinishedGoodBarcode string `json:"finishedgoodbarcode" bson:"finishedgoodbarcode"`
}

type ProductBarcodeBOMView struct {
	BOMProductBarcode `bson:"inline"`
	ImageURI          string                   `json:"imageuri" bson:"imageuri"`
	BOM               *[]ProductBarcodeBOMView `json:"bom" bson:"bom"`
}

type ProductBarcodeBOMVersion struct {
	GuidFixed string                   `json:"guidfixed" bson:"guidfixed"`
	StartDate time.Time                `json:"startdate" bson:"startdate"`
	EndDate   *time.Time               `json:"enddate" bson:"enddate"`
	BOM       *[]ProductBarcodeBOMView `json:"bom" bson:"bom"`
}

type ProductBarcodeBOMSaveRequest struct {
	GuidFixed     string          `json:"guidfixed"`
	Barcode       string          `json:"barcode"`
	Names         *[]models.NameX `json:"names"`
	ItemUnitCode  string          `json:"itemunitcode"`
	ItemUnitNames *[]models.NameX `json:"itemunitnames"`
	Price         float64         `json:"price"`
	// OutputQty is how many recipe units one batch of this BOM produces (e.g. เค้ก 1 สูตร = 10 ชิ้น,
	// น้ำซุป 1 หม้อ = 50 ถ้วย, งานผลิต 1 batch = 100 ชิ้น). Stored on the recipe root's qty field;
	// unit cost = rollup cost / outputqty. Defaults to 1 when omitted or <= 0.
	OutputQty float64                    `json:"outputqty"`
	BOM       []ProductBarcodeBOMView    `json:"bom"`
	BOMs      []ProductBarcodeBOMVersion `json:"boms"`
	// Root-level costing fields — see BOMProductBarcode comment for semantics.
	CostMode            string  `json:"costmode"`
	StandardCost        float64 `json:"standardcost"`
	LaborCost           float64 `json:"laborcost"`
	OverheadCost        float64 `json:"overheadcost"`
	ScrapPercent        float64 `json:"scrappercent"`
	FinishedGoodBarcode string  `json:"finishedgoodbarcode"`
}

func (b *ProductBarcodeBOMView) EmptyOnNil() {

	if b.Names == nil {
		b.Names = &[]models.NameX{}
	}

	if b.ItemUnitNames == nil {
		b.ItemUnitNames = &[]models.NameX{}
	}

	if b.BOM == nil {
		b.BOM = &[]ProductBarcodeBOMView{}
	}
}

type ProductBarcodeBOMViewInfo struct {
	models.DocIdentity    `bson:"inline"`
	ProductBarcodeBOMView `bson:"inline"`
	CheckSum              string                      `json:"checksum" bson:"checksum"`
	IsCurrentUse          bool                        `json:"iscurrentuse" bson:"iscurrentuse"`
	UseInDate             time.Time                   `json:"useindate" bson:"useindate"`
	StartDate             time.Time                   `json:"startdate" bson:"startdate"`
	EndDate               *time.Time                  `json:"enddate" bson:"enddate"`
	BOMs                  *[]ProductBarcodeBOMVersion `json:"boms" bson:"boms,omitempty"`
}

func (ProductBarcodeBOMViewInfo) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewData struct {
	models.HoldingCodeentity  `bson:"inline"`
	ProductBarcodeBOMViewInfo `bson:"inline"`
}

type ProductBarcodeBOMViewDoc struct {
	ID                        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductBarcodeBOMViewData `bson:"inline"`
	models.ActivityDoc        `bson:"inline"`
}

func (ProductBarcodeBOMViewDoc) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewGuid struct {
	models.DocIdentity `bson:"inline"`
}

func (ProductBarcodeBOMViewGuid) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewActivity struct {
	ProductBarcodeBOMViewData `bson:"inline"`
	models.ActivityTime       `bson:"inline"`
}

func (ProductBarcodeBOMViewActivity) CollectionName() string {
	return productBarcodeBOMCollectionName
}

type ProductBarcodeBOMViewDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ProductBarcodeBOMViewDeleteActivity) CollectionName() string {
	return productBarcodeBOMCollectionName
}
