package models

import "time"

type ActivityDoc struct {
	CreatedBy string    `json:"-" bson:"createdby"`
	CreatedAt time.Time `json:"-" bson:"created_at"`
	UpdatedBy string    `json:"-" bson:"updatedby,omitempty"`
	UpdatedAt time.Time `json:"-" bson:"updated_at,omitempty"`
	DeletedBy string    `json:"-" bson:"deleted_by,omitempty"`
	DeletedAt time.Time `json:"-" bson:"deleted_at,omitempty"`
}

type Activity struct {
	CreatedBy string    `json:"createdby" bson:"createdby"`
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedBy string    `json:"updatedby" bson:"updatedby,omitempty"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedBy string    `json:"deleted_by" bson:"deleted_by,omitempty"`
	DeletedAt time.Time `json:"deleted_at" bson:"deleted_at,omitempty"`
}

type ActivityTime struct {
	CreatedAt time.Time `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time `json:"updated_at" bson:"updated_at,omitempty"`
	DeletedAt time.Time `json:"deleted_at" bson:"deleted_at,omitempty"`
}

type LastActivity struct {
	New interface{} `json:"new,omitempty" `
	Remove interface{} `json:"remove,omitempty"`
}

type LastUpdate struct {
	LastUpdatedAt time.Time `json:"lastupdatedat" bson:"lastupdatedat"`
}
