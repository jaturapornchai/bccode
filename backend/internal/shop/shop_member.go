package shop

/*
import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"strconv"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func (svc *ShopServiceOld) SearchMember(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	q := ctx.QueryParam("q")
	page, err := strconv.Atoi(ctx.QueryParam("page"))
	if err != nil {
		page = 1
	}

	limit, err := strconv.Atoi(ctx.QueryParam("limit"))

	if err != nil {
		limit = 20
	}

	shopList := []models.MemberInfo{}

	searchText := bson.M{"$regex": primitive.Regex{
		Pattern: ".*" + q + ".*",
		Options: "",
	}}

	pagination, err := pst.FindPage(&models.MemberInfo{}, limit, page, bson.M{"createdby": authUsername, "shopid": shopID, "$or": []interface{}{
		bson.M{"email": searchText},
		bson.M{"username": searchText},
		bson.M{"name": searchText},
	}}, &shopList)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success:    true,
		Pagination: pagination,
		Data:       shopList,
	})
	return nil
}

func (svc *ShopServiceOld) CreateMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	memberReq := &models.MemberRequest{}
	err := json.Unmarshal([]byte(input), &memberReq)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	if err = ctx.Validate(memberReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	findMember := &models.Member{}
	err = pst.FindOne(&models.Member{}, bson.M{"shop_id": shopID, "$or": []interface{}{
		bson.M{"email": memberReq.Email},
		bson.M{"username": memberReq.Username},
	}}, findMember)

	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if findMember.Email == memberReq.Email {
		ctx.ResponseError(400, "email is exists.")
		return err
	}

	if findMember.Username == memberReq.Username {
		ctx.ResponseError(400, "username is exists.")
		return err
	}

	hashPassword, err := utils.HashPassword(memberReq.Password)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	member := &models.Member{
		ShopID: shopID,
		Email:      memberReq.Email,
		Username:   memberReq.Username,
		Password:   hashPassword,
		Name:       memberReq.Name,
		CreatedBy:  authUsername,
		CreatedAt:  time.Now(),
	}

	idx, err := pst.Create(&models.Member{}, member)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

func (svc *ShopServiceOld) EditMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	input := ctx.ReadInput()

	memberReq := &models.MemberRequestEdit{}
	err := json.Unmarshal([]byte(input), &memberReq)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	if err = ctx.Validate(memberReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	findIDx, _ := primitive.ObjectIDFromHex(id)
	findMember := &models.Member{}
	err = pst.FindByID(&models.Member{}, "guidfixed", findIDx, findMember)
	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if authUsername != findMember.CreatedBy {
		ctx.ResponseError(400, "shop invalid")
		return err
	}

	if len(findMember.Email) < 1 || len(findMember.Username) < 1 {
		ctx.ResponseError(400, "member is not exists")
		return err
	}

	findMember.Email = memberReq.Email
	findMember.Username = memberReq.Username
	findMember.Name = memberReq.Name

	idx, _ := primitive.ObjectIDFromHex(id)
	err = pst.UpdateOne(&models.Member{}, "guidfixed", idx, findMember)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success: true,
	})
	return nil
}

func (svc *ShopServiceOld) DeleteMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	findMember := &models.Member{}

	idx, _ := primitive.ObjectIDFromHex(id)
	err := pst.FindOne(&models.Member{}, bson.M{"_id": idx, "createdby": authUsername}, findMember)

	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if len(findMember.CreatedBy) < 1 {
		ctx.ResponseError(400, "id invalid")
		return err
	}

	err = pst.DeleteByID(&models.Member{}, id)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success: true,
	})
	return nil
}

func (svc *ShopServiceOld) GetMemberInfo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	memberInfo := &models.MemberInfo{}

	err := pst.FindOne(&models.MemberInfo{}, bson.M{"guidfixed": id, "createdby": authUsername, "shop_id": shopID}, memberInfo)

	if err != nil {
		ctx.ResponseError(400, "not found")
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success: true,
		Data:    memberInfo,
	})
	return nil
}

func (svc *ShopServiceOld) ChangePasswordMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username

	id := ctx.Param("id")

	pst := svc.ms.MongoPersister(svc.cfg.MongoPersisterConfig())

	input := ctx.ReadInput()

	memberReq := &models.MemberRequestPassword{}
	err := json.Unmarshal([]byte(input), &memberReq)

	if err != nil {
		ctx.ResponseError(400, "user payload invalid")
		return err
	}

	if err = ctx.Validate(memberReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	findMember := &models.Member{}
	err = pst.FindByID(&models.Member{}, "guidfixed", id, findMember)
	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if authUsername != findMember.CreatedBy {
		ctx.ResponseError(400, "shop invalid")
		return err
	}

	if len(memberReq.Password) > 0 {
		hashPassword, err := utils.HashPassword(memberReq.Password)

		if err != nil {
			ctx.ResponseError(400, err.Error())
			return err
		}
		findMember.Password = hashPassword
	}

	err = pst.UpdateOne(&models.Member{}, "guidfixed", id, findMember)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, models.ApiResponse{
		Success: true,
	})
	return nil
}
*/
