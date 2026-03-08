package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterExpenseCacheRepository interface {
	CreateCode(shopID string, code string, expire time.Duration) (bool, error)
	ClearCreatedCode(shopID string, code string) error
}

type MasterExpenseCacheRepository struct {
	cache microservice.ICacher
}

func NewMasterExpenseCacheRepository(cache microservice.ICacher) MasterExpenseCacheRepository {
	return MasterExpenseCacheRepository{
		cache: cache,
	}
}

func (r MasterExpenseCacheRepository) CreateCode(shopID string, code string, expire time.Duration) (bool, error) {
	cacheKey := r.createCodeCacheKey(shopID, code)
	return r.cache.SetNX(cacheKey, "", expire)
}

func (r MasterExpenseCacheRepository) ClearCreatedCode(shopID string, code string) error {
	cacheKey := r.createCodeCacheKey(shopID, code)
	return r.cache.Del(cacheKey)
}

func (r MasterExpenseCacheRepository) createCodeCacheKey(shopID string, code string) string {
	return fmt.Sprintf("masterexpense:%s-%s:createcode", shopID, code)
}
