package services

import (
	"fmt"
	common "smlcloudplatform/internal/models"
	trans_models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	"time"
)

type SaleInvocieExport struct{}

func (exp SaleInvocieExport) ParseCSV(languageCode string, data models.SaleInvoiceInfo) [][]string {

	results := [][]string{}

	if data.Details == nil {
		return results
	}

	for _, value := range *data.Details {

		tempDetail := exp.ParseDetailString(languageCode, data.DocNo, data.DocDatetime, value)

		results = append(results, tempDetail)
	}

	return results
}

func (exp SaleInvocieExport) ParseDetailString(languageCode string, docNo string, docDate time.Time, detail trans_models.Detail) []string {

	productName := exp.GetName(detail.ItemNames, languageCode)
	unitName := exp.GetName(detail.UnitNames, languageCode)

	qty := fmt.Sprintf("%.2f", detail.Qty)
	price := fmt.Sprintf("%.2f", detail.Price)
	discountAmount := fmt.Sprintf("%.2f", detail.DiscountAmount)
	sumAmount := fmt.Sprintf("%.2f", detail.SumAmount)

	dateLayout := "2006-01-02"
	docDateTxt := docDate.Format(dateLayout)

	return []string{docDateTxt, docNo, detail.Barcode, productName, detail.UnitCode, unitName, qty, price, discountAmount, sumAmount}
}

func (SaleInvocieExport) GetName(names *[]common.NameX, langCode string) string {
	if names == nil {
		return ""
	}

	for _, name := range *names {
		if *name.Code == langCode {
			return *name.Name
		}
	}

	return ""
}
