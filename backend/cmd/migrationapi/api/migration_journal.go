package api

import (
	"context"
	"smlcloudplatform/internal/utils"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	journalRepo "smlcloudplatform/internal/vfgl/journal/repositories"
)

func (m *MigrationService) ImportJournal(journals []journalModels.JournalDoc) error {
	journalpgRepo := journalRepo.NewJournalRepository(m.mongoPersister)
	journalMQRepo := journalRepo.NewJournalMqRepository(m.mqPersister)

	for _, journal := range journals {
		journal.GuidFixed = utils.NewGUID()
		_, err := journalpgRepo.Create(context.Background(), journal)
		if err != nil {
			return err
		}

		err = journalMQRepo.Create(journal)
		if err != nil {
			return err
		}
	}
	return nil
}
