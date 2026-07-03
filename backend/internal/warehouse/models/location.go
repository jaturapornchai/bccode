package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const warehouseLocationCollectionName = "warehouselocation"

// LocationType enumerates the allowed warehouselocation.locationtype values.
const (
	LocationTypeStorage   = "storage"
	LocationTypePicking   = "picking"
	LocationTypeReceiving = "receiving"
	LocationTypeQC        = "qc"
	LocationTypeWIP       = "wip"
	LocationTypeDamaged   = "damaged"
	LocationTypeTransit   = "transit"
)

// WarehouseLocation is master data (no stock/qty) — see scopeofwork/warehouse.md ("ที่เก็บสินค้า").
// It lives in its own collection, referencing the parent Warehouse by WarehouseGuid.
type WarehouseLocation struct {
	WarehouseGuid         string          `json:"warehouseguid" bson:"warehouseguid" validate:"required"`
	Code                  string          `json:"code" bson:"code" validate:"required"`
	Names                 *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,dive"`
	LocationType          string          `json:"locationtype" bson:"locationtype" validate:"required,oneof=storage picking receiving qc wip damaged transit"`
	AllowedProductClasses []string        `json:"allowedproductclasses" bson:"allowedproductclasses"`
	HazardClasses         []string        `json:"hazardclasses" bson:"hazardclasses"`
	AllowPutaway          bool            `json:"allowputaway" bson:"allowputaway"`
	AllowPick             bool            `json:"allowpick" bson:"allowpick"`
	BlockedIn             bool            `json:"blockedin" bson:"blockedin"`
	BlockedOut            bool            `json:"blockedout" bson:"blockedout"`
	SortCode              string          `json:"sortcode" bson:"sortcode"`
	Status                string          `json:"status" bson:"status"`
	// CompanyGuids restricts which companies may use this location. Empty = inherit all companies
	// the parent Warehouse allows. Must always be a subset of the parent Warehouse's CompanyGuids
	// when the warehouse itself is restricted — enforced by validateLocationCompanyScope.
	CompanyGuids []string `json:"companyguids" bson:"companyguids,omitempty"`
}

type WarehouseLocationInfo struct {
	models.DocIdentity `bson:"inline"`
	WarehouseLocation  `bson:"inline"`
}

func (WarehouseLocationInfo) CollectionName() string {
	return warehouseLocationCollectionName
}

type WarehouseLocationData struct {
	models.HoldingCodeentity `bson:"inline"`
	WarehouseLocationInfo    `bson:"inline"`
}

type WarehouseLocationDoc struct {
	ID                    primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WarehouseLocationData `bson:"inline"`
	models.ActivityDoc    `bson:"inline"`
}

func (WarehouseLocationDoc) CollectionName() string {
	return warehouseLocationCollectionName
}

type WarehouseLocationItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (WarehouseLocationItemGuid) CollectionName() string {
	return warehouseLocationCollectionName
}

type WarehouseLocationActivity struct {
	WarehouseLocationData `bson:"inline"`
	models.ActivityTime   `bson:"inline"`
}

func (WarehouseLocationActivity) CollectionName() string {
	return warehouseLocationCollectionName
}

type WarehouseLocationDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WarehouseLocationDeleteActivity) CollectionName() string {
	return warehouseLocationCollectionName
}
