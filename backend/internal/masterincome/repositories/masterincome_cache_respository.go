package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMasterIncomeCacheRepository interface {
	CreateCode(holdingCode string, code string, expire time.Duration) (bool, error)
	ClearCreatedCode(holdingCode string, code string) error
}

type MasterIncomeCacheRepository struct {
	cache microservice.ICacher
}

func NewMasterIncomeCacheRepository(cache microservice.ICacher) MasterIncomeCacheRepository {
	return MasterIncomeCacheRepository{
		cache: cache,
	}
}

func (r MasterIncomeCacheRepository) CreateCode(holdingCode string, code string, expire time.Duration) (bool, error) {
	cacheKey := r.createCodeCacheKey(holdingCode, code)
	return r.cache.SetNX(cacheKey, "", expire)
}

func (r MasterIncomeCacheRepository) ClearCreatedCode(holdingCode string, code string) error {
	cacheKey := r.createCodeCacheKey(holdingCode, code)
	return r.cache.Del(cacheKey)
}

func (r MasterIncomeCacheRepository) createCodeCacheKey(holdingCode string, code string) string {
	return fmt.Sprintf("masterincome:%s-%s:createcode", holdingCode, code)
}
