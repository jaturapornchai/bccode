package transactionadmin

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/systemadmin/transactionadmin/creditorpayment"
	"smlcloudplatform/internal/systemadmin/transactionadmin/debtorpayment"
	"smlcloudplatform/internal/systemadmin/transactionadmin/purchase"
	"smlcloudplatform/internal/systemadmin/transactionadmin/purchasereceive"
	"smlcloudplatform/internal/systemadmin/transactionadmin/purchasereturn"
	"smlcloudplatform/internal/systemadmin/transactionadmin/saleinvoice"
	"smlcloudplatform/internal/systemadmin/transactionadmin/saleinvoicereturn"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stockadjustment"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stockbalance"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stockpickupproduct"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stockreceiveproduct"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stockreturnproduct"
	"smlcloudplatform/internal/systemadmin/transactionadmin/stocktransfer"
	"smlcloudplatform/pkg/microservice"
)

type ITransactionAdminHttp interface {
	RegisterHttp(ms *microservice.Microservice, prefix string)
}

type TransactionAdminHttp struct {
	purchaseAdminHttp          purchase.IPurchaseTransactionAdminHttp
	purchaseReceiveAdminHttp   purchasereceive.IPurchaseReceiveTransactionAdminHttp
	purchaseReturnAdminHttp    purchasereturn.IPurchaseReturnTransactionAdminHttp
	saleInvoiceAdminHttp       saleinvoice.ISaleInvoiceTransactionAdminHttp
	saleInvoiceReturnAdminHttp saleinvoicereturn.ISaleInvoiceReturnTransactionAdminHttp
	stockAdjustAdminHttp       stockadjustment.IStockAdjustmentTransactionAdminHttp
	stockReceiveProductHttp    stockreceiveproduct.IStockReceiveTransactionAdminHttp
	stockBalanceAdminHttp      stockbalance.IStockBalanceTransactionAdminHttp
	stockPickupProductHttp     stockpickupproduct.IStockPickupTransactionAdminHttp
	stockReturnProductHttp     stockreturnproduct.IStockReturnProductTransactionAdminHttp
	stockTransferHttp          stocktransfer.IStockTransferTransactionAdminHttp
	creditorPaymentAdminHttp   creditorpayment.ICreditorPaymentTransactionAdminHttp
	debtorPaymentAdminHttp     debtorpayment.IDebtorPaymentTransactionAdminHttp
}

func NewTransactionAdminHttp(ms *microservice.Microservice, cfg config.IConfig) ITransactionAdminHttp {

	purchaseAdminHttp := purchase.NewPurchaseTransactionAdminHttp(ms, cfg)
	purchaseReceiveAdminHttp := purchasereceive.NewPurchaseReceiveTransactionAdminHttp(ms, cfg)
	purchaseReturnAdminHttp := purchasereturn.NewPurchaseReturnTransactionAdminHttp(ms, cfg)
	saleInvoiceAdminHttp := saleinvoice.NewSaleInvoiceTransactionAdminHttp(ms, cfg)
	saleInvoiceReturnAdminHttp := saleinvoicereturn.NewSaleInvoiceReturnTransactionAdminHttp(ms, cfg)
	stockAdjustAdminHttp := stockadjustment.NewStockAdjustmentTransactionAdminHttp(ms, cfg)
	stockReceiveProductHttp := stockreceiveproduct.NewStockReceiveTransactionAdminHttp(ms, cfg)
	stockPickupProductHttp := stockpickupproduct.NewStockPickupTransactionAdminHttp(ms, cfg)
	stockReturnProductHttp := stockreturnproduct.NewStockReturnProductTransactionAdminHttp(ms, cfg)
	stockTransferHttp := stocktransfer.NewStockTransferTransactionAdminHttp(ms, cfg)
	creditorPaymentAdminHttp := creditorpayment.NewCreditorPaymentTransactionAdminHttp(ms, cfg)
	debtorPaymentAdminHttp := debtorpayment.NewDebtorPaymentTransactionAdminHttp(ms, cfg)
	stockBalanceAdminHttp := stockbalance.NewStockBalanceTransactionAdminHttp(ms, cfg)

	return &TransactionAdminHttp{
		purchaseAdminHttp:          purchaseAdminHttp,
		purchaseReceiveAdminHttp:   purchaseReceiveAdminHttp,
		purchaseReturnAdminHttp:    purchaseReturnAdminHttp,
		saleInvoiceAdminHttp:       saleInvoiceAdminHttp,
		saleInvoiceReturnAdminHttp: saleInvoiceReturnAdminHttp,
		stockAdjustAdminHttp:       stockAdjustAdminHttp,
		stockReceiveProductHttp:    stockReceiveProductHttp,
		stockPickupProductHttp:     stockPickupProductHttp,
		stockReturnProductHttp:     stockReturnProductHttp,
		stockTransferHttp:          stockTransferHttp,
		creditorPaymentAdminHttp:   creditorPaymentAdminHttp,
		debtorPaymentAdminHttp:     debtorPaymentAdminHttp,
		stockBalanceAdminHttp:      stockBalanceAdminHttp,
	}
}

func (s *TransactionAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	s.purchaseAdminHttp.RegisterHttp(ms, prefix)
	s.purchaseReceiveAdminHttp.RegisterHttp(ms, prefix)
	s.purchaseReturnAdminHttp.RegisterHttp(ms, prefix)
	s.saleInvoiceAdminHttp.RegisterHttp(ms, prefix)
	s.saleInvoiceReturnAdminHttp.RegisterHttp(ms, prefix)
	s.stockAdjustAdminHttp.RegisterHttp(ms, prefix)
	s.stockReceiveProductHttp.RegisterHttp(ms, prefix)
	s.stockPickupProductHttp.RegisterHttp(ms, prefix)
	s.stockReturnProductHttp.RegisterHttp(ms, prefix)
	s.stockTransferHttp.RegisterHttp(ms, prefix)
	s.creditorPaymentAdminHttp.RegisterHttp(ms, prefix)
	s.debtorPaymentAdminHttp.RegisterHttp(ms, prefix)
	s.stockBalanceAdminHttp.RegisterHttp(ms, prefix)
}
