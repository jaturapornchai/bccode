package models

import (
	"gorm.io/gorm"
	pkgModels "smlcloudplatform/internal/models"
	"time"
)

type WarehousePg struct {
	HoldingCode string          `json:"holdingcode" gorm:"column:holdingcode;index"`
	GuidFixed   string          `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	Code        string          `json:"code" gorm:"column:code;index:idxwarehousecode,unique"`
	Names       pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	Latitude    float64         `json:"latitude" gorm:"column:latitude"`
	Longitude   float64         `json:"longitude" gorm:"column:longitude"`
	IsActive    bool            `json:"isactive" gorm:"column:isactive;default:true"`
	CreatedAt   time.Time       `json:"createdat"`
	UpdatedAt   time.Time       `json:"updatedat"`
	DeletedAt   gorm.DeletedAt  `json:"deletedat" gorm:"index"`
	Zones       []ZonePg        `json:"location" gorm:"foreignKey:WarehouseGuid;references:GuidFixed"`
}

func (WarehousePg) TableName() string {
	return "warehouse"
}

type CompanyWarehousePg struct {
	CompanyGuid   string `json:"companyguid" gorm:"column:companyguid;primaryKey;index"`
	WarehouseGuid string `json:"warehouseguid" gorm:"column:warehouseguid;primaryKey;index"`
}

func (CompanyWarehousePg) TableName() string {
	return "companywarehouses"
}

type ZonePg struct {
	HoldingCode          string          `json:"holdingcode" gorm:"column:holdingcode;index"`
	GuidFixed            string          `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	WarehouseGuid        string          `json:"warehouseguid" gorm:"column:warehouseguid;index"`
	Code                 string          `json:"code" gorm:"column:code;index:idxzonecode,unique"`
	Names                pkgModels.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	SuitableProductTypes string          `json:"suitableproducttypes" gorm:"column:suitableproducttypes"`
	IsActive             bool            `json:"isactive" gorm:"column:isactive;default:true"`
	CreatedAt            time.Time       `json:"createdat"`
	UpdatedAt            time.Time       `json:"updatedat"`
	DeletedAt            gorm.DeletedAt  `json:"deletedat" gorm:"index"`
	Shelves              []ShelfPg       `json:"shelf" gorm:"foreignKey:ZoneGuid;references:GuidFixed"`
}

func (ZonePg) TableName() string {
	return "warehousezones"
}

type ShelfPg struct {
	HoldingCode string         `json:"holdingcode" gorm:"column:holdingcode;index"`
	GuidFixed   string         `json:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	ZoneGuid    string         `json:"zoneguid" gorm:"column:zoneguid;index"`
	Code        string         `json:"code" gorm:"column:code;index:idxshelfcode,unique"`
	Name        string         `json:"name" gorm:"column:name"`
	IsActive    bool           `json:"isactive" gorm:"column:isactive;default:true"`
	CreatedAt   time.Time      `json:"createdat"`
	UpdatedAt   time.Time      `json:"updatedat"`
	DeletedAt   gorm.DeletedAt `json:"deletedat" gorm:"index"`
}

func (ShelfPg) TableName() string {
	return "warehouseshelves"
}
