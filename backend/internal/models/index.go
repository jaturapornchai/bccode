package models

type Index struct {
	ID       string `json:"id" gorm:"id"`
	Identity
}
