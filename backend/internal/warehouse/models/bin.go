package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const warehouseBinCollectionName = "warehousebin"

// BinType enumerates the allowed warehousebin.bintype values.
const (
	BinTypeStorage = "storage"
	BinTypePicking = "picking"
	BinTypeStaging = "staging"
	BinTypeWIP     = "wip"
	BinTypeQC      = "qc"
	BinTypeDamaged = "damaged"
)

// BinFixedItem is a product reference hint for putaway suggestions only — never a balance/qty.
// See s/warehouse.md ("ที่วางสินค้า") and the Product Stock/Cost rule: bins must never
// carry accounting qty.
type BinFixedItem struct {
	GuidFixed string          `json:"guidfixed" bson:"guidfixed" validate:"required"`
	Barcode   string          `json:"barcode" bson:"barcode"`
	Unitcode  string          `json:"unitcode" bson:"unitcode"`
	Names     *[]models.NameX `json:"names" bson:"names"`
}

// WarehouseBin is master data (no stock/qty) — "ที่วางสินค้า" under a WarehouseLocation.
// widthmm/lengthmm/heightmm/maxweightgram/maxvolumecm3/mintemperaturedeci/maxtemperaturedeci are
// integer physical-spec fields (ERP Accounting Decimal Iron Rule applied to physical dimensions —
// naming encodes the scale so no float rounding ever enters these fields).
type WarehouseBin struct {
	WarehouseGuid         string         `json:"warehouseguid" bson:"warehouseguid" validate:"required"`
	LocationGuid          string         `json:"locationguid" bson:"locationguid" validate:"required"`
	Code                  string         `json:"code" bson:"code" validate:"required"`
	Name                  string         `json:"name" bson:"name" validate:"required"`
	Barcode               string         `json:"barcode" bson:"barcode"`
	BinType               string         `json:"bintype" bson:"bintype" validate:"required,oneof=storage picking staging wip qc damaged"`
	AisleCode             string         `json:"aislecode" bson:"aislecode"`
	RackCode              string         `json:"rackcode" bson:"rackcode"`
	LevelCode             string         `json:"levelcode" bson:"levelcode"`
	PositionCode          string         `json:"positioncode" bson:"positioncode"`
	WidthMm               int64          `json:"widthmm" bson:"widthmm"`
	LengthMm              int64          `json:"lengthmm" bson:"lengthmm"`
	HeightMm              int64          `json:"heightmm" bson:"heightmm"`
	MaxWeightGram         int64          `json:"maxweightgram" bson:"maxweightgram"`
	MaxVolumeCm3          int64          `json:"maxvolumecm3" bson:"maxvolumecm3"`
	MinTemperatureDeci    int64          `json:"mintemperaturedeci" bson:"mintemperaturedeci"`
	MaxTemperatureDeci    int64          `json:"maxtemperaturedeci" bson:"maxtemperaturedeci"`
	AllowedProductClasses []string       `json:"allowedproductclasses" bson:"allowedproductclasses"`
	HazardClasses         []string       `json:"hazardclasses" bson:"hazardclasses"`
	FixedItems            []BinFixedItem `json:"fixeditems" bson:"fixeditems"`
	AllowPutaway          bool           `json:"allowputaway" bson:"allowputaway"`
	AllowPick             bool           `json:"allowpick" bson:"allowpick"`
	BlockedIn             bool           `json:"blockedin" bson:"blockedin"`
	BlockedOut            bool           `json:"blockedout" bson:"blockedout"`
	SortCode              string         `json:"sortcode" bson:"sortcode"`
	Status                string         `json:"status" bson:"status"`
}

type WarehouseBinInfo struct {
	models.DocIdentity `bson:"inline"`
	WarehouseBin       `bson:"inline"`
}

func (WarehouseBinInfo) CollectionName() string {
	return warehouseBinCollectionName
}

type WarehouseBinData struct {
	models.HoldingCodeentity `bson:"inline"`
	WarehouseBinInfo         `bson:"inline"`
}

type WarehouseBinDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WarehouseBinData   `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (WarehouseBinDoc) CollectionName() string {
	return warehouseBinCollectionName
}

type WarehouseBinItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (WarehouseBinItemGuid) CollectionName() string {
	return warehouseBinCollectionName
}

type WarehouseBinActivity struct {
	WarehouseBinData    `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WarehouseBinActivity) CollectionName() string {
	return warehouseBinCollectionName
}

type WarehouseBinDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WarehouseBinDeleteActivity) CollectionName() string {
	return warehouseBinCollectionName
}
