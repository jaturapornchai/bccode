package models

import (
	"smlcloudplatform/internal/models"
)

type Employee struct {
	Code                string                `json:"code"`
	Email               string                `json:"email"`
	Name                string                `json:"name"`
	ProfilePicture      string                `json:"profilepicture"`
	ProfilePictureThumb string                `json:"profilepicturethumb"`
	Roles               *[]string             `json:"roles"`
	IsEnabled           bool                  `json:"isenabled"`
	IsUsePOS            bool                  `json:"isusepos"`
	Contact             EmployeeContact       `json:"contact"`
	PinCode             string                `json:"pincode"`
	Branches            *[]EmployeeBranch     `json:"branches"`
	AccessScopes        []EmployeeAccessScope `json:"accessscopes,omitempty"`
}

type EmployeeBranch struct {
	models.DocIdentity
	Code  string          `json:"code"`
	Names *[]models.NameX `json:"names"`
}

type EmployeeContact struct {
	Address         string  `json:"address"`
	CountryCode     string  `json:"countrycode"`
	ProvinceCode    string  `json:"provincecode"`
	DistrictCode    string  `json:"districtcode"`
	SubDistrictCode string  `json:"subdistrictcode"`
	ZipCode         string  `json:"zipcode"`
	PhoneNumber     string  `json:"phonenumber"`
	Latitude        float64 `json:"latitude"`
	Longitude       float64 `json:"longitude"`
}

// EmployeeAccessScope — บริษัท/สาขาที่พนักงานเข้าใช้งานได้ (จอ /employee ส่ง
// accessscopes มาใน PUT; ก่อนหน้านี้ backend ไม่มี field นี้ ทำให้ save แล้ว
// ข้อมูล scope ถูกตัดทิ้งเงียบ ๆ — เคสจริง 2026-09-01)
type EmployeeAccessScope struct {
	ScopeType    string `json:"scopetype"`
	CompanyUID   string `json:"companyuid,omitempty"`
	BranchUID    string `json:"branchuid,omitempty"`
	BusinessCode string `json:"businesscode,omitempty"`
	BranchCode   string `json:"branchcode,omitempty"`
	AllBranches  bool   `json:"allbranches,omitempty"`
}

type EmployeeInfo struct {
	models.DocIdentity
	Employee
}

type EmployeeRequestRegister struct {
	Employee
}

type EmployeeRequestUpdate struct {
	Employee
}
