package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterExpenseCacheRepository interface {
	CreateCode(holdingCode string, code string, expire time.Duration) (bool, error)
	ClearCreatedCode(holdingCode string, code string) error
}

type MasterExpenseCacheRepository struct {
	cache microservice.ICacher
}

func NewMasterExpenseCacheRepository(cache microservice.ICacher) MasterExpenseCacheRepository {
	return MasterExpenseCacheRepository{
		cache: cache,
	}
}

func (r MasterExpenseCacheRepository) CreateCode(holdingCode string, code string, expire time.Duration) (bool, error) {
	cacheKey := r.createCodeCacheKey(holdingCode, code)
	return r.cache.SetNX(cacheKey, "", expire)
}

func (r MasterExpenseCacheRepository) ClearCreatedCode(holdingCode string, code string) error {
	cacheKey := r.createCodeCacheKey(holdingCode, code)
	return r.cache.Del(cacheKey)
}

func (r MasterExpenseCacheRepository) createCodeCacheKey(holdingCode string, code string) string {
	return fmt.Sprintf("masterexpense:%s-%s:createcode", holdingCode, code)
}
