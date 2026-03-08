package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterIncomeCacheRepository interface {
	CreateCode(shopID string, code string, expire time.Duration) (bool, error)
	ClearCreatedCode(shopID string, code string) error
}

type MasterIncomeCacheRepository struct {
	cache microservice.ICacher
}

func NewMasterIncomeCacheRepository(cache microservice.ICacher) MasterIncomeCacheRepository {
	return MasterIncomeCacheRepository{
		cache: cache,
	}
}

func (r MasterIncomeCacheRepository) CreateCode(shopID string, code string, expire time.Duration) (bool, error) {
	cacheKey := r.createCodeCacheKey(shopID, code)
	return r.cache.SetNX(cacheKey, "", expire)
}

func (r MasterIncomeCacheRepository) ClearCreatedCode(shopID string, code string) error {
	cacheKey := r.createCodeCacheKey(shopID, code)
	return r.cache.Del(cacheKey)
}

func (r MasterIncomeCacheRepository) createCodeCacheKey(shopID string, code string) string {
	return fmt.Sprintf("masterincome:%s-%s:createcode", shopID, code)
}
