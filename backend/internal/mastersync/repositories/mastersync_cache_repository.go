package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type IMasterSyncCacheRepository interface {
	Save(shopID string, moduleName string) error
	Get(shopID string, moduleName string) (time.Time, error)
}

type MasterSyncCacheRepository struct {
	cache        microservice.ICacher
	allMasterKey string
}

func NewMasterSyncCacheRepository(cache microservice.ICacher) MasterSyncCacheRepository {
	return MasterSyncCacheRepository{
		cache:        cache,
		allMasterKey: "all",
	}
}

func (repo MasterSyncCacheRepository) Save(shopID string, moduleName string) error {
	changeTime := time.Now().Format(time.RFC3339)
	cacheModuleKey := repo.getCacheModuleKeyWithModule(shopID, moduleName)
	cacheModuleAllMasterKey := repo.getCacheModuleKeyWithModule(shopID, repo.allMasterKey)

	repo.cache.SetNoExpire(cacheModuleAllMasterKey, changeTime)
	return repo.cache.SetNoExpire(cacheModuleKey, changeTime)
}

func (repo MasterSyncCacheRepository) Get(shopID string, moduleName string) (time.Time, error) {
	cacheModuleKey := repo.getCacheModuleKeyWithModule(shopID, moduleName)

	strTime, err := repo.cache.Get(cacheModuleKey)
	if err != nil {
		fmt.Println(err)
		return time.Time{}, nil
	}

	if len(strTime) == 0 {
		return time.Time{}, nil
	}

	strTime = strings.ReplaceAll(strTime, "\"", "")

	valTime, err := time.Parse(time.RFC3339, strTime)
	if err != nil {
		fmt.Println(err)
		return time.Time{}, nil
	}

	return valTime, nil
}

func (repo MasterSyncCacheRepository) getCacheModuleKeyWithModule(shopID string, moduleName string) string {
	return fmt.Sprintf("mastersync-%s::%s", shopID, moduleName)
}
