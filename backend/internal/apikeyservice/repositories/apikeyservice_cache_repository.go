package repositories

import (
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IApiKeyServiceCacheRepository interface {
	HSet(key string, data map[string]interface{}) error
	HGet(key string, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	HFields(key string, pattern string) ([]string, error)
	HDel(key string, fields ...string) error
	HExists(key string, field string) (bool, error)
	Exists(key string) (bool, error)
	Del(key ...string) error
	Expire(key string, expire time.Duration) error
}

type ApiKeyServiceCacheRepository struct {
	cache microservice.ICacher
}

func NewApiKeyServiceCacheRepository(cache microservice.ICacher) *ApiKeyServiceCacheRepository {
	return &ApiKeyServiceCacheRepository{cache: cache}
}

func (repo ApiKeyServiceCacheRepository) HSet(key string, data map[string]interface{}) error {

	return repo.cache.HMSet(key, data)
}

func (repo ApiKeyServiceCacheRepository) HGet(key string, field string) (string, error) {
	return repo.cache.HGet(key, field)
}

func (repo ApiKeyServiceCacheRepository) HGetAll(key string) (map[string]string, error) {
	return repo.cache.HGetAll(key)
}

func (repo ApiKeyServiceCacheRepository) HFields(key string, pattern string) ([]string, error) {
	return repo.cache.HFields(key, pattern)
}

func (repo ApiKeyServiceCacheRepository) HExists(key string, field string) (bool, error) {
	return repo.cache.HExists(key, field)
}

func (repo ApiKeyServiceCacheRepository) Exists(key string) (bool, error) {
	return repo.cache.Exists(key)
}

func (repo ApiKeyServiceCacheRepository) HDel(key string, fields ...string) error {
	return repo.cache.HDel(key, fields...)
}

func (repo ApiKeyServiceCacheRepository) Del(key ...string) error {
	return repo.cache.Del(key...)
}

func (repo ApiKeyServiceCacheRepository) Expire(key string, expire time.Duration) error {
	return repo.cache.Expire(key, expire)
}
