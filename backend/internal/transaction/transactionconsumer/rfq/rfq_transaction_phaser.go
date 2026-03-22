package rfq

import (
	"encoding/json"
	"errors"
	pkgModels "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/transaction/models"
	rfqModel "smlcloudplatform/internal/transaction/rfq/models"
)

type RFQTransactionPhaser struct{}

func (p RFQTransactionPhaser) PhaseSingleDoc(msg string) (*models.RFQTransactionPG, error) {
	doc := rfqModel.RFQDoc{}
	err := json.Unmarshal([]byte(msg), &doc)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal RFQDoc Message: " + err.Error())
	}
	trx, err := p.PhaseRFQTransactionDoc(doc)
	if err != nil {
		return nil, errors.New("Error on Convert RFQDoc to Transaction: " + err.Error())
	}
	return trx, err
}

func (p RFQTransactionPhaser) PhaseMultipleDoc(input string) (*[]models.RFQTransactionPG, error) {
	docs := []rfqModel.RFQDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, errors.New("Cannot Unmarshal RFQDoc Message: " + err.Error())
	}
	docsList := make([]models.RFQTransactionPG, len(docs))
	for i, doc := range docs {
		trx, err := p.PhaseRFQTransactionDoc(doc)
		if err != nil {
			return nil, errors.New("Error on Convert RFQDoc to Transaction: " + err.Error())
		}
		docsList[i] = *trx
	}
	return &docsList, nil
}

func (p RFQTransactionPhaser) PhaseRFQTransactionDoc(doc rfqModel.RFQDoc) (*models.RFQTransactionPG, error) {
	details := []models.RFQDetailTransactionPG{}

	if doc.Details != nil {
		details = make([]models.RFQDetailTransactionPG, len(*doc.Details))
		for i, detail := range *doc.Details {
			rfqDetail := models.RFQDetailTransactionPG{
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
			details[i] = rfqDetail
		}
	}

	transaction := models.RFQTransactionPG{
		TransactionPG: models.TransactionPG{
			GuidFixed: doc.GuidFixed,
			GuidRef:   doc.GuidRef,
			ShopIdentity: pkgModels.ShopIdentity{
				ShopID: doc.ShopID,
			},
			TransFlag:      22, // RFQ = transflag 22
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
		RefPRDocNo:       doc.RefPRDocNo,
		RefPRGuidFixed:   doc.RefPRGuidFixed,
		SelectedVendor:   doc.SelectedVendor,
		SelectionReason:  doc.SelectionReason,
		RefPODocNo:       doc.RefPODocNo,
		RefPOGuidFixed:   doc.RefPOGuidFixed,
		MinVendors:       doc.MinVendors,
		ConversionStatus: doc.ConversionStatus,
		Items:            &details,
	}

	return &transaction, nil
}
