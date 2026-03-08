package models

type Index struct {
	ID       string `json:"id" bson:"_id,omitempty" gorm:"id"`
	Identity `bson:"inline"`
}
