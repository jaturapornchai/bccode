package repositories

import (
	"fmt"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"time"
)

type ICacheRepository interface {
	Save(holdingCode, prefixDocNo string, docNumber int, expire time.Duration) error
	Get(holdingCode, prefixDocNo string) (int, error)
}

type CacheRepository struct {
	cache microservice.ICacher
}

func NewCacheRepository(cache microservice.ICacher) CacheRepository {
	return CacheRepository{
		cache: cache,
	}
}

func (repo CacheRepository) Save(holdingCode, prefixDocNo string, docNumber int, expire time.Duration) error {
	cacheKey := repo.generateKey(holdingCode, prefixDocNo)
	return repo.cache.Set(cacheKey, docNumber, expire)
}

func (repo CacheRepository) Get(holdingCode, prefixDocNo string) (int, error) {
	cacheKey := repo.generateKey(holdingCode, prefixDocNo)
	rawResult, err := repo.cache.Get(cacheKey)

	if err != nil {
		return 0, err
	}

	result, err := strconv.Atoi(rawResult)

	if err != nil {
		return 0, err
	}

	return result, nil
}

func (repo CacheRepository) generateKey(holdingCode, prefixDocNo string) string {
	cacheKey := fmt.Sprintf("%s:%s:doc", holdingCode, prefixDocNo)
	return cacheKey
}
