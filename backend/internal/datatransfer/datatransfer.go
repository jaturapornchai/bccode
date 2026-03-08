package datatransfer

import "context"

type IDataTransfer interface {
	StartTransfer(ctx context.Context, shopID string, targetShopID string) error
	// CheckingBeforeTransfer(ctx context.Context, shopID string) (bool, error)
}

type IDBTransfer interface {
	BeginTransfer(shopID string, targetShopID string)
}
