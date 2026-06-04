package models

import (
	"gorm.io/gorm"
	pkgModels "smlcloudplatform/internal/models"
	"time"
)

type WarehousePg struct {
	HoldingCode string          `json:"holding_code" gorm:"column:holding_code;index"`
	GuidFixed   string          `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	Code        string          `json:"code" gorm:"column:code;index:idx_warehouse_code,unique"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	Latitude    float64         `json:"latitude" gorm:"column:latitude"`
	Longitude   float64         `json:"longitude" gorm:"column:longitude"`
	IsActive    bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
	Zones       []ZonePg        `json:"location" gorm:"foreignKey:WarehouseGuid;references:GuidFixed"`
}

func (WarehousePg) TableName() string {
	return "warehouse"
}

type CompanyWarehousePg struct {
	CompanyGuid   string `json:"company_guid" gorm:"column:company_guid;primaryKey;index"`
	WarehouseGuid string `json:"warehouse_guid" gorm:"column:warehouse_guid;primaryKey;index"`
}

func (CompanyWarehousePg) TableName() string {
	return "company_warehouses"
}

type ZonePg struct {
	HoldingCode          string          `json:"holding_code" gorm:"column:holding_code;index"`
	GuidFixed            string          `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	WarehouseGuid        string          `json:"warehouse_guid" gorm:"column:warehouse_guid;index"`
	Code                 string          `json:"code" gorm:"column:code;index:idx_zone_code,unique"`
	Names                pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	SuitableProductTypes string          `json:"suitable_product_types" gorm:"column:suitable_product_types"`
	IsActive             bool            `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
	DeletedAt            gorm.DeletedAt  `json:"deleted_at" gorm:"index"`
	Shelves              []ShelfPg       `json:"shelf" gorm:"foreignKey:ZoneGuid;references:GuidFixed"`
}

func (ZonePg) TableName() string {
	return "warehouse_zones"
}

type ShelfPg struct {
	HoldingCode string         `json:"holding_code" gorm:"column:holding_code;index"`
	GuidFixed   string         `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	ZoneGuid    string         `json:"zone_guid" gorm:"column:zone_guid;index"`
	Code        string         `json:"code" gorm:"column:code;index:idx_shelf_code,unique"`
	Name        string         `json:"name" gorm:"column:name"`
	IsActive    bool           `json:"is_active" gorm:"column:is_active;default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"deleted_at" gorm:"index"`
}

func (ShelfPg) TableName() string {
	return "warehouse_shelves"
}
