package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const pointTransactionCollectionName = "pointTransactions"

type PointTransaction struct {
	models.PartitionIdentity `bson:"inline"`
	TransactionDocNo         string    `json:"transactiondocno" bson:"transactiondocno"`
	TransactionDate          time.Time `json:"transaction_date" bson:"transaction_date"`
	DebtorCode               string    `json:"debtorcode" bson:"debtorcode"`           // Customer who made the transaction
	PointsCode               string    `json:"points_code" bson:"points_code"`         // Customer who receives the points
	TransactionType          int8      `json:"transactiontype" bson:"transactiontype"` // 1=earn, 2=redeem
	PointAmount              float64   `json:"point_amount" bson:"point_amount"`
	BalanceBefore            float64   `json:"balancebefore" bson:"balancebefore"`
	BalanceAfter             float64   `json:"balanceafter" bson:"balanceafter"`
	Description              string    `json:"description" bson:"description"`
}

type PointTransactionInfo struct {
	models.DocIdentity `bson:"inline"`
	PointTransaction   `bson:"inline"`
}

func (PointTransactionInfo) CollectionName() string {
	return pointTransactionCollectionName
}

type PointTransactionData struct {
	models.HoldingCodeentity `bson:"inline"`
	PointTransactionInfo     `bson:"inline"`
}

type PointTransactionDoc struct {
	ID                   primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	PointTransactionData `bson:"inline"`
	models.ActivityDoc   `bson:"inline"`
}

func (PointTransactionDoc) CollectionName() string {
	return pointTransactionCollectionName
}

type PointTransactionItemGuid struct {
	ItemGUID string `json:"item_guid" bson:"item_guid"`
	GuidRef  string `json:"guid_ref" bson:"guid_ref"`
}

func (PointTransactionItemGuid) CollectionName() string {
	return pointTransactionCollectionName
}

type PointTransactionActivity struct {
	PointTransactionDoc `bson:"inline"`
}

func (PointTransactionActivity) CollectionName() string {
	return pointTransactionCollectionName
}

type PointTransactionDeleteActivity struct {
	models.HoldingCodeentity `bson:"inline"`
	models.DocIdentity       `bson:"inline"`
	models.ActivityDoc       `bson:"inline"`
}

func (PointTransactionDeleteActivity) CollectionName() string {
	return pointTransactionCollectionName
}

// OpeningBalancePointRequest represents a request to create opening balance points
type OpeningBalancePointRequest struct {
	PointsCode  string  `json:"points_code" binding:"required"`
	PointAmount float64 `json:"point_amount" binding:"required,gt=0"`
	Description string  `json:"description,omitempty"`
}

// BulkImportPointResult represents the result of bulk import operation
type BulkImportPointResult struct {
	Success      int                         `json:"success"`
	Failed       int                         `json:"failed"`
	Total        int                         `json:"total"`
	FailedItems  []BulkImportPointFailedItem `json:"failed_items,omitempty"`
	SuccessItems []string                    `json:"success_items,omitempty"`
}

// BulkImportPointFailedItem represents a failed item in bulk import
type BulkImportPointFailedItem struct {
	PointsCode  string  `json:"points_code"`
	PointAmount float64 `json:"point_amount"`
	Reason      string  `json:"reason"`
}
