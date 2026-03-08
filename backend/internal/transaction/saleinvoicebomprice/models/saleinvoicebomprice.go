package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const saleInvoiceBOMPriceCollectionName = "transactionSaleinvoiceBOMPrices"

type SaleInvoiceBomPrice struct {
	BOMGuid string             `json:"bomguid" bson:"bomguid"`
	DocNo   string             `json:"docno" bson:"docno"`
	Prices  []SaleInvoicePrice `json:"prices" bson:"prices"`
	Barcode string             `json:"barcode" bson:"barcode"`
	Qty     float64            `json:"qty" bson:"qty"`
	Price   float64            `json:"price" bson:"price"`
	Ratio   float64            `json:"ratio" bson:"ratio"`
}

type SaleInvoicePrice struct {
	Barcode string  `json:"barcode" bson:"barcode"`
	Qty     float64 `json:"qty" bson:"qty"`
	Price   float64 `json:"price" bson:"price"`
	Ratio   float64 `json:"ratio" bson:"ratio"`
}

type SaleInvoiceBomPriceInfo struct {
	models.DocIdentity  `bson:"inline"`
	SaleInvoiceBomPrice `bson:"inline"`
}

func (SaleInvoiceBomPriceInfo) CollectionName() string {
	return saleInvoiceBOMPriceCollectionName
}

type SaleInvoiceBomPriceData struct {
	models.ShopIdentity     `bson:"inline"`
	SaleInvoiceBomPriceInfo `bson:"inline"`
}

type SaleInvoiceBomPriceDoc struct {
	ID                      primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	SaleInvoiceBomPriceData `bson:"inline"`
	models.ActivityDoc      `bson:"inline"`
}

func (SaleInvoiceBomPriceDoc) CollectionName() string {
	return saleInvoiceBOMPriceCollectionName
}

type SaleInvoiceBomPriceItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (SaleInvoiceBomPriceItemGuid) CollectionName() string {
	return saleInvoiceBOMPriceCollectionName
}

type SaleInvoiceBomPriceActivity struct {
	SaleInvoiceBomPriceData `bson:"inline"`
	models.ActivityTime     `bson:"inline"`
}

func (SaleInvoiceBomPriceActivity) CollectionName() string {
	return saleInvoiceBOMPriceCollectionName
}

type SaleInvoiceBomPriceDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (SaleInvoiceBomPriceDeleteActivity) CollectionName() string {
	return saleInvoiceBOMPriceCollectionName
}
