package services

import (
	"encoding/json"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"

	"gorm.io/gorm"
)

type IJournalConsumeService interface {
	// Create(doc models.JournalDoc) error
	// Update(shopID string, docNo string, doc models.JournalDoc) error
	// SaveInBatch(docList []models.JournalDoc) error
	Delete(shopID string, guid string) error
	UpSert(shopID string, docNo string, doc models.JournalDoc) (*models.JournalPg, error)
}

type JournalConsumeService struct {
	repo repositories.IJournalPgRepository
}

func NewJournalConsumeService(repo repositories.IJournalPgRepository) IJournalConsumeService {
	return &JournalConsumeService{
		repo: repo,
	}
}

func (svc *JournalConsumeService) Create(doc models.JournalDoc) (*models.JournalPg, error) {
	pgDoc := models.JournalPg{}

	tmpJsonDoc, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal([]byte(tmpJsonDoc), &pgDoc)
	if err != nil {
		return nil, err
	}

	err = svc.repo.Create(pgDoc)

	if err != nil {
		return nil, err
	}
	return &pgDoc, nil
}

func (svc *JournalConsumeService) Update(shopID string, docNo string, doc models.JournalDoc) error {
	pgDoc := models.JournalPg{}

	tmpJsonDoc, err := json.Marshal(doc)
	if err != nil {
		return err
	}

	err = json.Unmarshal([]byte(tmpJsonDoc), &pgDoc)
	if err != nil {
		return err
	}

	err = svc.repo.Update(shopID, docNo, pgDoc)

	if err != nil {
		return err
	}
	return nil
}

func (svc *JournalConsumeService) Delete(shopID string, docNo string) error {
	err := svc.repo.Delete(shopID, docNo)

	if err != nil {
		return err
	}
	return nil
}

func (svc *JournalConsumeService) SaveInBatch(docList []models.JournalDoc) error {
	pgDocList := []models.JournalPg{}

	for _, doc := range docList {
		tmpJsonDoc, err := json.Marshal(doc)
		if err != nil {
			return err
		}
		tmpDoc := models.JournalPg{}
		err = json.Unmarshal([]byte(tmpJsonDoc), &tmpDoc)
		if err != nil {
			return err
		}
		pgDocList = append(pgDocList, tmpDoc)
	}

	err := svc.repo.CreateInBatch(pgDocList)
	if err != nil {
		return err
	}

	return nil
}

func (svc *JournalConsumeService) UpSert(shopID string, docNo string, doc models.JournalDoc) (*models.JournalPg, error) {
	docPg := models.JournalPg{}

	tmpJsonDoc, err := json.Marshal(doc)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(tmpJsonDoc), &docPg)
	if err != nil {
		return nil, err
	}

	// Initialize slices if they're nil
	if docPg.Vats == nil {
		docPg.Vats = &[]models.JournalVatPg{}
	}
	if docPg.Taxes == nil {
		docPg.Taxes = &[]models.JournalTaxPg{}
	}
	if docPg.AccountBook == nil {
		docPg.AccountBook = &[]models.JournalDetailPg{}
	}

	data, err := svc.repo.Get(shopID, docNo)
	if err == gorm.ErrRecordNotFound {
		data, err = svc.Create(doc)
		if err != nil {
			return nil, err
		}
	} else if data != nil {
		data.JournalBody = doc.JournalBody

		// แก้ไขส่วนการตรวจสอบและอัปเดต AccountBook
		if data.AccountBook != nil && docPg.AccountBook != nil {
			for _, tmp := range *data.AccountBook {
				for i, detail := range *docPg.AccountBook {
					if tmp.AccountCode == detail.AccountCode && tmp.AccountName == detail.AccountName && tmp.CreditAmount == detail.CreditAmount && tmp.DebitAmount == detail.DebitAmount {
						(*docPg.AccountBook)[i].ID = tmp.ID
					}
				}
			}
		}

		// แก้ไขส่วนการตรวจสอบและอัปเดต Vats
		if data.Vats != nil && docPg.Vats != nil {
			for _, tmp := range *data.Vats {
				for i, vat := range *docPg.Vats {
					if tmp.VatDocNo == vat.VatDocNo {
						(*docPg.Vats)[i].ID = tmp.ID
						break
					}
				}
			}
		}

		// แก้ไขส่วนการตรวจสอบและอัปเดต Taxes
		if data.Taxes != nil && docPg.Taxes != nil {
			for _, tmp := range *data.Taxes {
				for i, tax := range *docPg.Taxes {
					// ตรวจสอบทั้ง TaxDocNo และฟิลด์อื่นๆ ที่เกี่ยวข้องเพื่อระบุรายการเดียวกัน
					if tmp.TaxDocNo == tax.TaxDocNo {
						(*docPg.Taxes)[i].ID = tmp.ID

						// If we need to preserve any existing Details, we can do that here
						// This is likely not needed but included for completeness
						if tmp.Details != nil && tax.Details == nil {
							(*docPg.Taxes)[i].Details = tmp.Details
						}

						break
					}
				}
			}
		}

		if err = svc.repo.Update(shopID, doc.DocNo, docPg); err != nil {
			return nil, err
		}
	}

	return &docPg, nil
}
