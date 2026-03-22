package purchaserequisition

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	prModel "smlcloudplatform/internal/transaction/purchaserequisition/models"
)

type PurchaseRequisitionTransactionPhaser struct{}

func (p PurchaseRequisitionTransactionPhaser) PhaseSingleDoc(msg string) (*models.PurchaseRequisitionTransactionPG, error) {
	doc := prModel.PurchaseRequisitionDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseRequisitionDoc Message: " + err.Error())
	}
	trx, err := p.PhasePurchaseRequisitionTransactionDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert PurchaseRequisitionDoc to Transaction: " + err.Error())
	}
	return trx, err
}

func (p PurchaseRequisitionTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.PurchaseRequisitionTransactionPG, error) {
	docs := []prModel.PurchaseRequisitionDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal PurchaseRequisitionDoc Message: " + err.Error())
	}
	docsList := make([]models.PurchaseRequisitionTransactionPG, len(docs))
	for i, doc := range docs {
		trx, err := p.PhasePurchaseRequisitionTransactionDoc(doc)
		if err != nil {
			return nil, errors.New("Error on Convert PurchaseRequisitionDoc to Transaction: " + err.Error())
		}
		docsList[i] = *trx
	}
	return &docsList, nil
}

func (p PurchaseRequisitionTransactionPhaser) PhasePurchaseRequisitionTransactionDoc(doc prModel.PurchaseRequisitionDoc) (*models.PurchaseRequisitionTransactionPG, error) {
	details := []models.PurchaseRequisitionDetailTransactionPG{}

	if doc.Details != nil {
		details = make([]models.PurchaseRequisitionDetailTransactionPG, len(*doc.Details))
		for i, detail := range *doc.Details {
			prDetail := models.PurchaseRequisitionDetailTransactionPG{
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
			details[i] = prDetail
		}
	}

	transaction := models.PurchaseRequisitionTransactionPG{
		TransactionPG: models.TransactionPG{
			GuidFixed: doc.GuidFixed,
			GuidRef:   doc.GuidRef,
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			TransFlag:      21, // PR = transflag 21
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
		RequesterCode:    doc.RequesterCode,
		RequesterName:    doc.RequesterName,
		DepartmentCode:   doc.DepartmentCode,
		DepartmentNames:  *pkgModels.DefaultArrayNameX(doc.DepartmentNames),
		Purpose:          doc.Purpose,
		BudgetCode:       doc.BudgetCode,
		BudgetAmount:     doc.BudgetAmount,
		Urgency:          doc.Urgency,
		ConversionStatus: doc.ConversionStatus,
		Items:            &details,
	}

	return &transaction, nil
}
