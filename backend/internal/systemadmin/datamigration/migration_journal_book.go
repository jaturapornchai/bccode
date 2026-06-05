package datamigration

import (
	"context"
	"encoding/json"
	journalBookModels "smlcloudplatform/internal/vfgl/journalbook/models"
)

func (m *MigrationService) InitJournalBookCenter() error {
	books := JournalBookCenter()
	for _, book := range *books {
		findBook, err := m.journalBookRepo.FindByGuid(context.TODO(), book.HoldingCode, book.GuidFixed)
		if err != nil {
			m.logger.Errorf("Error Find Account Group %s:%s", book.Code, book.Name1)
		}

		if findBook.GuidFixed == "" {
			_, err = m.journalBookRepo.Create(context.Background(), book)
			if err != nil {
				m.logger.Errorf("Error Create Account %s:%s", book.Code, book.Name1)
			}
		} else {
			m.logger.Infof("Account %s:%s:%s is Already", book.HoldingCode, book.Code, book.Name1)
		}
	}
	return nil
}

func JournalBookCenter() *[]journalBookModels.JournalBookDoc {
	books := &[]journalBookModels.JournalBookDoc{}
	jsonStr := `[
		{
		  "holdingcode": "999999999",
		  "guidfixed": "1",
		  "code": "1",
		  "name1": "สมุดรายวันทั่วไป",
		  "iscenterbook": true
		},
		{
		  "holdingcode": "999999999",
		  "guidfixed": "2",
		  "code": "2",
		  "name1": "สมุดเงินสดรับ",
		  "iscenterbook": true
		},
		{
		  "holdingcode": "999999999",
		  "guidfixed": "3",
		  "code": "3",
		  "name1": "สมุดเงินสดจ่าย",
		  "iscenterbook": true
		},
		{
		  "holdingcode": "999999999",
		  "guidfixed": "4",
		  "code": "4",
		  "name1": "สมุดรายวันขาย",
		  "iscenterbook": true
		},
		{
		  "holdingcode": "999999999",
		  "guidfixed": "5",
		  "code": "5",
		  "name1": "สมุดรายวันซื้อ",
		  "iscenterbook": true
		}
	  ]`

	_ = json.Unmarshal([]byte(jsonStr), &books)

	return books
}
