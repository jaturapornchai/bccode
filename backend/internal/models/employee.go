package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const employeeCollectionName string = "employees"

type EmployeePassword struct {
	Password string `json:"password" bson:"password"`
}

type Employee struct {
	Code  string `json:"code" bson:"code"`
	Email string `json:"email" bson:"email"`
	// Username       string    `json:"username" bson:"username"`
	Name           string    `json:"name" bson:"name"`
	ProfilePicture string    `json:"profilepicture" bson:"profilepicture"`
	Roles          *[]string `json:"roles" bson:"roles"`
	IsEnabled      bool      `json:"isenabled" bson:"isenabled"`
}

type EmployeeInfo struct {
	DocIdentity `bson:"inline" gorm:"embedded;"`
	Employee    `bson:"inline" gorm:"embedded;"`
}

func (EmployeeInfo) CollectionName() string {
	return employeeCollectionName
}

type EmployeeData struct {
	ShopIdentity `bson:"inline" gorm:"embedded;"`
	EmployeeInfo `bson:"inline" gorm:"embedded;"`
}

type EmployeeDoc struct {
	ID               primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	EmployeeData     `bson:"inline"`
	ActivityDoc      `bson:"inline"`
	EmployeePassword `bson:"inline" gorm:"embedded;"`
}

func (EmployeeDoc) CollectionName() string {
	return employeeCollectionName
}

type EmployeeRequestRegister struct {
	Employee         `bson:"inline" gorm:"embedded;"`
	EmployeePassword `bson:"inline" gorm:"embedded;"`
}

type EmployeeRequestLogin struct {
	ShopIdentity
	Code     string `json:"code" bson:"code"`
	Password string `json:"password" bson:"password"`
}

type EmployeeRequestUpdate struct {
	Employee `bson:"inline" gorm:"embedded;"`
}

type EmployeeRequestPassword struct {
	Code     string `json:"code" bson:"code"`
	Password string `json:"password" bson:"password"`
}

type EmployeePageResponse struct {
	Success    bool                   `json:"success"`
	Data       []EmployeeInfo         `json:"data,omitempty"`
	Pagination PaginationDataResponse `json:"pagination,omitempty"`
}

type EmployeeActivity struct {
	EmployeeData `bson:"inline"`
	CreatedAt    *time.Time `json:"createdat,omitempty" bson:"createdat,omitempty"`
	UpdatedAt    *time.Time `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	DeletedAt    *time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
}

func (EmployeeActivity) CollectionName() string {
	return employeeCollectionName
}

type EmployeeDeleteActivity struct {
	Identity  `bson:"inline"`
	CreatedAt *time.Time `json:"createdat,omitempty" bson:"createdat,omitempty"`
	UpdatedAt *time.Time `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	DeletedAt *time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
}

func (EmployeeDeleteActivity) CollectionName() string {
	return employeeCollectionName
}
