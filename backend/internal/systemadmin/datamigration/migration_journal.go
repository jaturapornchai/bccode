package datamigration

import (
	"context"
	"smlcloudplatform/internal/utils"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	journalRepo "smlcloudplatform/internal/vfgl/journal/repositories"
)

func (m *MigrationService) ImportJournal(journals []journalModels.JournalDoc) error {
	journalMongoRepo := journalRepo.NewJournalRepository(m.mongoPersister)
	journalMQRepo := journalRepo.NewJournalMqRepository(m.mqPersister)

	// for _, journal := range journals {
	// 	journal.GuidFixed = utils.NewGUID()
	// 	_, err := journalpgRepo.Create(journal)
	// 	if err != nil {
	// 		return err
	// 	}

	// 	err = journalMQRepo.Create(journal)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	for index := range journals {
		journals[index].GuidFixed = utils.NewGUID()
	}

	err := journalMongoRepo.CreateInBatch(context.Background(), journals)

	if err != nil {
		return err
	}

	journalMQRepo.CreateInBatch(journals)

	return nil
}
