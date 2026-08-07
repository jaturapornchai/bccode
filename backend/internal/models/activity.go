package models

import "time"

type ActivityDoc struct {
	CreatedBy     string    `json:"-" bson:"createdby"`
	CreatedByName string    `json:"-" bson:"createdbyname,omitempty"`
	CreatedAt     time.Time `json:"-" bson:"createdat"`
	UpdatedBy     string    `json:"-" bson:"updatedby,omitempty"`
	UpdatedByName string    `json:"-" bson:"updatedbyname,omitempty"`
	UpdatedAt     time.Time `json:"-" bson:"updatedat,omitempty"`
	DeletedBy     string    `json:"-" bson:"deletedby,omitempty"`
	DeletedByName string    `json:"-" bson:"deletedbyname,omitempty"`
	DeletedAt     time.Time `json:"-" bson:"deletedat,omitempty"`
}

type Activity struct {
	CreatedBy     string    `json:"createdby" bson:"createdby"`
	CreatedByName string    `json:"createdbyname,omitempty" bson:"createdbyname,omitempty"`
	CreatedAt     time.Time `json:"createdat" bson:"createdat"`
	UpdatedBy     string    `json:"updatedby" bson:"updatedby,omitempty"`
	UpdatedByName string    `json:"updatedbyname,omitempty" bson:"updatedbyname,omitempty"`
	UpdatedAt     time.Time `json:"updatedat" bson:"updatedat,omitempty"`
	DeletedBy     string    `json:"deletedby" bson:"deletedby,omitempty"`
	DeletedByName string    `json:"deletedbyname,omitempty" bson:"deletedbyname,omitempty"`
	DeletedAt     time.Time `json:"deletedat" bson:"deletedat,omitempty"`
}

type ActivityTime struct {
	CreatedAt time.Time `json:"createdat" bson:"createdat"`
	UpdatedAt time.Time `json:"updatedat" bson:"updatedat,omitempty"`
	DeletedAt time.Time `json:"deletedat" bson:"deletedat,omitempty"`
}

type LastActivity struct {
	New    interface{} `json:"new,omitempty" `
	Remove interface{} `json:"remove,omitempty"`
}

type LastUpdate struct {
	LastUpdatedAt time.Time `json:"lastupdatedat" bson:"lastupdatedat"`
}
