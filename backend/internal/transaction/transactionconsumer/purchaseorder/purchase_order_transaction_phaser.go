package purchaseorder

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	purchaseOrderModel "smlcloudplatform/internal/transaction/purchaseorder/models"
)

type PurchaseOrderTransactionPhaser struct{}

func (p PurchaseOrderTransactionPhaser) PhaseSingleDoc(msg string) (*models.PurchaseOrderTransactionPG, error) {

	doc := purchaseOrderModel.PurchaseOrderDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseOrderDoc Message : " + err.Error())
	}
	trx, err := p.PhasePurchaseOrderTransactionDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseOrderDoc to StockTransaction : " + err.Error())
	}
	return trx, err
}

func (p PurchaseOrderTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.PurchaseOrderTransactionPG, error) {

	docs := []purchaseOrderModel.PurchaseOrderDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseOrderDoc Message : " + err.Error())
	}

	docsList := make([]models.PurchaseOrderTransactionPG, len(docs))
	for i, doc := range docs {
		trx, err := p.PhasePurchaseOrderTransactionDoc(doc)
		if err != nil {
			return nil, errors.New("Error on Convert PurchaseOrderDoc to StockTransaction : " + err.Error())
		}
		docsList[i] = *trx
	}

	return &docsList, nil
}

func (p PurchaseOrderTransactionPhaser) PhasePurchaseOrderTransactionDoc(doc purchaseOrderModel.PurchaseOrderDoc) (*models.PurchaseOrderTransactionPG, error) {

	details := []models.PurchaseOrderDetailTransactionPG{}

	if doc.Details != nil {
		details = make([]models.PurchaseOrderDetailTransactionPG, len(*doc.Details))

		for i, detail := range *doc.Details {

			purchaseOrderDetail := models.PurchaseOrderDetailTransactionPG{
				TransactionDetailPG: models.TransactionDetailPG{
					GuidFixed:           doc.GuidFixed,
					DocRef:              detail.DocRef,
					DocRefDateTime:      detail.DocRefDatetime,
					DocNo:               doc.DocNo,
					ShopID:              doc.ShopID,
					LineNumber:          int8(detail.LineNumber),
					Barcode:             detail.Barcode,
					Qty:                 detail.Qty,
					Price:               detail.Price,
					PriceExcludeVat:     detail.PriceExcludeVat,
					Discount:            detail.Discount,
					DiscountAmount:      detail.DiscountAmount,
					SumAmount:           detail.SumAmount,
					SumAmountExcludeVat: detail.SumAmountExcludeVat,
					TotalValueVat:       detail.TotalValueVat,
					WhCode:              detail.WhCode,
					LocationCode:        detail.LocationCode,
					VatType:             detail.VatType,
					TaxType:             detail.TaxType,
					StandValue:          detail.StandValue,
					DivideValue:         detail.DivideValue,
					ItemType:            detail.ItemType,
					ItemGuid:            detail.ItemGuid,
					Remark:              detail.Remark,
					UnitCode:            detail.UnitCode,
					UnitNames:           *pkgModels.DefaultArrayNameX(detail.UnitNames),
					ItemNames:           *pkgModels.DefaultArrayNameX(detail.ItemNames),
					WhNames:             *pkgModels.DefaultArrayNameX(detail.WhNames),
					LocationNames:       *pkgModels.DefaultArrayNameX(detail.LocationNames),
					GroupCode:           detail.GroupCode,
					GroupNames:          *pkgModels.DefaultArrayNameX(detail.GroupNames),
					DocDate:             detail.DocDatetime,
					VatCal:              int8(detail.VatCal),
				},
			}
			details[i] = purchaseOrderDetail
		}
	}

	purchaseOrderTransaction := models.PurchaseOrderTransactionPG{
		TransactionPG: models.TransactionPG{
			GuidFixed: doc.GuidFixed,
			GuidRef:   doc.GuidRef,
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			TransFlag:      6,
			DocNo:          doc.DocNo,
			DocDate:        doc.DocDatetime,
			BranchCode:     doc.Branch.Code,
			BranchNames:    *pkgModels.DefaultArrayNameX(doc.Branch.Names),
			TaxDocNo:       doc.TaxDocNo,
			TaxDocDate:     doc.TaxDocDate,
			Description:    doc.Description,
			InquiryType:    doc.InquiryType,
			VatType:        doc.VatType,
			VatRate:        doc.VatRate,
			DocRefType:     doc.DocRefType,
			DocRefNo:       doc.DocRefNo,
			DocRefDate:     doc.DocRefDate,
			TotalValue:     doc.TotalValue,
			DiscountWord:   doc.DiscountWord,
			TotalDiscount:  doc.TotalDiscount,
			TotalBeforeVat: doc.TotalBeforeVat,
			TotalVatValue:  doc.TotalVatValue,
			TotalExceptVat: doc.TotalExceptVat,
			TotalAfterVat:  doc.TotalAfterVat,
			TotalAmount:    doc.TotalAmount,
			IsCancel:       doc.IsCancel,
		},
		Items:         &details,
		CreditorCode:  doc.CustCode,
		CreditorNames: *pkgModels.DefaultArrayNameX(doc.CustNames),
	}

	return &purchaseOrderTransaction, nil
}
