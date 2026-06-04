package saleinvoicereturn

import (
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
)

type SaleInvoiceReturnDebtorTransactionPhaser struct{}

func (p SaleInvoiceReturnDebtorTransactionPhaser) PhaseSingleDoc(doc models.SaleInvoiceReturnTransactionPG) (*models.DebtorTransactionPG, error) {

	transaction, err := p.PhaseSaleInvoiceDebtorDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return transaction, err
}

func (p SaleInvoiceReturnDebtorTransactionPhaser) PhaseSaleInvoiceDebtorDoc(doc models.SaleInvoiceReturnTransactionPG) (*models.DebtorTransactionPG, error) {
	transaction := models.DebtorTransactionPG{
		HoldingCodeentity: pkgModels.HoldingCodeentity{
			HoldingCode: doc.HoldingCode,
		},
		GuidFixed:      doc.GuidFixed,
		DocNo:          doc.DocNo,
		DocDate:        doc.DocDate,
		DebtorCode:     doc.DebtorCode,
		DebtorNames:    doc.DebtorNames,
		InquiryType:    doc.InquiryType,
		TransFlag:      doc.TransFlag,
		TotalValue:     doc.TotalValue,
		TotalBeforeVat: doc.TotalBeforeVat,
		TotalAfterVat:  doc.TotalAfterVat,
		TotalVatValue:  doc.TotalVatValue,
		TotalExceptVat: doc.TotalExceptVat,
		TotalAmount:    doc.TotalAmount,
		BalanceAmount:  doc.TotalAmount,
		PaidAmount:     0,
	}

	return &transaction, nil
}
