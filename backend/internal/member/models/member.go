package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const memberCollectionName = "members"
const memberIndexName string = "members_index"

type MemberType int8

const (
	MemberTypeCustomer MemberType = 0
	MemberTypeLine     MemberType = 1
)

type Member struct {
	Gender       uint8            `json:"gender" bson:"gender"`
	PictureUrl   string           `json:"pictureurl" bson:"pictureurl"`
	Telephone    string           `json:"telephone" bson:"telephone"`
	Email        string           `json:"email" bson:"email"`
	Name         string           `json:"name" bson:"name"`
	Surname      string           `json:"surname" bson:"surname"`
	TaxID        string           `json:"taxid" bson:"taxid"`
	ContactType  int              `json:"contacttype" bson:"contacttype"`
	PersonalType int              `json:"personaltype" bson:"personaltype"`
	BranchType   int              `json:"branchtype" bson:"branchtype"`
	BranchCode   string           `json:"branchcode" bson:"branchcode"`
	LineUID      string           `json:"lineuid" bson:"lineuid"`
	Addresses    *[]MemberAddress `json:"addresses" bson:"addresses"`
	MemberType   MemberType       `json:"membertype" bson:"membertype"`
	Provider     *[]string        `json:"provider" bson:"provider"`
}

type MemberAddress struct {
	Name                  string  `json:"name" bson:"name"`
	Telephone             string  `json:"telephone" bson:"telephone"`
	HomeNumber            string  `json:"homenumber" bson:"homenumber"`
	Build                 string  `json:"build" bson:"build"`
	Floor                 string  `json:"floor" bson:"floor"`
	Village               string  `json:"village" bson:"village"`
	Soi                   string  `json:"soi" bson:"soi"`
	VillageNo             string  `json:"villageno" bson:"villageno"`
	Road                  string  `json:"road" bson:"road"`
	Route                 string  `json:"route" bson:"route"`
	Province              string  `json:"province" bson:"province"`
	District              string  `json:"district" bson:"district"`
	Subdistrict           string  `json:"subdistrict" bson:"subdistrict"`
	Postcode              string  `json:"postcode" bson:"postcode"`
	Remark                string  `json:"remark" bson:"remark"`
	IsMain                bool    `json:"ismain" bson:"ismain"`
	Latitude              string  `json:"latitude" bson:"latitude"`
	Longitude             string  `json:"longitude" bson:"longitude"`
	LalaMovePrice         float64 `json:"lalamoveprice" bson:"lalamoveprice"`
	Distance              float64 `json:"distance" bson:"distance"`
	EstimatedDeliveryTime string  `json:"estimateddeliverytime" bson:"estimateddeliverytime"`
	DistanceImage         string  `json:"distanceimage" bson:"distanceimage"`
}

type MemberInfo struct {
	models.DocIdentity `bson:"inline"`
	Member             `bson:"inline"`
}

func (MemberInfo) CollectionName() string {
	return memberCollectionName
}

type MemberData struct {
	Shops      *[]string `json:"shops" bson:"shops"`
	MemberInfo `bson:"inline"`
}
type MemberDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	MemberData         `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (MemberDoc) CollectionName() string {
	return memberCollectionName
}

type MemberIndex struct {
	models.Index `bson:"inline"`
}

func (MemberIndex) TableName() string {
	return memberIndexName
}

type MemberRequestEdit struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Username string `json:"username" validate:"required"`
}

func (MemberRequestEdit) CollectionName() string {
	return memberCollectionName
}

type MemberRequestPassword struct {
	Password string `json:"password" validate:"required,gte=6"`
}

func (MemberRequestPassword) CollectionName() string {
	return memberCollectionName
}

type MemberActivity struct {
	MemberData          `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MemberActivity) CollectionName() string {
	return memberCollectionName
}

type MemberDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (MemberDeleteActivity) CollectionName() string {
	return memberCollectionName
}

type MemberLastActivityResponse struct {
	New    []MemberActivity       `json:"new" `
	Remove []MemberDeleteActivity `json:"remove"`
}

type MemberFetchUpdateResponse struct {
	Success    bool                          `json:"success"`
	Data       MemberLastActivityResponse    `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type MemberInfoResponse struct {
	Success bool       `json:"success"`
	Data    MemberInfo `json:"data,omitempty"`
}

type MemberPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []MemberInfo                  `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}
