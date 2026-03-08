package repositories_test

import (
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

const MockShopID = "TESTSHOP"
const prefixName = "ws"

var cache *microservice.Cacher
var repoCacheMock repositories.JournalCacheRepository

func init() {
	cacheConfig := mock.NewCacherConfig()
	cache = microservice.NewCacher(cacheConfig)
	repoCacheMock = *repositories.NewJournalCacheRepository(cache)
	// repoMock = category.NewCategoryRepository(mongoPersister)
}

func TestMSet(t *testing.T) {
	data := map[string]interface{}{
		"ws:s1": "mdata1",
		"ws:s2": "mdata2",
	}
	cacheKey := prefixName + "-sh1p1"
	repoCacheMock.HSet(cacheKey, data)
}

func TestMGet(t *testing.T) {
	cacheKey := prefixName + "-sh1p1"
	result, err := repoCacheMock.HGet(cacheKey, "m1")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
}

func TestHGetAll(t *testing.T) {
	cacheKey := prefixName + "-sh1p1"
	result, err := repoCacheMock.HFields(cacheKey, "ws:*")

	if err != nil {
		t.Error(err)
		return
	}

	t.Log(result)
}
