package repositories

import (
	"smlcloudplatform/pkg/microservice"
	"time"

	"github.com/go-redis/redis/v8"
)

type IJournalCacheRepository interface {
	Pub(channel string, message interface{}) error
	Sub(channel string) (<-chan *redis.Message, string, error)
	Unsub(subID string) error

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

type JournalCacheRepository struct {
	cache microservice.ICacher
}

func NewJournalCacheRepository(cache microservice.ICacher) *JournalCacheRepository {
	return &JournalCacheRepository{cache: cache}
}

func (repo JournalCacheRepository) Pub(channel string, message interface{}) error {
	return repo.cache.Pub(channel, message)
}

func (repo JournalCacheRepository) Sub(channel string) (<-chan *redis.Message, string, error) {
	return repo.cache.Sub(channel)
}

func (repo JournalCacheRepository) Unsub(subID string) error {
	return repo.cache.Unsub(subID)
}

func (repo JournalCacheRepository) HSet(key string, data map[string]interface{}) error {

	return repo.cache.HMSet(key, data)
}

func (repo JournalCacheRepository) HGet(key string, field string) (string, error) {
	return repo.cache.HGet(key, field)
}

func (repo JournalCacheRepository) HGetAll(key string) (map[string]string, error) {
	return repo.cache.HGetAll(key)
}

func (repo JournalCacheRepository) HFields(key string, pattern string) ([]string, error) {
	return repo.cache.HFields(key, pattern)
}

func (repo JournalCacheRepository) HExists(key string, field string) (bool, error) {
	return repo.cache.HExists(key, field)
}

func (repo JournalCacheRepository) Exists(key string) (bool, error) {
	return repo.cache.Exists(key)
}

func (repo JournalCacheRepository) HDel(key string, fields ...string) error {
	return repo.cache.HDel(key, fields...)
}

func (repo JournalCacheRepository) Del(key ...string) error {
	return repo.cache.Del(key...)
}

func (repo JournalCacheRepository) Expire(key string, expire time.Duration) error {
	return repo.cache.Expire(key, expire)
}
