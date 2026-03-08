package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const employeeCollectionName = "employees"

type Employee struct {
	Code           string            `json:"code" bson:"code"`
	Email          string            `json:"email" bson:"email"`
	Name           string            `json:"name" bson:"name"`
	ProfilePicture string            `json:"profilepicture" bson:"profilepicture"`
	Roles          *[]string         `json:"roles" bson:"roles"`
	IsEnabled      bool              `json:"isenabled" bson:"isenabled"`
	IsUsePOS       bool              `json:"isusepos" bson:"isusepos"`
	Contact        EmployeeContact   `json:"contact" bson:"contact"`
	PinCode        string            `json:"pincode" bson:"pincode"`
	Branches       *[]EmployeeBranch `json:"branches" bson:"branches"`
}

type EmployeeBranch struct {
	models.DocIdentity `bson:"inline"`
	Code               string          `json:"code" bson:"code"`
	Names              *[]models.NameX `json:"names" bson:"names"`
}

type EmployeeContact struct {
	Address         string  `json:"address" bson:"address"`
	CountryCode     string  `json:"countrycode" bson:"countrycode"`
	ProvinceCode    string  `json:"provincecode" bson:"provincecode"`
	DistrictCode    string  `json:"districtcode" bson:"districtcode"`
	SubDistrictCode string  `json:"subdistrictcode" bson:"subdistrictcode"`
	ZipCode         string  `json:"zipcode" bson:"zipcode"`
	PhoneNumber     string  `json:"phonenumber" bson:"phonenumber"`
	Latitude        float64 `json:"latitude" bson:"latitude"`
	Longitude       float64 `json:"longitude" bson:"longitude"`
}

type EmployeeInfo struct {
	models.DocIdentity `bson:"inline"`
	Employee           `bson:"inline"`
}

func (EmployeeInfo) CollectionName() string {
	return employeeCollectionName
}

type EmployeeData struct {
	models.ShopIdentity `bson:"inline"`
	EmployeeInfo        `bson:"inline"`
}

type EmployeeDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EmployeeData       `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	EmployeePassword   `bson:"inline" gorm:"embedded;"`
}

func (EmployeeDoc) CollectionName() string {
	return employeeCollectionName
}

type EmployeeItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (EmployeeItemGuid) CollectionName() string {
	return employeeCollectionName
}

type EmployeeActivity struct {
	EmployeeData        `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (EmployeeActivity) CollectionName() string {
	return employeeCollectionName
}

type EmployeeDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (EmployeeDeleteActivity) CollectionName() string {
	return employeeCollectionName
}

type EmployeeRequestRegister struct {
	Employee         `bson:"inline" gorm:"embedded;"`
	EmployeePassword `bson:"inline" gorm:"embedded;"`
}

type EmployeeRequestLogin struct {
	models.ShopIdentity
	Code     string `json:"code" bson:"code"`
	Password string `json:"password" bson:"password"`
}

type EmployeeRequestUpdate struct {
	Employee `bson:"inline" gorm:"embedded;"`
}

type EmployeeRequestPassword struct {
	Code            string `json:"code" bson:"code"`
	CurrentPassword string `json:"currentpassword" bson:"currentpassword"`
	NewPassword     string `json:"newpassword" bson:"newpassword"`
}

type EmployeePassword struct {
	Password string `json:"password" bson:"password"`
}
