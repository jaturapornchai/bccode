package repositories

import (
	"context"
	"encoding/json"
	"errors"
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
	UpdateUserByUID(ctx context.Context, userUID string, user models.UserDoc) error
	DeleteUser(ctx context.Context, username string) error
	FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error)
	FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error)
	CreateAuthAudit(ctx context.Context, audit models.AuthAudit) error
	CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error)
	EnsureGoogleIdentityIndexes(ctx context.Context) error
}

func (r AuthenticationMongoCacheRepository) FindGoogleIdentity(ctx context.Context, issuer string, subject string) (*models.GoogleIdentity, error) {
	identity := &models.GoogleIdentity{}
	if err := r.pst.FindOne(ctx, &models.GoogleIdentity{}, bson.M{"issuer": issuer, "subject": subject}, identity); err != nil {
		return nil, err
	}
	return identity, nil
}

func (r AuthenticationMongoCacheRepository) FindUserByUID(ctx context.Context, userUID string) (*models.UserDoc, error) {
	user := &models.UserDoc{}
	if err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"uid": userUID, "isdeleted": bson.M{"$ne": true}}, user); err != nil {
		return nil, err
	}
	return user, nil
}

func (r AuthenticationMongoCacheRepository) CreateAuthAudit(ctx context.Context, audit models.AuthAudit) error {
	_, err := r.pst.Create(ctx, &models.AuthAudit{}, audit)
	return err
}

func (r AuthenticationMongoCacheRepository) CreateGoogleUserIdentity(ctx context.Context, user models.UserDoc, identity models.GoogleIdentity, audit models.AuthAudit) (models.UserDoc, error) {
	if err := r.EnsureGoogleIdentityIndexes(ctx); err != nil {
		return models.UserDoc{}, err
	}
	linkedUser := models.UserDoc{}
	err := r.pst.Transaction(ctx, func(transactionContext context.Context) error {
		email := strings.ToLower(strings.TrimSpace(user.Email))
		if email == "" {
			return errors.New("verified email is required")
		}
		userFilter := verifiedEmailUserFilter(email)
		matches, err := r.pst.Count(transactionContext, &models.UserDoc{}, userFilter)
		if err != nil {
			return err
		}
		if matches > 1 {
			return errors.New("verified email matches multiple users")
		}
		if matches == 1 {
			if err := r.pst.FindOne(transactionContext, &models.UserDoc{}, userFilter, &linkedUser); err != nil {
				return err
			}
			if strings.TrimSpace(linkedUser.UID) == "" {
				return errors.New("existing user has no stable identity")
			}
		} else {
			if _, err := r.pst.Create(transactionContext, &models.UserDoc{}, user); err != nil {
				return err
			}
			linkedUser = user
		}
		identity.UserUID = linkedUser.UID
		audit.UserUID = linkedUser.UID
		if _, err := r.pst.Create(transactionContext, &models.GoogleIdentity{}, identity); err != nil {
			return err
		}
		_, err = r.pst.Create(transactionContext, &models.AuthAudit{}, audit)
		return err
	})
	if err != nil {
		return models.UserDoc{}, err
	}
	return linkedUser, nil
}

func verifiedEmailUserFilter(email string) bson.M {
	return bson.M{
		"isdeleted": bson.M{"$ne": true},
		"$expr": bson.M{"$eq": bson.A{
			bson.M{"$toLower": bson.M{"$trim": bson.M{
				"input": bson.M{"$ifNull": bson.A{"$email", ""}},
			}}},
			strings.ToLower(strings.TrimSpace(email)),
		}},
	}
}

func (r AuthenticationMongoCacheRepository) EnsureGoogleIdentityIndexes(ctx context.Context) error {
	indexes := []struct {
		model interface{}
		name  string
		keys  bson.D
	}{
		{&models.UserDoc{}, "uniq_users_guidfixed", bson.D{{Key: "guidfixed", Value: 1}}},
		{&models.UserDoc{}, "uniq_users_uid", bson.D{{Key: "uid", Value: 1}}},
		{&models.GoogleIdentity{}, "uniq_googleidentities_identityuid", bson.D{{Key: "identityuid", Value: 1}}},
		{&models.GoogleIdentity{}, "uniq_googleidentities_issuer_subject", bson.D{{Key: "issuer", Value: 1}, {Key: "subject", Value: 1}}},
		{&models.AuthAudit{}, "uniq_authaudits_audituid", bson.D{{Key: "audituid", Value: 1}}},
	}
	for _, index := range indexes {
		if _, err := r.pst.CreateIndex(ctx, index.model, index.name, index.keys); err != nil {
			return fmt.Errorf("ensure %s: %w", index.name, err)
		}
	}
	if _, err := r.pst.CreatePartialUniqueIndex(
		ctx,
		&models.UserDoc{},
		"uniq_users_username",
		bson.D{{Key: "username", Value: 1}},
		bson.M{"username": bson.M{"$type": "string", "$gt": ""}},
	); err != nil {
		return fmt.Errorf("ensure uniq_users_username: %w", err)
	}
	return nil
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

func (r AuthenticationMongoCacheRepository) UpdateUserByUID(ctx context.Context, userUID string, user models.UserDoc) error {
	userUID = strings.TrimSpace(userUID)
	if userUID == "" {
		return fmt.Errorf("user identity is required")
	}
	existingUser := &models.UserDoc{}
	if err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"uid": userUID}, existingUser); err != nil {
		return err
	}
	user.UID = userUID
	if err := r.pst.UpdateOne(ctx, &models.UserDoc{}, map[string]interface{}{"uid": userUID}, user); err != nil {
		return err
	}
	r.clearnCache(existingUser.Username)
	if user.Username != existingUser.Username {
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
