package models

import "fmt"

// InsufficientPointsError represents an error when a customer doesn't have enough points
type InsufficientPointsError struct {
	Current  float64
	Required float64
	CustCode string
}

func (e *InsufficientPointsError) Error() string {
	return fmt.Sprintf("customer %s has insufficient points: %.2f < %.2f",
		e.CustCode, e.Current, e.Required)
}

// InvalidPointAmountError represents an error when point amounts are invalid
type InvalidPointAmountError struct {
	Amount    float64
	Operation string
	Reason    string
}

func (e *InvalidPointAmountError) Error() string {
	return fmt.Sprintf("invalid point amount for %s: %.2f (%s)",
		e.Operation, e.Amount, e.Reason)
}

// PointTransactionError represents a general point transaction error
type PointTransactionError struct {
	Operation string
	CustCode  string
	Reason    string
	Err       error
}

func (e *PointTransactionError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("point transaction %s failed for customer %s: %s (%v)",
			e.Operation, e.CustCode, e.Reason, e.Err)
	}
	return fmt.Sprintf("point transaction %s failed for customer %s: %s",
		e.Operation, e.CustCode, e.Reason)
}

func (e *PointTransactionError) Unwrap() error {
	return e.Err
}

// Point Transaction Constants
const (
	TransactionTypeEarn          = 1
	TransactionTypeRedeem        = 2
	TransactionTypeOpeningBalance = 3  // การเพิ่มแต้มด้วยตนเอง (Manual Add - can be used for opening balance, adjustments, promotions, etc.)
	TransactionTypeAdjustment    = 4  // สำรอง (Reserved for future use)
	TransactionTypeCancelSale    = 5
	TransactionTypeCancelRefund  = 6
	TransactionTypeReturnCancel  = 7
	TransactionTypeReturnRefund  = 8
)
