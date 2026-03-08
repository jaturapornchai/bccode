package api

import (
	"context"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	chartofaccountrepositories "smlcloudplatform/internal/vfgl/chartofaccount/repositories"
)

func (m *MigrationService) ImportChartOfAccount(charts []accountModel.ChartOfAccountDoc) error {

	chartRepo := chartofaccountrepositories.NewChartOfAccountRepository(m.mongoPersister)
	chartMQRepo := chartofaccountrepositories.NewChartOfAccountMQRepository(m.mqPersister)

	for _, chart := range charts {
		// t.logger.Infof("Process Chart %s:%s", charts[i].AccountCode, charts[i].AccountName)

		findAccount, err := chartRepo.FindByGuid(context.Background(), chart.ShopID, chart.AccountCode)
		if err != nil {
			//t.logger.Errorf("Error Find Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			return err
		}

		if findAccount.GuidFixed == "" {

			_, err := chartRepo.Create(context.Background(), chart)
			if err != nil {
				//t.logger.Errorf("Error Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
				return err
				// } else {
				// 	t.logger.Infof("Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			}

			err = chartMQRepo.Create(chart)
			if err != nil {
				return err
			}
		} else {
			m.logger.Infof("Account %s:%s:%s is Already", chart.ShopID, chart.AccountCode, chart.AccountName)
		}
	}

	return nil
}
