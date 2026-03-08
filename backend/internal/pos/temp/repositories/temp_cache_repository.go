package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICacheRepository interface {
	Save(shopID, branchCode string, doc string, expire time.Duration) error
	Get(shopID, branchCode string) (string, error)
	Delete(shopID, branchCode string) error
}

type CacheRepository struct {
	cache microservice.ICacher
}

func NewCacheRepository(cache microservice.ICacher) CacheRepository {
	return CacheRepository{
		cache: cache,
	}
}

func (repo CacheRepository) Save(shopID, branchCode string, doc string, expire time.Duration) error {
	cacheKey := repo.generateKey(shopID, branchCode)
	return repo.cache.Set(cacheKey, doc, expire)
}

func (repo CacheRepository) Get(shopID, branchCode string) (string, error) {
	cacheKey := repo.generateKey(shopID, branchCode)
	result, err := repo.cache.Get(cacheKey)

	if err != nil {
		return "", err
	}

	return result, nil
}

func (repo CacheRepository) Delete(shopID, branchCode string) error {
	cacheKey := repo.generateKey(shopID, branchCode)
	return repo.cache.Del(cacheKey)
}

func (repo CacheRepository) generateKey(shopID, branchCode string) string {
	cacheKey := fmt.Sprintf("pos:%s:%s:temp", shopID, branchCode)
	return cacheKey
}
