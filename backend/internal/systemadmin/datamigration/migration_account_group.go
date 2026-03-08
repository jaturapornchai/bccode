package datamigration

import (
	"context"
	"encoding/json"
	accountGroupModels "smlcloudplatform/internal/vfgl/accountgroup/models"
)

func (m *MigrationService) InitCenterAccountGroup() error {

	//chartSvc := chartOfAccountService.NewChartOfAccountHttpService()

	//m.chartService.SaveInBatch()
	accountGroups := masterAccountGroup()

	for _, accountGroup := range *accountGroups {
		// t.logger.Infof("Process Chart %s:%s", charts[i].AccountCode, charts[i].AccountName)

		findAccount, err := m.accountGroupRepo.FindByGuid(context.TODO(), accountGroup.ShopID, accountGroup.Code)
		if err != nil {
			m.logger.Errorf("Error Find Account Group %s:%s", accountGroup.Code, accountGroup.Name1)
			//return err
		}

		if findAccount.GuidFixed == "" {

			_, err := m.accountGroupRepo.Create(context.Background(), accountGroup)
			if err != nil {
				m.logger.Errorf("Error Create Account %s:%s", accountGroup.Code, accountGroup.Name1)
				//return err
				// } else {
				// 	t.logger.Infof("Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			}

		} else {
			m.logger.Infof("Account %s:%s:%s is Already", accountGroup.ShopID, accountGroup.Code, accountGroup.Name1)
		}
	}

	return nil
}

func masterAccountGroup() *[]accountGroupModels.AccountGroupDoc {

	docReq := &[]accountGroupModels.AccountGroupDoc{}
	data := `[
		{
			"shopid": "999999999",
			"guidfixed": "1",
			"code": "1",
			"name1": "เงินทุนที่ได้การสนับสนุนจากรัฐบาล (เงินล้าน)",
			"iscentergroup" : true
		},
		{
			"shopid": "999999999",
			"guidfixed": "2",
			"code": "2",
			"name1": "เงินทุนจาก เงินออม เงินสัจจะ และ เงินหุ้น",
			"iscentergroup" : true
		},
		{
			"shopid": "999999999",
			"guidfixed": "3",
			"code": "3",
			"name1": "เงินทุนจากการกู้ยืมธนาคารพาณิชย์",
			"iscentergroup" : true
		},
		{
			"shopid": "999999999",
			"guidfixed": "4",
			"code": "4",
			"name1": "เงินทุนจากกิจกรรมสถาบันการเงินชุมชนและเงินอื่นๆ",
			"iscentergroup" : true
		},
		{
			"shopid": "999999999",
			"guidfixed": "5",
			"code": "5",
			"name1": "ร้านค้าชุมชนตามโครงการประชารัฐ",
			"iscentergroup" : true
		}
	
	]`

	_ = json.Unmarshal([]byte(data), &docReq)

	return docReq
}
