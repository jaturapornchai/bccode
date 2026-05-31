package models

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const warehouseCollectionName = "warehouse"

type Warehouse struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Location *[]Location     `json:"location" bson:"location" validate:"omitempty,unique=Code,dive"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

type Location struct {
	Code                 string          `json:"code" bson:"code"`
	Names                *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	Shelf                []Shelf         `json:"shelf" bson:"shelf" validate:"omitempty,unique=Code,dive"`
	SuitableProductTypes string          `json:"suitable_product_types" bson:"suitable_product_types"`
}

type Shelf struct {
	Code                 string                `json:"code" bson:"code"`
	Name                 string                `json:"name" bson:"name" validate:"required,min=1"`
	Min                  int                   `json:"min" bson:"min" validate:"omitempty,gte=0"`
	Max                  int                   `json:"max" bson:"max" validate:"omitempty,gte=0"`
	ProductItems         []ShelfProductBarcode `json:"productitems" bson:"productitems"`
	MaxWeight            float64               `json:"max_weight" bson:"max_weight"`
	Width                float64               `json:"width" bson:"width"`
	Length               float64               `json:"length" bson:"length"`
	Height               float64               `json:"height" bson:"height"`
	SuitableProductTypes string                `json:"suitable_product_types" bson:"suitable_product_types"`
}

type ShelfProductBarcode struct {
	GuidFixed string          `json:"guid_fixed" bson:"guid_fixed"`
	Barcode string          `json:"barcode" bson:"barcode"`
	Unitcode string          `json:"unitcode" bson:"unitcode" validate:"required"`
	UnitNames *[]models.NameX `json:"unitnames" bson:"unitnames"`
	Names *[]models.NameX `json:"names" bson:"names"`
}

// Helper methods for Shelf
func (s *Shelf) AddProductBarcode(guidFixed, barcode string, names *[]models.NameX) {
	// Check if product already exists
	for i, product := range s.ProductItems {
		if product.GuidFixed == guidFixed {
			// Update existing product
			s.ProductItems[i].Barcode = barcode
			s.ProductItems[i].Names = names
			return
		}
	}
	// Add new product
	s.ProductItems = append(s.ProductItems, ShelfProductBarcode{
		GuidFixed: guidFixed,
		Barcode:   barcode,
		Names:     names,
	})
}

func (s *Shelf) RemoveProductBarcode(guidFixed string) bool {
	for i, product := range s.ProductItems {
		if product.GuidFixed == guidFixed {
			s.ProductItems = append(s.ProductItems[:i], s.ProductItems[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Shelf) HasProductBarcode(guidFixed string) bool {
	for _, product := range s.ProductItems {
		if product.GuidFixed == guidFixed {
			return true
		}
	}
	return false
}

func (s *Shelf) GetProductBarcodeByGuid(guidFixed string) *ShelfProductBarcode {
	for _, product := range s.ProductItems {
		if product.GuidFixed == guidFixed {
			return &product
		}
	}
	return nil
}

// Bulk operations request models
type BulkShelfProductRequest struct {
	Products []ShelfProductBarcode `json:"products" validate:"required,dive"`
}

type BulkRemoveProductRequest struct {
	ProductGuidFixedList []string `json:"productguidfixedlist" validate:"required,min=1"`
}

type BulkProductOperationResponse struct {
	Success bool                  `json:"success"`
	TotalCount int                   `json:"totalcount"`
	SuccessCount int                   `json:"successcount"`
	FailedCount int                   `json:"failedcount"`
	Errors []BulkOperationError  `json:"errors,omitempty"`
	Results []BulkOperationResult `json:"results,omitempty"`
}

type BulkOperationError struct {
	Index int    `json:"index"`
	ProductGuid string `json:"productguid,omitempty"`
	ErrorMessage string `json:"errormessage"`
}

type BulkOperationResult struct {
	Index int    `json:"index"`
	ProductGuid string `json:"productguid"`
	Status string `json:"status"` // "success", "updated", "removed"
}

// Bulk operations methods for Shelf
func (s *Shelf) AddMultipleProductBarcodes(products []ShelfProductBarcode) BulkProductOperationResponse {
	response := BulkProductOperationResponse{
		TotalCount: len(products),
		Results:    []BulkOperationResult{},
		Errors:     []BulkOperationError{},
	}

	for i, product := range products {
		// Check if product already exists
		exists := s.HasProductBarcode(product.GuidFixed)

		s.AddProductBarcode(product.GuidFixed, product.Barcode, product.Names)

		status := "success"
		if exists {
			status = "updated"
		}

		response.Results = append(response.Results, BulkOperationResult{
			Index:       i,
			ProductGuid: product.GuidFixed,
			Status:      status,
		})
		response.SuccessCount++
	}

	response.Success = response.FailedCount == 0
	return response
}

func (s *Shelf) RemoveMultipleProductBarcodes(guidFixedList []string) BulkProductOperationResponse {
	response := BulkProductOperationResponse{
		TotalCount: len(guidFixedList),
		Results:    []BulkOperationResult{},
		Errors:     []BulkOperationError{},
	}

	for i, guidFixed := range guidFixedList {
		if s.RemoveProductBarcode(guidFixed) {
			response.Results = append(response.Results, BulkOperationResult{
				Index:       i,
				ProductGuid: guidFixed,
				Status:      "removed",
			})
			response.SuccessCount++
		} else {
			response.Errors = append(response.Errors, BulkOperationError{
				Index:        i,
				ProductGuid:  guidFixed,
				ErrorMessage: "Product not found",
			})
			response.FailedCount++
		}
	}

	response.Success = response.FailedCount == 0
	return response
}

type LocationInfo struct {
	GuidFixed string          `json:"guid_fixed" bson:"guid_fixed"`
	WarehouseCode string          `json:"warehousecode" bson:"warehousecode"`
	WarehouseNames *[]models.NameX `json:"warehousenames" bson:"warehousenames"`
	LocationCode string          `json:"locationcode" bson:"locationcode"`
	LocationNames *[]models.NameX `json:"locationnames" bson:"locationnames"`
	Shelf []Shelf         `json:"shelf" bson:"shelf"`
}

func (LocationInfo) CollectionName() string {
	return warehouseCollectionName
}

type ShelfInfo struct {
	GuidFixed string          `json:"guid_fixed" bson:"guid_fixed"`
	WarehouseCode string          `json:"warehousecode" bson:"warehousecode"`
	WarehouseNames *[]models.NameX `json:"warehousenames" bson:"warehousenames"`
	LocationCode string          `json:"locationcode" bson:"locationcode"`
	LocationNames *[]models.NameX `json:"locationnames" bson:"locationnames"`
	ShelfCode string          `json:"shelfcode" bson:"shelfcode"`
	ShelfName string          `json:"shelfname" bson:"shelfname"`
}

func (ShelfInfo) CollectionName() string {
	return warehouseCollectionName
}

type LocationRequest struct {
	WarehouseCode string          `json:"warehousecode" bson:"warehousecode" validate:"required"`
	Code string          `json:"locationcode" bson:"locationcode" validate:"required"`
	Names *[]models.NameX `json:"locationnames" bson:"locationnames"`
	Shelf []Shelf         `json:"shelf" bson:"shelf"`
}

type ShelfRequest struct {
	WarehouseCode string `json:"warehousecode" bson:"warehousecode" validate:"required"`
	LocationCode string `json:"locationcode" bson:"locationcode" validate:"required"`
	Code string `json:"shelfcode" bson:"shelfcode" validate:"required"`
	Name string `json:"shelfname" bson:"shelfname"`
}

type WarehouseInfo struct {
	models.DocIdentity `bson:"inline"`
	Warehouse  `bson:"inline"`
}

func (WarehouseInfo) CollectionName() string {
	return warehouseCollectionName
}

type WarehouseData struct {
	models.ShopIdentity `bson:"inline"`
	WarehouseInfo  `bson:"inline"`
}

type WarehouseDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	WarehouseData  `bson:"inline"`
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
	WarehouseData  `bson:"inline"`
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
	models.ShopIdentity `bson:"inline"`
	models.DocIdentity  `bson:"inline"`
	Warehouse  `bson:"inline"`
}

type WarehousePG struct {
	models.ShopIdentity `gorm:"embedded;"`
	GuidFixed string       `json:"guid_fixed" gorm:"column:guid_fixed;primaryKey"`
	Code string       `json:"code" gorm:"column:code"`
	Names models.JSONB `json:"names" gorm:"column:names;type:jsonb"`
	Location LocationsPG  `json:"location" gorm:"column:location;type:jsonb"`
	Latitude  float64      `json:"latitude" gorm:"column:latitude"`
	Longitude float64      `json:"longitude" gorm:"column:longitude"`
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
		cmpopts.IgnoreFields(WarehousePG{}, "guid_fixed"),
	)

	return diff == ""
}

type LocationsPG []Location

func (a LocationsPG) Value() (driver.Value, error) {

	j, err := json.Marshal(a)
	return j, err
}

func (a *LocationsPG) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &a)
}
