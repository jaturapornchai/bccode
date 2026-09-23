package models

import "time"

type ActivityDoc struct {
	CreatedBy     string    `json:"-"`
	CreatedByName string    `json:"-"`
	CreatedAt     time.Time `json:"-"`
	UpdatedBy     string    `json:"-"`
	UpdatedByName string    `json:"-"`
	UpdatedAt     time.Time `json:"-"`
	DeletedBy     string    `json:"-"`
	DeletedByName string    `json:"-"`
	DeletedAt     time.Time `json:"-"`
}

type Activity struct {
	CreatedBy     string    `json:"createdby"`
	CreatedByName string    `json:"createdbyname,omitempty"`
	CreatedAt     time.Time `json:"createdat"`
	UpdatedBy     string    `json:"updatedby"`
	UpdatedByName string    `json:"updatedbyname,omitempty"`
	UpdatedAt     time.Time `json:"updatedat"`
	DeletedBy     string    `json:"deletedby"`
	DeletedByName string    `json:"deletedbyname,omitempty"`
	DeletedAt     time.Time `json:"deletedat"`
}

type ActivityTime struct {
	CreatedAt time.Time `json:"createdat"`
	UpdatedAt time.Time `json:"updatedat"`
	DeletedAt time.Time `json:"deletedat"`
}

type LastActivity struct {
	New    interface{} `json:"new,omitempty" `
	Remove interface{} `json:"remove,omitempty"`
}

type LastUpdate struct {
	LastUpdatedAt time.Time `json:"lastupdatedat"`
}
