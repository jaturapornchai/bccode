package repositories

import (
	"context"
	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IAuthenticationRepository interface {
	FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error)
	FindUser(ctx context.Context, id string) (*models.UserDoc, error)
	FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error)
	CreateUser(ctx context.Context, doc models.UserDoc) (primitive.ObjectID, error)
	UpdateUser(ctx context.Context, username string, user models.UserDoc) error
	DeleteUser(ctx context.Context, username string) error
}

type AuthenticationRepository struct {
	pst microservice.IPersisterMongo
}

func NewAuthenticationRepository(pst microservice.IPersisterMongo) IAuthenticationRepository {
	return AuthenticationRepository{
		pst: pst,
	}
}

func (r AuthenticationRepository) FindByIdentity(ctx context.Context, fieldName string, value string) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{fieldName: value}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationRepository) FindUser(ctx context.Context, username string) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"username": username}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationRepository) FindByPhonenumber(ctx context.Context, phonenumber models.PhoneNumberField) (*models.UserDoc, error) {

	findUser := &models.UserDoc{}
	err := r.pst.FindOne(ctx, &models.UserDoc{}, bson.M{"countrycode": phonenumber.CountryCode, "phonenumber": phonenumber.PhoneNumber}, findUser)

	if err != nil {
		return nil, err
	}

	return findUser, nil
}

func (r AuthenticationRepository) CreateUser(ctx context.Context, user models.UserDoc) (primitive.ObjectID, error) {

	idx, err := r.pst.Create(ctx, &models.UserDoc{}, user)

	if err != nil {
		return primitive.NilObjectID, err
	}
	return idx, nil
}

func (r AuthenticationRepository) UpdateUser(ctx context.Context, username string, user models.UserDoc) error {
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
	return nil
}

func (r AuthenticationRepository) DeleteUser(ctx context.Context, username string) error {

	filterDoc := map[string]interface{}{
		"username": username,
	}

	err := r.pst.Delete(ctx, &models.UserDoc{}, filterDoc)

	if err != nil {
		return err
	}
	return nil
}
