package repositories

import (
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"github.com/jellydator/ttlcache/v3"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IAuthenticationMongoCacheRepository interface {
	FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error)
	FindUser(ctx context.Context, id string) (*models.UserDoc, error)
	FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error)
	FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error)
	CreateUser(ctx context.Context, doc models.UserDoc) (primitive.ObjectID, error)
	UpdateUser(ctx context.Context, username string, user models.UserDoc) error
	DeleteUser(ctx context.Context, username string) error
}

type AuthenticationMongoCacheRepository struct {
	pst         microservice.IPersisterMongo
	cache       microservice.ICacher
	memorycache *ttlcache.Cache[string, models.UserDoc]
}

func NewAuthenticationMongoCacheRepository(pst microservice.IPersisterMongo, cache microservice.ICacher) IAuthenticationMongoCacheRepository {
	cachex := ttlcache.New[string, models.UserDoc](
		ttlcache.WithTTL[string, models.UserDoc](15 * time.Second),
	)
	return AuthenticationMongoCacheRepository{
		pst:         pst,
		cache:       cache,
		memorycache: cachex,
	}
}

func (r AuthenticationMongoCacheRepository) getCacheKey(username string) string {
	return fmt.Sprintf("user:%s", username)
}

func (r AuthenticationMongoCacheRepository) clearnCache(username string) {
	cacheKey := r.getCacheKey(username)
	r.memorycache.Delete(cacheKey)
	r.cache.Del(cacheKey)
}

func (r AuthenticationMongoCacheRepository) FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{fieldName: value}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationMongoCacheRepository) FindUser(ctx context.Context, username string) (*models.UserDoc, error) {

	cacheKey := r.getCacheKey(username)

	cacheItem := r.memorycache.Get(cacheKey)

	if cacheItem != nil {
		userCacheMem := cacheItem.Value()
		return &userCacheMem, nil
	}

	userCache, err := r.cache.Get(cacheKey)

	if err != nil {
		fmt.Println(err.Error())
	}

	if len(userCache) > 0 {
		user := &models.UserDoc{}
		err := json.Unmarshal([]byte(userCache), user)

		r.memorycache.Set(cacheKey, *user, time.Second*15)
		if err == nil {
			return user, nil
		}
	}

	findUser := &models.UserDoc{}
	err = r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"username": username}, findUser)

	if err != nil {
		return nil, err
	}

	tempUser, err := json.Marshal(findUser)

	if err != nil {
		fmt.Println(err.Error())
	}

	if err == nil {
		func() {
			err = r.cache.SetS(cacheKey, string(tempUser), time.Second*60)
			r.memorycache.Set(cacheKey, *findUser, time.Second*15)

			if err != nil {
				fmt.Println(err.Error())
			}
		}()
	}

	return findUser, nil
}

// FindByLineUserID — ค้นหา user จาก LINE User ID (จาก users collection)
func (r AuthenticationMongoCacheRepository) FindByLineUserID(ctx context.Context, lineUserID string) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"lineuserid": lineUserID}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationMongoCacheRepository) FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"countrycode": phonenumber.CountryCode, "phonenumber": phonenumber.PhoneNumber}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationMongoCacheRepository) CreateUser(ctx context.Context, user models.UserDoc) (primitive.ObjectID, error) {

	idx, err := r.pst.Create(ctx, &models.UserDoc{}, user)

	if err != nil {
		return primitive.NilObjectID, err
	}

	r.clearnCache(user.Username)

	return idx, nil
}

func (r AuthenticationMongoCacheRepository) UpdateUser(ctx context.Context, username string, user models.UserDoc) error {
	existingUser := &models.UserDoc{}
	if err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"username": username}, existingUser); err == nil {
		if strings.TrimSpace(existingUser.UID) != "" {
			user.UID = existingUser.UID
		}
	}
	if strings.TrimSpace(user.UID) == "" {
		user.UID = utils.NewGUID()
	}

	filterDoc := map[string]interface{}{
		"username": username,
	}

	err := r.pst.UpdateOne(ctx, &models.UserDoc{}, filterDoc, user)

	if err != nil {
		return err
	}

	r.clearnCache(username)
	if user.Username != username {
		r.clearnCache(user.Username)
	}

	return nil
}

func (r AuthenticationMongoCacheRepository) DeleteUser(ctx context.Context, username string) error {

	filterDoc := map[string]interface{}{
		"username": username,
	}

	err := r.pst.Delete(ctx, &models.UserDoc{}, filterDoc)

	if err != nil {
		return err
	}

	r.clearnCache(username)

	return nil
}

// func (r AuthenticationMongoCacheRepository) UserDeactiveLoginError(ctx context.Context, user models.UserDoc) error {

// 	filterDoc := map[string]interface{}{
// 		"username": user.Username,
// 	}

// 	err := r.pst.UpdateOne(ctx, &models.UserDoc{}, filterDoc, user)
// 	if err != nil {
// 		return err
// 	}

// 	r.clearnCache(user.Username)

// 	return nil
// }
