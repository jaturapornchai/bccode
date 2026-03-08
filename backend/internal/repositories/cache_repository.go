package repositories

import (
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ICacheRepository interface {
	Save(shopID string, moduleName string) error
	Get(shopID string, moduleName string) (time.Time, error)
}

type CacheRepository struct {
	cache microservice.ICacher
}

func NewCacheRepository(cache microservice.ICacher) CacheRepository {
	return CacheRepository{
		cache: cache,
	}
}

func (repo CacheRepository) Save(key string, value interface{}, expire time.Duration) error {

	return repo.cache.Set(key, value, expire)
}

func (repo CacheRepository) Get(key string) (string, error) {
	return repo.cache.Get(key)
}
