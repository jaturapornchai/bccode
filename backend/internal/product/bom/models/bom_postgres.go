package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

// JSONB type to handle JSONB columns in PostgreSQL efficiently
type JSONB struct {
	Raw json.RawMessage
}

// Value implements the driver.Valuer interface for JSONB serialization.
func (j JSONB) Value() (driver.Value, error) {
	if len(j.Raw) == 0 {
		return nil, nil
	}
	return j.Raw, nil
}

// Scan implements the sql.Scanner interface for JSONB deserialization.
func (j *JSONB) Scan(value interface{}) error {
	if value == nil {
		j.Raw = nil
		return nil
	}
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	j.Raw = append(j.Raw[:0], b...)
	return nil
}

// BomProductBarcodePg holds data for the BOM product barcode
type BomProductBarcodePg struct {
	BarcodeGuidFixed string  `json:"guidfixed" gorm:"column:guidfixed"`
	Level            int     `json:"level" gorm:"column:level"`
	Names            JSONB   `json:"names" gorm:"column:names;type:jsonb"`
	ItemUnitCode     string  `json:"itemunitcode" gorm:"column:itemunitcode"`
	ItemUnitNames    JSONB   `json:"itemunitnames" gorm:"column:itemunitnames;type:jsonb"`
	Barcode          string  `json:"barcode" gorm:"column:barcode" validate:"required,min=1"`
	Condition        bool    `json:"condition" gorm:"column:condition"`
	DivideValue      float64 `json:"dividevalue" gorm:"column:dividevalue"`
	StandValue       float64 `json:"standvalue" gorm:"column:standvalue"`
	Qty              float64 `json:"qty" gorm:"column:qty"`
}

// ProductBarcodeBOMViewPG represents a view for product barcode BOMs
type ProductBarcodeBOMViewPG struct {
	ShopID            string              `json:"shopid" gorm:"column:shopid"`
	GuidFixed         string              `json:"guidfixed" bson:"guidfixed" gorm:"column:guidfixed;primaryKey"`
	BOMProductBarcode BomProductBarcodePg `gorm:"embedded;"`
	ImageURI          string              `json:"imageuri" gorm:"column:imageuri"`
	BOM               BOMViewPg           `json:"bom" gorm:"foreignKey:BOMForeignKey;references:BOMReference"` // Update ForeignKey and References
}

// TableName sets the custom table name for GORM
func (b *ProductBarcodeBOMViewPG) TableName() string {
	return "productbarcodeboms"
}

// CompareTo provides comparison for two ProductBarcodeBOMViewPG objects
func (s *ProductBarcodeBOMViewPG) CompareTo(other *ProductBarcodeBOMViewPG) bool {
	diff := cmp.Diff(s, other, cmpopts.IgnoreFields(ProductBarcodeBOMViewPG{}, "ShopID", "GuidFixed"))
	return diff == ""
}

// BOMViewPg defines a list of ProductBarcodeBOMViewPG for custom DB operations
type BOMViewPg []ProductBarcodeBOMViewPG

// Value serializes the BOMViewPg to JSON for database storage
func (a BOMViewPg) Value() (driver.Value, error) {
	j, err := json.Marshal(a)
	return j, err
}

// Scan deserializes JSON from the database into BOMViewPg
func (a *BOMViewPg) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
