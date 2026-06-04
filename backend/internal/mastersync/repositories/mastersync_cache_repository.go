package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"
)

type IMasterSyncCacheRepository interface {
	Save(holdingCode string, moduleName string) error
	Get(holdingCode string, moduleName string) (time.Time, error)
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

func (repo MasterSyncCacheRepository) Save(holdingCode string, moduleName string) error {
	changeTime := time.Now().Format(time.RFC3339)
	cacheModuleKey := repo.getCacheModuleKeyWithModule(holdingCode, moduleName)
	cacheModuleAllMasterKey := repo.getCacheModuleKeyWithModule(holdingCode, repo.allMasterKey)

	repo.cache.SetNoExpire(cacheModuleAllMasterKey, changeTime)
	return repo.cache.SetNoExpire(cacheModuleKey, changeTime)
}

func (repo MasterSyncCacheRepository) Get(holdingCode string, moduleName string) (time.Time, error) {
	cacheModuleKey := repo.getCacheModuleKeyWithModule(holdingCode, moduleName)

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

func (repo MasterSyncCacheRepository) getCacheModuleKeyWithModule(holdingCode string, moduleName string) string {
	return fmt.Sprintf("mastersync-%s::%s", holdingCode, moduleName)
}
