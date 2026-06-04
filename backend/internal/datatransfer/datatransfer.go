package datatransfer

import "context"

type IDataTransfer interface {
	StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error
	// CheckingBeforeTransfer(ctx context.Context, holdingCode string) (bool, error)
}

type IDBTransfer interface {
	BeginTransfer(holdingCode string, targetHoldingCode string)
}
