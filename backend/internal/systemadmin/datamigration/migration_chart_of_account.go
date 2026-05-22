package datamigration

import (
	"context"
	"encoding/json"
	chartOfAccountKafkaConfig "smlcloudplatform/internal/vfgl/chartofaccount/config"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
)

func (m *MigrationService) ImportChartOfAccount(charts []accountModel.ChartOfAccountDoc) error {

	//chartSvc := chartOfAccountService.NewChartOfAccountHttpService()

	//m.chartService.SaveInBatch()

	for _, chart := range charts {
		// t.logger.Infof("Process Chart %s:%s", charts[i].AccountCode, charts[i].AccountName)

		findAccount, err := m.chartRepo.FindByGuid(context.TODO(), chart.ShopID, chart.AccountCode)
		if err != nil {
			m.logger.Errorf("Error Find Account %s:%s", chart.AccountCode, chart.AccountName)
			//return err
		}

		if findAccount.GuidFixed == "" {

			_, err := m.chartRepo.Create(context.Background(), chart)
			if err != nil {
				m.logger.Errorf("Error Create Account %s:%s", chart.AccountCode, chart.AccountName)
				//return err
				// } else {
				// 	t.logger.Infof("Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			}

			err = m.chartMQRepo.Create(chart)
			if err != nil {
				chartKafkaConfig := chartOfAccountKafkaConfig.ChartOfAccountMessageQueueConfig{}
				m.logger.Errorf("Error Create Message in Topic[%s]for consume %s, %s:%s", chartKafkaConfig.TopicCreated(), chart.AccountCode, chart.AccountName)
			}
		} else {
			m.logger.Infof("Account %s:%s:%s is Already", chart.ShopID, chart.AccountCode, chart.AccountName)
		}
	}

	return nil
}

func (m *MigrationService) ResyncChartOfAccount(charts []accountModel.ChartOfAccountDoc) error {

	for _, chart := range charts {
		// t.logger.Infof("Process Chart %s:%s", charts[i].AccountCode, charts[i].AccountName)

		findAccount, err := m.chartRepo.FindByGuid(context.TODO(), chart.ShopID, chart.AccountCode)
		if err != nil {
			m.logger.Errorf("Error Find Account %s:%s", chart.ShopID, chart.AccountCode)
			//return err
		}
		if findAccount.GuidFixed != "" {
			err = m.chartMQRepo.Update(chart)
			if err != nil {
				chartKafkaConfig := chartOfAccountKafkaConfig.ChartOfAccountMessageQueueConfig{}
				m.logger.Errorf("Error Create Message in Topic[%s]for consume %s, %s:%s", chartKafkaConfig.TopicUpdated(), chart.AccountCode, chart.AccountName)
			}
		}
	}
	return nil
}

func (m *MigrationService) InitialChartOfAccountCenter() error {

	charts := CenterChartOfAccount()
	err := m.ImportChartOfAccount(charts)
	return err
}

func CenterChartOfAccount() []accountModel.ChartOfAccountDoc {

	docs := []accountModel.ChartOfAccountDoc{}
	jsonStr := `[
		{
		  "shopid": "999999999",
		  "accountcode": "10000",
		  "guid_fixed": "10000",
		  "accountname": "**สินทรัพย์**",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 1,
		  "consolidateaccountcode": "",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "11000",
		  "guid_fixed": "11000",
		  "accountname": "*เงินสด*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "11010",
		  "guid_fixed": "11010",
		  "accountname": "เงินสด - เงินกองทุน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "11000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "11020",
		  "guid_fixed": "11020",
		  "accountname": "เงินสด - ร้านค้าประชารัฐ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "11000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "11030",
		  "guid_fixed": "11030",
		  "accountname": "เงินสด - โครงการประชารัฐ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "11000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12000",
		  "guid_fixed": "12000",
		  "accountname": "*เงินฝากธนาคาร*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12100",
		  "guid_fixed": "12100",
		  "accountname": "*เงินฝากธนาคาร บัญชี 1 (เงินล้าน)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12110",
		  "guid_fixed": "12110",
		  "accountname": "เงินฝากธนาคาร บัญชี 1 (เงินล้าน) ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12100",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12120",
		  "guid_fixed": "12120",
		  "accountname": "เงินฝากธนาคาร บัญชี 1 (เงินล้าน) ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12100",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12130",
		  "guid_fixed": "12130",
		  "accountname": "เงินฝากธนาคาร บัญชี 1 (เงินล้าน) ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12100",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12200",
		  "guid_fixed": "12200",
		  "accountname": "*เงินฝากธนาคาร บัญชี 2 (เงินออม)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12210",
		  "guid_fixed": "12210",
		  "accountname": "เงินฝากธนาคาร บัญชี 2 (เงินออม) ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12200",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12220",
		  "guid_fixed": "12220",
		  "accountname": "เงินฝากธนาคาร บัญชี 2 (เงินออม) ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12200",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12230",
		  "guid_fixed": "12230",
		  "accountname": "เงินฝากธนาคาร บัญชี 2 (เงินออม) ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12200",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12300",
		  "guid_fixed": "12300",
		  "accountname": "*เงินฝากธนาคาร บัญชี 3 (เงินสัจจะ)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12310",
		  "guid_fixed": "12310",
		  "accountname": "เงินฝากธนาคาร บัญชี 3 (เงินสัจจะ)  ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12300",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12320",
		  "guid_fixed": "12320",
		  "accountname": "เงินฝากธนาคาร บัญชี 3 (เงินสัจจะ)  ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12300",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12330",
		  "guid_fixed": "12330",
		  "accountname": "เงินฝากธนาคาร บัญชี 3 (เงินสัจจะ)  ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12300",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12400",
		  "guid_fixed": "12400",
		  "accountname": "*เงินฝากธนาคาร บัญชี 4 (สวัสดิการ)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12410",
		  "guid_fixed": "12410",
		  "accountname": "เงินฝากธนาคาร บัญชี 4 (สวัสดิการ)   ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12400",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12420",
		  "guid_fixed": "12420",
		  "accountname": "เงินฝากธนาคาร บัญชี 4 (สวัสดิการ)   ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12400",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12430",
		  "guid_fixed": "12430",
		  "accountname": "เงินฝากธนาคาร บัญชี 4 (สวัสดิการ)   ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12400",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12500",
		  "guid_fixed": "12500",
		  "accountname": "*เงินฝากธนาคาร บัญชี 5 (ร้านค้าชุมชน)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12510",
		  "guid_fixed": "12510",
		  "accountname": "เงินฝากธนาคาร บัญชี 5 (ร้านค้าชุมชน)   ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12500",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12520",
		  "guid_fixed": "12520",
		  "accountname": "เงินฝากธนาคาร บัญชี 5 (ร้านค้าชุมชน) )   ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12500",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12530",
		  "guid_fixed": "12530",
		  "accountname": "เงินฝากธนาคาร บัญชี 5 (ร้านค้าชุมชน)    ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12500",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12600",
		  "guid_fixed": "12600",
		  "accountname": "*เงินฝากธนาคาร บัญชี 6 (โครงการประชารัฐ)*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12610",
		  "guid_fixed": "12610",
		  "accountname": "เงินฝากธนาคาร บัญชี 6 (โครงการประชารัฐ)   ธนาคารออมสิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12660",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12620",
		  "guid_fixed": "12620",
		  "accountname": "เงินฝากธนาคาร บัญชี 6 (โครงการประชารัฐ)   ธนาคารเพื่อการเกษตรและสหกรณ์ (ธกส)",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12660",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12630",
		  "guid_fixed": "12630",
		  "accountname": "เงินฝากธนาคาร บัญชี 6 (โครงการประชารัฐ)   ธนาคารกรุงไทย",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12660",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12700",
		  "guid_fixed": "12700",
		  "accountname": "*เงินฝากธนาคารอื่น*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "12000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "12710",
		  "guid_fixed": "12710",
		  "accountname": "เงินฝากธนาคาร",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "12700",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13000",
		  "guid_fixed": "13000",
		  "accountname": "*ลูกหนี้เงินกู้ยืม*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13010",
		  "guid_fixed": "13010",
		  "accountname": "ลูกหนี้เงินกู้ - สามัญ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "13000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13020",
		  "guid_fixed": "13020",
		  "accountname": "ลูกหนี้เงินกู้ - ฉุกเฉิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "13000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13030",
		  "guid_fixed": "13030",
		  "accountname": "ลูกหนี้เงินกู้ - เงินกู้อีน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "13000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13100",
		  "guid_fixed": "13100",
		  "accountname": "*ค่าเผื่อหนี้สงสัยจะสูญ*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 3,
		  "consolidateaccountcode": "13000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "13110",
		  "guid_fixed": "13110",
		  "accountname": "ค่าเผื่อหนี้สงสัยจะสูญ-ลูกหนี้เงินกู้",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 4,
		  "consolidateaccountcode": "13100",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "14000",
		  "guid_fixed": "14000",
		  "accountname": "*ลูกหนี้การค้า*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "14010",
		  "guid_fixed": "14010",
		  "accountname": "ลูกหนี้การค้า",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "14000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "14020",
		  "guid_fixed": "14020",
		  "accountname": "ลูกหนี้อื่น",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "14000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "15000",
		  "guid_fixed": "15000",
		  "accountname": "*สินค้าคงเหลือ*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "15010",
		  "guid_fixed": "15010",
		  "accountname": "สินค้าสำเร็จรูป",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "15000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "15020",
		  "guid_fixed": "15020",
		  "accountname": "วัตถุดิบเพื่อการผลิต",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "15000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "16000",
		  "guid_fixed": "16000",
		  "accountname": "*สินทรัพย์หมุนเวียนอื่น ๆ*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "16010",
		  "guid_fixed": "16010",
		  "accountname": "เงินกันสำรองหนี้สูญ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "16000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "16020",
		  "guid_fixed": "16020",
		  "accountname": "รายได้ค้างรับ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "16000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "16030",
		  "guid_fixed": "16030",
		  "accountname": "สินทรัพย์หมุนเวียนอื่น ๆ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "16000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "16040",
		  "guid_fixed": "16040",
		  "accountname": "ภาษีซื้อ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "16000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "17000",
		  "guid_fixed": "17000",
		  "accountname": "*สินทรัพย์ไม่หมุนเวียน*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "17010",
		  "guid_fixed": "17010",
		  "accountname": "เงินลงทุนระยะยาว",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "17000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "17020",
		  "guid_fixed": "17020",
		  "accountname": "เงินลงทุน-ฉลาก",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "17010",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "17030",
		  "guid_fixed": "17030",
		  "accountname": "เงินลงทุน-อื่น",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "17010",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18000",
		  "guid_fixed": "18000",
		  "accountname": "*ที่ดิน อาคาร และอุปกรณ์*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18010",
		  "guid_fixed": "18010",
		  "accountname": "ที่ดิน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18020",
		  "guid_fixed": "18020",
		  "accountname": "อาคารสำนักงาน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18025",
		  "guid_fixed": "18025",
		  "accountname": "ค่าเสื่อมราคาสะสม - อาคารสำนักงาน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18030",
		  "guid_fixed": "18030",
		  "accountname": "อุปกรณ์สำนักงาน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18035",
		  "guid_fixed": "18035",
		  "accountname": "ค่าเสื่อมราคาสะสม - อุปกรณ์สำนักงาน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18040",
		  "guid_fixed": "18040",
		  "accountname": "ครุภัณฑ์",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18045",
		  "guid_fixed": "18045",
		  "accountname": "ค่าเสื่อมราคาสะสม - ครุภัณฑ์",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18050",
		  "guid_fixed": "18050",
		  "accountname": "ยานพาหนะ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "18055",
		  "guid_fixed": "18055",
		  "accountname": "ค่าเสื่อมราคาสะสม - ยานพาหนะ",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "18000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "19000",
		  "guid_fixed": "19000",
		  "accountname": "*ทรัพย์สินไม่หมุนเวียนอื่น*",
		  "accountcategory": 1,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "19100",
		  "guid_fixed": "19100",
		  "accountname": "ทรัพย์สินไม่มีตัวตน",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "19000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "19200",
		  "guid_fixed": "19200",
		  "accountname": "ทรัพย์สินไม่หมุนเวียนอื่น",
		  "accountcategory": 1,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "10000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "20000",
		  "guid_fixed": "20000",
		  "accountname": "**หนี้สิน**",
		  "accountcategory": 2,
		  "accountgroup": "0",
		  "accountlevel": 1,
		  "consolidateaccountcode": "",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "21000",
		  "guid_fixed": "21000",
		  "accountname": "*เจ้าหนี้ - เงินรับฝาก*",
		  "accountcategory": 2,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "20000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "21010",
		  "guid_fixed": "21010",
		  "accountname": "เจ้าหนี้ - เงินรับฝากออมทรัพย์/เผื่อเรียก",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "21000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "21020",
		  "guid_fixed": "21020",
		  "accountname": "เจ้าหนี้ - เงินรับฝากออมทรัพย์/เผื่อเรียก พิเศษ",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "21000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "21030",
		  "guid_fixed": "21030",
		  "accountname": "เจ้าหนี้ - เงินรับฝากประจำ  6  เดือน",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "21000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "21040",
		  "guid_fixed": "21040",
		  "accountname": "เจ้าหนี้ - เงินรับฝากประจำ  12  เดือน",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "21000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "22000",
		  "guid_fixed": "22000",
		  "accountname": "*เงินกู้ยืมระยะสั้น*",
		  "accountcategory": 2,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "20000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "22010",
		  "guid_fixed": "22010",
		  "accountname": "เจ้าหนี้ - เงินกู้เบิกเกินบัญชี (OD)",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "22000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "22020",
		  "guid_fixed": "22020",
		  "accountname": "เจ้าหนี้ - เงินกู้ระยะสั้น",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "22000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "22030",
		  "guid_fixed": "22030",
		  "accountname": "เจ้าหนี้ - สถาบันการเงินอื่น",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "22000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23000",
		  "guid_fixed": "23000",
		  "accountname": "*หนี้สินหมุนเวียนอื่น*",
		  "accountcategory": 2,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "20000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23010",
		  "guid_fixed": "23010",
		  "accountname": "เจ้าหนี้การค้า",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23020",
		  "guid_fixed": "23020",
		  "accountname": "ภาษีขาย รอนำส่ง",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23030",
		  "guid_fixed": "23030",
		  "accountname": "เงินได้รับล่วงหน้า",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23040",
		  "guid_fixed": "23040",
		  "accountname": "ค่าใช้จ่ายค้างจ่าย",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23050",
		  "guid_fixed": "23050",
		  "accountname": "ภาษีหัก ณ ที่จ่าย ค้างจ่าย",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23060",
		  "guid_fixed": "23060",
		  "accountname": "ภาษีเงินได้ค้างจ่าย",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "23070",
		  "guid_fixed": "23070",
		  "accountname": "เจ้าหนี้อื่นๆ",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "23000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "24000",
		  "guid_fixed": "24000",
		  "accountname": "*หนี้สินระยะยาว*",
		  "accountcategory": 2,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "20000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "24010",
		  "guid_fixed": "24010",
		  "accountname": "เงินกู้ยืมจากธนาคาร ",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "24000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "24020",
		  "guid_fixed": "24020",
		  "accountname": "เงินกู้ยืมที่มีอายุเกิน 1 ปี",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "24000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "24030",
		  "guid_fixed": "24030",
		  "accountname": "เงินกู้ยืมระยะยาวอื่น",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "24000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "24040",
		  "guid_fixed": "24040",
		  "accountname": "หนี้สินไม่หมุนเวียนอื่น",
		  "accountcategory": 2,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "24000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "30000",
		  "guid_fixed": "30000",
		  "accountname": "**ทุนและส่วนของผู้ถือหุ้น**",
		  "accountcategory": 3,
		  "accountgroup": "0",
		  "accountlevel": 1,
		  "consolidateaccountcode": "",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "31000",
		  "guid_fixed": "31000",
		  "accountname": "*ทุน*",
		  "accountcategory": 3,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "30000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "31010",
		  "guid_fixed": "31010",
		  "accountname": "ทุน - หุ้นสมาชิก",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "31000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "31020",
		  "guid_fixed": "31020",
		  "accountname": "ทุน - เงินออมสัจจะ",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "31000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32000",
		  "guid_fixed": "32000",
		  "accountname": "*ทุน - เงินจัดสรรจากรัฐบาล*",
		  "accountcategory": 3,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "30000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32010",
		  "guid_fixed": "32010",
		  "accountname": "ทุน - เงินล้าน",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32020",
		  "guid_fixed": "32020",
		  "accountname": "ทุน - โครงการ 3A",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32030",
		  "guid_fixed": "32030",
		  "accountname": "ทุน - เงินเพิ่มทุนระยะ 2",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32040",
		  "guid_fixed": "32040",
		  "accountname": "ทุน - เงินเพิ่มทุนระยะ 3",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32050",
		  "guid_fixed": "32050",
		  "accountname": "ทุน - โครงการประชารัฐ",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "32060",
		  "guid_fixed": "32060",
		  "accountname": "ทุน - อื่น",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "32000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "33000",
		  "guid_fixed": "33000",
		  "accountname": "*กำไรสะสม (ขาดทุน) สะสม*",
		  "accountcategory": 3,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "30000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "33010",
		  "guid_fixed": "33010",
		  "accountname": "กำไร (ขาดทุน) สะสม",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "33000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "33020",
		  "guid_fixed": "33020",
		  "accountname": "กำไร (ขาดทุน)",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "33000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34000",
		  "guid_fixed": "34000",
		  "accountname": "*กำไรที่จัดสรร*",
		  "accountcategory": 3,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "30000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34100",
		  "guid_fixed": "34100",
		  "accountname": "ทุนสำรองตามกฏหมาย",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34200",
		  "guid_fixed": "34200",
		  "accountname": "เงินสมทบกองทุน",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34300",
		  "guid_fixed": "34300",
		  "accountname": "เงินเฉลี่ยคืน",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34400",
		  "guid_fixed": "34400",
		  "accountname": "เงินปันผล",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34500",
		  "guid_fixed": "34500",
		  "accountname": "ค่าตอบแทนคณะกรรมการ",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34510",
		  "guid_fixed": "34510",
		  "accountname": "ทุนสาธารณะประโยชน์",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34520",
		  "guid_fixed": "34520",
		  "accountname": "เงินประกันความเสี่ยง",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34530",
		  "guid_fixed": "34530",
		  "accountname": "เงินสวัสดิการกองทุน",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34540",
		  "guid_fixed": "34540",
		  "accountname": "เงินสมทบเพื่อการศึกษา",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34550",
		  "guid_fixed": "34550",
		  "accountname": "ค่าดำเนินงาน/ค่าบริหารจัดการ",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "34560",
		  "guid_fixed": "34560",
		  "accountname": "เงินฌาปณกิจ",
		  "accountcategory": 3,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "34000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "40000",
		  "guid_fixed": "40000",
		  "accountname": "**รายได้**",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 1,
		  "consolidateaccountcode": "",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "41000",
		  "guid_fixed": "41000",
		  "accountname": "*รายได้ดอกเบี้ย-จากการปล่อยกู้*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "41010",
		  "guid_fixed": "41010",
		  "accountname": "รายได้ดอกเบี้ยเงินกู้ - สามัญ",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "41000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "41020",
		  "guid_fixed": "41020",
		  "accountname": "รายได้ดอกเบี้ยเงินกู้ - ฉุกเฉิน",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "41000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "41030",
		  "guid_fixed": "41030",
		  "accountname": "รายได้ดอกเบี้ยเงินกู้ - อื่น",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "41000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "42000",
		  "guid_fixed": "42000",
		  "accountname": "*รายได้ค่าปรับเงินกู้*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "42010",
		  "guid_fixed": "42010",
		  "accountname": "รายได้ค่าปรับเงินกู้ - สามัญ",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "42000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "42020",
		  "guid_fixed": "42020",
		  "accountname": "รายได้ค่าปรับเงินกู้ - ฉุกเฉิน",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "42000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "42030",
		  "guid_fixed": "42030",
		  "accountname": "รายได้ค่าปรับเงินกู้ - อื่น",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "42000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "43000",
		  "guid_fixed": "43000",
		  "accountname": "*รายได้ค่าธรรมเนียม*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "43010",
		  "guid_fixed": "43010",
		  "accountname": "รายได้ค่าธรรมเนียม-แรกเข้า",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "43000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "43020",
		  "guid_fixed": "43020",
		  "accountname": "รายได้ค่าธรรมเนียม-ขอกู้",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "43000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "43030",
		  "guid_fixed": "43030",
		  "accountname": "รายได้ค่าธรรมเนียม-ติดตามหนี้",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "43000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "43040",
		  "guid_fixed": "43040",
		  "accountname": "รายได้ค่าธรรมเนียมอื่น",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "43000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "44000",
		  "guid_fixed": "44000",
		  "accountname": "*รายได้จากการขายและให้บริการ*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "44010",
		  "guid_fixed": "44010",
		  "accountname": "รายได้จากการขายสินค้า",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "44000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "44020",
		  "guid_fixed": "44020",
		  "accountname": "รายได้จากการให้บริการ",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "44000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "45000",
		  "guid_fixed": "45000",
		  "accountname": "*รายได้ดอกเบี้ยธนาคารและผลประโยชน์อื่น*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "45010",
		  "guid_fixed": "45010",
		  "accountname": "รายได้ดอกเบี้ยเงินฝากธนาคาร",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "45000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "45020",
		  "guid_fixed": "45020",
		  "accountname": "รายได้ดอกเบี้ยอื่น",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "45000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "46000",
		  "guid_fixed": "46000",
		  "accountname": "*รายได้อื่น ๆ*",
		  "accountcategory": 4,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "40000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "46010",
		  "guid_fixed": "46010",
		  "accountname": "รายได้เบ็ดเตล็ด",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "46000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "46020",
		  "guid_fixed": "46020",
		  "accountname": "รายได้เงินรับบริจาค",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "46000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "46030",
		  "guid_fixed": "46030",
		  "accountname": "รายได้จากการขายทรัพย์สิน",
		  "accountcategory": 4,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "46000",
		  "accountbalancetype": 2
		},
		{
		  "shopid": "999999999",
		  "accountcode": "50000",
		  "guid_fixed": "50000",
		  "accountname": "**ค่าใช้จ่าย**",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 1,
		  "consolidateaccountcode": "",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "51000",
		  "guid_fixed": "51000",
		  "accountname": "*ดอกเบี้ยจ่าย-เงินรับฝาก*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "51010",
		  "guid_fixed": "51010",
		  "accountname": "ดอกเบี้ยจ่าย-เงินฝากออมทรัพย์",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "51000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "51020",
		  "guid_fixed": "51020",
		  "accountname": "ดอกเบี้ยจ่าย-เงินฝากประจำ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "51000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "51030",
		  "guid_fixed": "51030",
		  "accountname": "ดอกเบี้ยจ่าย-เงินฝากอื่น",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "51000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "52000",
		  "guid_fixed": "52000",
		  "accountname": "*ต้นทุนขายสินค้า*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "52010",
		  "guid_fixed": "52010",
		  "accountname": "ซื้อสินค้า",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "52000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "52020",
		  "guid_fixed": "52020",
		  "accountname": "ส่งคืนและส่วนลด",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "52000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "52030",
		  "guid_fixed": "52030",
		  "accountname": "ค่าขนส่ง",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "52000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "53000",
		  "guid_fixed": "53000",
		  "accountname": "*ค่าใช้จ่ายในการบริหาร*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "53010",
		  "guid_fixed": "53010",
		  "accountname": "เงินเดือนและค่าตอบแทน",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "53000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "53020",
		  "guid_fixed": "53020",
		  "accountname": "ค่าเบี้ยเลี้ยงกรรมการ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "53000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "53030",
		  "guid_fixed": "53030",
		  "accountname": "โบนัสและผลตอบแทนอื่น",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "53000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "53040",
		  "guid_fixed": "53040",
		  "accountname": "ค่ารับรอง",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "53000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54000",
		  "guid_fixed": "54000",
		  "accountname": "*ค่าใช้จ่ายในการดำเนินการ*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54010",
		  "guid_fixed": "54010",
		  "accountname": "ค่าเช่า",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54020",
		  "guid_fixed": "54020",
		  "accountname": "ค่าน้ำ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54030",
		  "guid_fixed": "54030",
		  "accountname": "ค่าไฟฟ้า",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54040",
		  "guid_fixed": "54040",
		  "accountname": "ค่าโทรศัพท์และอินเตอร์เน็ต",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54050",
		  "guid_fixed": "54050",
		  "accountname": "ค่าวัสดุอุปกรณ์สิ้นเปลือง",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54060",
		  "guid_fixed": "54060",
		  "accountname": "ค่าพาหนะเดินทาง",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54070",
		  "guid_fixed": "54070",
		  "accountname": "ค่าซ่อมบำรุง",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54080",
		  "guid_fixed": "54080",
		  "accountname": "ค่าใช้จ่ายในการศึกษาดูงาน",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "54090",
		  "guid_fixed": "54090",
		  "accountname": "ค่าใช้จ่ายเบ็ดเตล็ด",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "54000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55000",
		  "guid_fixed": "55000",
		  "accountname": "*ดอกเบี้ย/ค่าธรรมเนียม/ภาษี/อื่นๆ",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55010",
		  "guid_fixed": "55010",
		  "accountname": "ดอกเบี้ยจ่าย",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "55000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55020",
		  "guid_fixed": "55020",
		  "accountname": "ค่าธรรมเนียมอื่นๆ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "55000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55030",
		  "guid_fixed": "55030",
		  "accountname": "ค่าภาษีโรงเรือน/ภาษีป้าย/ภาษีอื่น",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "55000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55040",
		  "guid_fixed": "55040",
		  "accountname": "ขาดทุนจากการปิดบัญชี",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "55000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "55050",
		  "guid_fixed": "55050",
		  "accountname": "หนี้สงสัยจะสูญ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "55000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56000",
		  "guid_fixed": "56000",
		  "accountname": "*ค่าใช้จ่ายสวัสดิการ*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56010",
		  "guid_fixed": "56010",
		  "accountname": "สวัสดิการ - รักษาพยาบาล",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56020",
		  "guid_fixed": "56020",
		  "accountname": "สวัสดิการ - ผู้สูงอายุ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56030",
		  "guid_fixed": "56030",
		  "accountname": "สวัสดิการ - แรกเกิด/บุตร",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56040",
		  "guid_fixed": "56040",
		  "accountname": "สวัสดิการ - เสียชีวิต",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56050",
		  "guid_fixed": "56050",
		  "accountname": "สวัสดิการ - อื่น",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56060",
		  "guid_fixed": "56060",
		  "accountname": "จ่ายเงินปันผล",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "56070",
		  "guid_fixed": "56070",
		  "accountname": "จ่ายเงินเฉลี่ยคืน",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "56000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "57000",
		  "guid_fixed": "57000",
		  "accountname": "*ค่าเสื่อมราคา*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "57010",
		  "guid_fixed": "57010",
		  "accountname": "ค่าเสื่อมราคา - อาคาร",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "57000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "57020",
		  "guid_fixed": "57020",
		  "accountname": "ค่าเสื่อมราคา - อุปกรณ์",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "57000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "57030",
		  "guid_fixed": "57030",
		  "accountname": "ค่าเสื่อมราคา - ครุภัณฑ์",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "57000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "57040",
		  "guid_fixed": "57040",
		  "accountname": "ค่าเสื่อมราคา - ยานพาหนะ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "57000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "58000",
		  "guid_fixed": "58000",
		  "accountname": "*ค่าใช้จ่ายอื่น*",
		  "accountcategory": 5,
		  "accountgroup": "0",
		  "accountlevel": 2,
		  "consolidateaccountcode": "50000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "58010",
		  "guid_fixed": "58010",
		  "accountname": "ตัดหนี้สูญ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "58000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "58020",
		  "guid_fixed": "58020",
		  "accountname": "ค่าใช้จ่ายอื่นๆ",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "58000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		},
		{
		  "shopid": "999999999",
		  "accountcode": "59000",
		  "guid_fixed": "59000",
		  "accountname": "บัญชีพัก",
		  "accountcategory": 5,
		  "accountgroup": "1",
		  "accountlevel": 3,
		  "consolidateaccountcode": "58000",
		  "accountbalancetype": 1,
		  "iscenterchart": true
		}
	  ]
	  `
	_ = json.Unmarshal([]byte(jsonStr), &docs)

	return docs
}
