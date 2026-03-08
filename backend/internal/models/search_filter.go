package models

type SearchFilter struct {
	Field string `json:"field" bson:"field"`
	Type  string `json:"type" bson:"type"`
}
