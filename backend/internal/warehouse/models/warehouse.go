package models

import (
	"smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const warehouseCollectionName = "warehouse"

// Warehouse is master data only (per scopeofwork/warehouse.md — no embedded location/shelf/stock).
// Locations live in the separate "warehouselocation" collection (see location.go), each referencing
// this warehouse by WarehouseGuid.
type Warehouse struct {
	models.PartitionIdentity `bson:"inline"`
	Code                     string          `json:"code" bson:"code"`
	Names                    *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Latitude                 float64         `json:"latitude" bson:"latitude"`
	Longitude                float64         `json:"longitude" bson:"longitude"`
	// CompanyGuids restricts which companies may use this warehouse. Empty = usable by every
	// company in the holding (existing convention, unchanged).
	CompanyGuids []string `json:"companyguids" bson:"companyguids"`
	Status       string   `json:"status" bson:"status"`
}

type WarehouseInfo struct {
	models.DocIdentity `bson:"inline"`
	Warehouse          `bson:"inline"`
}

func (WarehouseInfo) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseData struct {
	models.HoldingCodeentity `bson:"inline"`
	WarehouseInfo            `bson:"inline"`
}

type WarehouseDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WarehouseData      `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (WarehouseDoc) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (WarehouseItemGuid) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseActivity struct {
	WarehouseData       `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WarehouseActivity) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (WarehouseDeleteActivity) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseMessageQueue struct {
	models.HoldingCodeentity `bson:"inline"`
	models.DocIdentity       `bson:"inline"`
	Warehouse                `bson:"inline"`
}

// WarehousePG is the Kafka-consumer projection target (see warehouse_consumer.go /
// warehouse_phaser.go). It mirrors Warehouse master-data fields only — no embedded
// location/shelf, matching the MongoDB source shape.
type WarehousePG struct {
	models.HoldingCodeentity `gorm:"embedded;"`
	GuidFixed                string       `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	Code                     string       `json:"code" gorm:"column:code"`
	Names                    models.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	Latitude                 float64      `json:"latitude" gorm:"column:latitude"`
	Longitude                float64      `json:"longitude" gorm:"column:longitude"`
}

func (WarehousePG) TableName() string {
	return "warehouse"
}

func (jd *WarehousePG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *WarehousePG) CompareTo(other *WarehousePG) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(WarehousePG{}, "guidfixed"),
	)

	return diff == ""
}
