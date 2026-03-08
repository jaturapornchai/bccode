package saleinvoicereturn

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	saleInvoiceReturnModels "smlcloudplatform/internal/transaction/saleinvoicereturn/models"
)

type SaleInvoiceReturnTransactionPhaser struct{}

func (p *SaleInvoiceReturnTransactionPhaser) PhaseSaleInvoiceReturnDoc(doc saleInvoiceReturnModels.SaleInvoiceReturnDoc) (*models.SaleInvoiceReturnTransactionPG, error) {

	details := make([]models.SaleInvoiceReturnTransactionDetailPG, len(*doc.Transaction.Details))

	for i, detail := range *doc.Transaction.Details {
		stockDetail := models.SaleInvoiceReturnTransactionDetailPG{
			TransactionDetailPG: models.TransactionDetailPG{
				GuidFixed:           doc.GuidFixed,
				DocNo:               doc.DocNo,
				ShopID:              doc.ShopID,
				LineNumber:          int8(detail.LineNumber),
				DocRef:              detail.DocRef,
				Barcode:             detail.Barcode,
				Qty:                 detail.Qty,
				Price:               detail.Price,
				PriceExcludeVat:     detail.PriceExcludeVat,
				Discount:            detail.Discount,
				DiscountAmount:      detail.DiscountAmount,
				SumAmount:           detail.SumAmount,
				SumAmountExcludeVat: detail.SumAmountExcludeVat,
				WhCode:              detail.WhCode,
				LocationCode:        detail.LocationCode,
				VatType:             detail.VatType,
				TaxType:             detail.TaxType,
				StandValue:          detail.StandValue,
				DivideValue:         detail.DivideValue,
				ItemType:            detail.ItemType,
				ItemGuid:            detail.ItemGuid,
				TotalValueVat:       detail.TotalValueVat,
				Remark:              detail.Remark,
				UnitCode:            detail.UnitCode,
				UnitNames:           *pkgModels.DefaultArrayNameX(detail.UnitNames),
				ItemNames:           *pkgModels.DefaultArrayNameX(detail.ItemNames),
				WhNames:             *pkgModels.DefaultArrayNameX(detail.WhNames),
				LocationNames:       *pkgModels.DefaultArrayNameX(detail.LocationNames),
				GroupCode:           detail.GroupCode,
				GroupNames:          *pkgModels.DefaultArrayNameX(detail.GroupNames),
				DocDate:             detail.DocDatetime,
			},
			ManufacturerGUID:  detail.ManufacturerGUID,
			ManufacturerCode:  detail.ManufacturerCode,
			ManufacturerNames: *pkgModels.DefaultArrayNameX(detail.ManufacturerNames),
		}
		details[i] = stockDetail
	}

	totalPayCreditAmount := float64(0)
	totalPayTransfer := float64(0)
	if doc.PaymentDetail.PaymentCreditCards != nil {

		for _, creditCard := range *doc.PaymentDetail.PaymentCreditCards {
			totalPayCreditAmount += creditCard.Amount
		}
	}
	if doc.PaymentDetail.PaymentTransfers != nil {

		for _, transfer := range *doc.PaymentDetail.PaymentTransfers {
			totalPayTransfer += transfer.Amount
		}
	}

	if doc.Branch.Names == nil {
		doc.Branch.Names = &[]pkgModels.NameX{}
	}

	stockTransaction := models.SaleInvoiceReturnTransactionPG{
		TransactionPG: models.TransactionPG{
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			GuidFixed:           doc.GuidFixed,
			GuidRef:             doc.GuidRef,
			GuidPos:             doc.GuidPos,
			DeviceName:          doc.DeviceName,
			InquiryType:         doc.InquiryType,
			TransFlag:           48,
			DocNo:               doc.DocNo,
			DocDate:             doc.DocDatetime,
			TaxDocDate:          doc.Transaction.TaxDocDate,
			TaxDocNo:            doc.Transaction.TaxDocNo,
			VatType:             doc.Transaction.VatType,
			VatRate:             doc.Transaction.VatRate,
			DocRefType:          doc.Transaction.DocRefType,
			DocRefNo:            doc.Transaction.DocRefNo,
			DocRefDate:          doc.Transaction.DocRefDate,
			Description:         doc.Transaction.Description,
			TotalValue:          doc.Transaction.TotalValue,
			DiscountWord:        doc.Transaction.DiscountWord,
			TotalDiscount:       doc.Transaction.TotalDiscount,
			TotalBeforeVat:      doc.Transaction.TotalBeforeVat,
			TotalVatValue:       doc.Transaction.TotalVatValue,
			TotalAfterVat:       doc.Transaction.TotalAfterVat,
			TotalExceptVat:      doc.Transaction.TotalExceptVat,
			TotalAmount:         doc.Transaction.TotalAmount,
			PointDiscountAmount: doc.Transaction.PointDiscountAmount,
			PayPointAmount:      doc.Transaction.PayPointAmount,
			IsCancel:            doc.IsCancel,
			BranchCode:          doc.Branch.Code,
			BranchNames:         *pkgModels.DefaultArrayNameX(doc.Branch.Names),
		},
		IsPOS:                        doc.IsPOS,
		SaleCode:                     doc.SaleCode,
		SaleName:                     doc.SaleName,
		DetailDiscountFormula:        doc.DetailDiscountFormula,
		DetailTotalAmount:            doc.DetailTotalAmount,
		TotalDiscountVatAmount:       doc.TotalDiscountVatAmount,
		TotalDiscountExceptVatAmount: doc.TotalDiscountExceptVatAmount,
		DetailTotalDiscount:          doc.DetailTotalDiscount,
		PointsCode:                   doc.PointsCode,

		DebtorCode:       doc.CustCode,
		DebtorNames:      *pkgModels.DefaultArrayNameX(doc.CustNames),
		TotalPayCash:     doc.Transaction.PaymentDetail.CashAmount,
		TotalPayCredit:   totalPayCreditAmount,
		TotalPayTransfer: totalPayTransfer,
		Items:            &details,
	}
	return &stockTransaction, nil
}

func (p SaleInvoiceReturnTransactionPhaser) PhaseSingleDoc(input string) (*models.SaleInvoiceReturnTransactionPG, error) {

	doc := saleInvoiceReturnModels.SaleInvoiceReturnDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal SaleInvoice Return Doc Error: " + err.Error())
	}

	transaction, err := p.PhaseSaleInvoiceReturnDoc(doc)
	if err != nil {
		return nil, errors.New("Cannot Phase SaleInvoice Return Doc Error: " + err.Error())
	}

	return transaction, nil
}

func (p SaleInvoiceReturnTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.SaleInvoiceReturnTransactionPG, error) {

	docs := []saleInvoiceReturnModels.SaleInvoiceReturnDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal SaleInvoice Return Doc Error: " + err.Error())
	}

	transactions := make([]models.SaleInvoiceReturnTransactionPG, len(docs))
	for i, doc := range docs {
		transaction, err := p.PhaseSaleInvoiceReturnDoc(doc)
		if err != nil {
			return nil, errors.New("Cannot Phase SaleInvoice Return Doc Error: " + err.Error())
		}
		transactions[i] = *transaction
	}
	return &transactions, nil
}

// func (p *StockTransactionSaleInvoicePhaser) PhaseSaleReturnDoc(doc saleReturnModel.SaleInvoiceReturnDoc) (*models.StockTransaction, error) {

// }

// func (p StockTransactionSaleInvoicePhaser) PhaseSingleDoc(msg string) (*models.StockTransaction, error) {

// }

// func (p StockTransactionSaleInvoicePhaser) PhaseMultipleDoc(input string) (*[]models.StockTransaction, error) {

// 	docs := []saleReturnModel.SaleInvoiceReturnDoc{}
// 	err := json.Unmarshal([]byte(input), &docs)
// 	if err != nil {
// 		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
// 	}

// 	stockTransactions := make([]models.StockTransaction, len(docs))

// 	for i, doc := range docs {
// 		trx, err := p.PhaseSaleReturnDoc(doc)
// 		if err != nil {
// 			return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
// 		}
// 		stockTransactions[i] = *trx
// 	}

// 	return &stockTransactions, nil
// }
