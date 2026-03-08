package repositories

import (
	"encoding/json"
	"smlcloudplatform/internal/restaurant/notifierdevice/models"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type INotifierDeviceCacheRepository interface {
	Exists(refCode string) (bool, error)
	Save(refCode string, value models.NotifierDeviceAuth, expire time.Duration) error
	Get(refCode string) (models.NotifierDeviceAuth, error)
}

type NotifierDeviceCacheRepository struct {
	cache microservice.ICacher
}

func NewNotifierDeviceCacheRepository(cache microservice.ICacher) NotifierDeviceCacheRepository {
	return NotifierDeviceCacheRepository{
		cache: cache,
	}
}

func (repo NotifierDeviceCacheRepository) Exists(refCode string) (bool, error) {
	cacheKey := repo.generateCacheKey(refCode)
	return repo.cache.Exists(cacheKey)
}

func (repo NotifierDeviceCacheRepository) Save(refCode string, value models.NotifierDeviceAuth, expire time.Duration) error {
	cacheKey := repo.generateCacheKey(refCode)

	return repo.cache.Set(cacheKey, value, expire)
}

func (repo NotifierDeviceCacheRepository) Get(refCode string) (models.NotifierDeviceAuth, error) {
	cacheKey := repo.generateCacheKey(refCode)
	rawResult, err := repo.cache.Get(cacheKey)

	if err != nil {
		return models.NotifierDeviceAuth{}, err
	}

	var result models.NotifierDeviceAuth

	err = json.Unmarshal([]byte(rawResult), &result)

	if err != nil {
		return models.NotifierDeviceAuth{}, err
	}

	return result, nil
}

func (repo NotifierDeviceCacheRepository) generateCacheKey(refCode string) string {
	return "notifier:" + refCode
}
