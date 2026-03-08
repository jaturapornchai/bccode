package services

import (
	"smlcloudplatform/internal/mastersync/repositories"
	"time"
)

type IMasterSyncService interface {
	GetStatus(shopID string, syncModules []string) (map[string]time.Time, error)
}

type MasterSyncService struct {
	cacheRepo repositories.IMasterSyncCacheRepository
}

func NewMasterSyncService(cacheRepo repositories.IMasterSyncCacheRepository) MasterSyncService {

	return MasterSyncService{
		cacheRepo: cacheRepo,
	}
}

func (svc MasterSyncService) GetStatus(shopID string, syncModules []string) (map[string]time.Time, error) {
	moduleStatus := map[string]time.Time{}

	for _, moduleName := range syncModules {
		lastTime, err := svc.cacheRepo.Get(shopID, moduleName)
		moduleStatus[moduleName] = lastTime
		if err != nil {
			return nil, err
		}
	}

	return moduleStatus, nil
}
