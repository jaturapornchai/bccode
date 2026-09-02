package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const employeeCollectionName string = "employees"

type EmployeePassword struct {
	Password string `json:"password" bson:"password"`
}

// EmployeeAccessScope — บริษัท/สาขาที่พนักงานเข้าใช้งานได้ (โครงเดียวกับ shop models AccessScope)
type EmployeeAccessScope struct {
	ScopeType    string `json:"scopetype" bson:"scopetype"`
	CompanyUID   string `json:"companyuid,omitempty" bson:"companyuid,omitempty"`
	BranchUID    string `json:"branchuid,omitempty" bson:"branchuid,omitempty"`
	BusinessCode string `json:"businesscode,omitempty" bson:"businesscode,omitempty"`
	BranchCode   string `json:"branchcode,omitempty" bson:"branchcode,omitempty"`
	AllBranches  bool   `json:"allbranches,omitempty" bson:"allbranches,omitempty"`
}

type Employee struct {
	Code  string `json:"code" bson:"code"`
	Email string `json:"email" bson:"email"`
	// Username       string    `json:"username" bson:"username"`
	Name                 string                `json:"name" bson:"name"`
	ProfilePicture       string                `json:"profilepicture" bson:"profilepicture"`
	ProfilePictureThumb  string                `json:"profilepicturethumb" bson:"profilepicturethumb"`
	Roles                *[]string             `json:"roles" bson:"roles"`
	IsEnabled            bool                  `json:"isenabled" bson:"isenabled"`
	AccessScopes         []EmployeeAccessScope `json:"accessscopes,omitempty" bson:"accessscopes,omitempty"`
}

type EmployeeInfo struct {
	DocIdentity `bson:"inline" gorm:"embedded;"`
	Employee    `bson:"inline" gorm:"embedded;"`
}

func (EmployeeInfo) CollectionName() string {
	return employeeCollectionName
}

type EmployeeData struct {
	HoldingCodeentity `bson:"inline" gorm:"embedded;"`
	EmployeeInfo      `bson:"inline" gorm:"embedded;"`
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
	HoldingCodeentity
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
