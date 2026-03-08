package models

type ObjectDesign struct {
	Type         int     `json:"type" bson:"type" validate:"required"`
	Width        float64 `json:"width" bson:"width"`
	Height       float64 `json:"height" bson:"height"`
	PositionTop  float64 `json:"positiontop" bson:"positiontop" `
	PositionLeft float64 `json:"positionleft" bson:"positionleft"`
}
