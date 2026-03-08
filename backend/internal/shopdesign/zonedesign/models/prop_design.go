package models

type PropDesign struct {
	Index        int `json:"index" bson:"index" validate:"required"`
	ObjectDesign `bson:"inline"`
	Label        string `json:"label" bson:"label"`
}
