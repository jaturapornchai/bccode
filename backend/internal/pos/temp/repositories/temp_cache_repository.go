package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICacheRepository interface {
	Save(holdingCode, branchCode string, doc string, expire time.Duration) error
	Get(holdingCode, branchCode string) (string, error)
	Delete(holdingCode, branchCode string) error
}

type CacheRepository struct {
	cache microservice.ICacher
}

func NewCacheRepository(cache microservice.ICacher) CacheRepository {
	return CacheRepository{
		cache: cache,
	}
}

func (repo CacheRepository) Save(holdingCode, branchCode string, doc string, expire time.Duration) error {
	cacheKey := repo.generateKey(holdingCode, branchCode)
	return repo.cache.Set(cacheKey, doc, expire)
}

func (repo CacheRepository) Get(holdingCode, branchCode string) (string, error) {
	cacheKey := repo.generateKey(holdingCode, branchCode)
	result, err := repo.cache.Get(cacheKey)

	if err != nil {
		return "", err
	}

	return result, nil
}

func (repo CacheRepository) Delete(holdingCode, branchCode string) error {
	cacheKey := repo.generateKey(holdingCode, branchCode)
	return repo.cache.Del(cacheKey)
}

func (repo CacheRepository) generateKey(holdingCode, branchCode string) string {
	cacheKey := fmt.Sprintf("pos:%s:%s:temp", holdingCode, branchCode)
	return cacheKey
}
