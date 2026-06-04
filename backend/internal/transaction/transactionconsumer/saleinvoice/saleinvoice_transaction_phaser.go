package saleinvoice

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	saleInvoiceModel "smlcloudplatform/internal/transaction/saleinvoice/models"
)

type SalesInvoiceTransactionPhaser struct{}

func (p SalesInvoiceTransactionPhaser) PhaseSaleInvoiceDoc(doc saleInvoiceModel.SaleInvoiceDoc) (*models.SaleInvoiceTransactionPG, error) {

	details := make([]models.SaleInvoiceTransactionDetailPG, len(*doc.Details))

	for i, detail := range *doc.Details {

		stockDetail := models.SaleInvoiceTransactionDetailPG{
			TransactionDetailPG: models.TransactionDetailPG{
				GuidFixed:   doc.GuidFixed,
				DocNo:       doc.DocNo,
				HoldingCode: doc.HoldingCode,

				LineNumber:          int8(detail.LineNumber),
				FoodType:            detail.FoodType,
				DocRef:              detail.DocRef,
				Barcode:             detail.Barcode,
				Qty:                 detail.Qty,
				Price:               detail.Price,
				PriceExcludeVat:     detail.PriceExcludeVat,
				Discount:            detail.Discount,
				DiscountAmount:      detail.DiscountAmount,
				SumAmount:           detail.SumAmount,
				SumAmountExcludeVat: detail.SumAmountExcludeVat,
				SumAmountChoice:     detail.SumAmountChoice,
				IsChoice:            detail.IsChoice,
				RefGuid:             detail.RefGuid,
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

	stockTransaction := models.SaleInvoiceTransactionPG{
		TransactionPG: models.TransactionPG{
			HoldingCodeentity: pkgModels.HoldingCodeentity{
				HoldingCode: doc.HoldingCode,
			},

			GuidFixed:           doc.GuidFixed,
			GuidRef:             doc.GuidRef,
			GuidPos:             doc.GuidPos,
			DeviceName:          doc.DeviceName,
			InquiryType:         doc.InquiryType,
			TransFlag:           44,
			DocNo:               doc.DocNo,
			DocDate:             doc.DocDatetime,
			TaxDocDate:          doc.TaxDocDate,
			TaxDocNo:            doc.TaxDocNo,
			VatType:             doc.VatType,
			VatRate:             doc.VatRate,
			DocRefType:          doc.DocRefType,
			DocRefNo:            doc.DocRefNo,
			DocRefDate:          doc.DocRefDate,
			Description:         doc.Description,
			TotalValue:          doc.TotalValue,
			DiscountWord:        doc.DiscountWord,
			TotalDiscount:       doc.TotalDiscount,
			TotalBeforeVat:      doc.TotalBeforeVat,
			TotalVatValue:       doc.TotalVatValue,
			TotalAfterVat:       doc.TotalAfterVat,
			TotalExceptVat:      doc.TotalExceptVat,
			TotalAmount:         doc.TotalAmount,
			PointDiscountAmount: doc.PointDiscountAmount,
			PayPointAmount:      doc.PayPointAmount,
			IsCancel:            doc.IsCancel,
			IsBom:               doc.IsBom,
			BranchCode:          doc.Branch.Code,
			BranchNames:         *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			AlcoholAmount:       doc.AlcoholAmount,
			DrinkAmount:         doc.DrinkAmount,
			OtherAmount:         doc.OtherAmount,
			FoodAmount:          doc.FoodAmount,
		},
		IsPOS:                        doc.IsPOS,
		IsBom:                        doc.IsBom,
		OrderNumber:                  doc.OrderNumber,
		DeliveryAmount:               doc.DeliveryAmount,
		SaleCode:                     doc.SaleCode,
		SaleName:                     doc.SaleName,
		DetailDiscountFormula:        doc.DetailDiscountFormula,
		DetailTotalAmount:            doc.DetailTotalAmount,
		TotalDiscountVatAmount:       doc.TotalDiscountVatAmount,
		TotalDiscountExceptVatAmount: doc.TotalDiscountExceptVatAmount,
		DetailTotalDiscount:          doc.DetailTotalDiscount,
		SaleChannelCode:              doc.SaleChannelCode,
		SaleChannelGP:                doc.SaleChannelGP,
		SaleChannelGPType:            doc.SaleChannelGPType,
		TakeAway:                     doc.TakeAway,
		PointsCode:                   doc.PointsCode,
		IsDelivery:                   doc.IsDelivery,
		IsTransport:                  doc.IsTransport,
		TransportCode:                doc.TransportCode,
		TransportAmount:              doc.TransportAmount,

		DebtorCode:       doc.CustCode,
		DebtorNames:      *pkgModels.DefaultArrayNameX(doc.CustNames),
		TotalPayCash:     doc.PaymentDetail.CashAmount,
		TotalPayCredit:   totalPayCreditAmount,
		TotalPayTransfer: totalPayTransfer,
		Items:            &details,
	}

	return &stockTransaction, nil
}

func (p SalesInvoiceTransactionPhaser) PhaseSingleDoc(msg string) (*models.SaleInvoiceTransactionPG, error) {

	doc := saleInvoiceModel.SaleInvoiceDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		//t.ms.Logger.Errorf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		// fmt.Printf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}
	trx, err := p.PhaseSaleInvoiceDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p SalesInvoiceTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.SaleInvoiceTransactionPG, error) {

	docs := []saleInvoiceModel.SaleInvoiceDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		//t.ms.Logger.Errorf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		// fmt.Printf("Cannot Unmarshal PurchaseDoc Message : %v", err.Error())
		return nil, errors.New("Cannot Unmarshal PurchaseDoc Message : " + err.Error())
	}

	stockTransactions := make([]models.SaleInvoiceTransactionPG, len(docs))

	for i, doc := range docs {
		trx, err := p.PhaseSaleInvoiceDoc(doc)
		if err != nil {
			//t.ms.Logger.Errorf("Cannot Phase PurchaseDoc to StockTransaction : %v", err.Error())
			// fmt.Printf("Cannot Phase PurchaseDoc to StockTransaction : %v", err.Error())
			return nil, errors.New("Error on Convert PurchaseDoc to StockTransaction : " + err.Error())
		}
		stockTransactions[i] = *trx
	}

	return &stockTransactions, nil
}
