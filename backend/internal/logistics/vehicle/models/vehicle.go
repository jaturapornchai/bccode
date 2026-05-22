package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const vehicleCollectionName = "vehicles"

// Vehicle คือโครงสร้างข้อมูลยานพาหนะสำหรับระบบ BCLMS
type Vehicle struct {
	models.PartitionIdentity `bson:"inline"`
	Code string          `json:"code" bson:"code" validate:"required"`
	Names *[]models.NameX `json:"names" bson:"names" validate:"required,min=1,unique=Code,dive"`
	VehicleType string          `json:"vehicletype" bson:"vehicletype"`
	LicensePlate string          `json:"licenseplate" bson:"licenseplate"`
	Province string          `json:"province" bson:"province"`
	Brand string          `json:"brand" bson:"brand"`
	Model string          `json:"model" bson:"model"`
	Color string          `json:"color" bson:"color"`
	Year int             `json:"year" bson:"year"`
	CapacityWeight float64         `json:"capacityweight" bson:"capacityweight"`
	CapacityVolume float64         `json:"capacityvolume" bson:"capacityvolume"`
	FuelType string          `json:"fueltype" bson:"fueltype"`
	Status int8            `json:"status" bson:"status"`
	DriverCode string          `json:"drivercode" bson:"drivercode"`
	DriverName string          `json:"drivername" bson:"drivername"`
	InsuranceExpiry time.Time       `json:"insuranceexpiry" bson:"insuranceexpiry"`
	RegistrationExpiry time.Time       `json:"registrationexpiry" bson:"registrationexpiry"`
	Notes string          `json:"notes" bson:"notes"`
	ImageURI string          `json:"imageuri" bson:"imageuri"`
}

type VehicleInfo struct {
	models.DocIdentity `bson:"inline"`
	Vehicle  `bson:"inline"`
}

func (VehicleInfo) CollectionName() string {
	return vehicleCollectionName
}

type VehicleData struct {
	models.ShopIdentity `bson:"inline"`
	VehicleInfo  `bson:"inline"`
}

type VehicleDoc struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	VehicleData  `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (VehicleDoc) CollectionName() string {
	return vehicleCollectionName
}

type VehicleItemGuid struct {
	Code string `json:"code" bson:"code"`
}

func (VehicleItemGuid) CollectionName() string {
	return vehicleCollectionName
}

type VehicleActivity struct {
	VehicleData  `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (VehicleActivity) CollectionName() string {
	return vehicleCollectionName
}

type VehicleDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (VehicleDeleteActivity) CollectionName() string {
	return vehicleCollectionName
}
