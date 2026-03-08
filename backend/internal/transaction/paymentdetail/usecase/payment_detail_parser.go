package usecase

import (
	"encoding/json"
	"regexp"
	transmodels "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/paymentdetail/models"
)

func ParseTransactionToPaymentDetail(other transmodels.TransactionMessageQueue) ([]models.TransactionPaymentDetail, error) {

	tempDetails := []models.TransactionPaymentDetail{}
	rawDetail := other.PaymentDetailRaw

	if rawDetail == "" || rawDetail == "[]" || rawDetail == "{}" || rawDetail == "null" {
		return []models.TransactionPaymentDetail{}, nil
	}

	r := regexp.MustCompile(`\[\s*{.*?}\s*\]`)
	matches := r.FindStringSubmatch(rawDetail)

	if len(matches) == 0 {
		tempDetail := models.TransactionPaymentDetail{}
		err := json.Unmarshal([]byte(rawDetail), &tempDetail)

		if err != nil {
			return []models.TransactionPaymentDetail{}, err
		}

		tempDetail.ShopID = other.ShopID
		tempDetail.DocNo = other.DocNo

		// switch field
		tempDetail.PaymentType = tempDetail.TransFlag
		tempDetail.TransFlag = other.TransFlag

		tempDetails = append(tempDetails, tempDetail)
	} else {
		err := json.Unmarshal([]byte(rawDetail), &tempDetails)

		if err != nil {
			return []models.TransactionPaymentDetail{}, err
		}

		for i := range tempDetails {
			tempDetails[i].ShopID = other.ShopID
			tempDetails[i].DocNo = other.DocNo

			// switch field
			tempDetails[i].PaymentType = tempDetails[i].TransFlag
			tempDetails[i].TransFlag = other.TransFlag
		}
	}

	return tempDetails, nil
}
