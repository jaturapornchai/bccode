package models

type ResultModel struct {
	Success     bool   `json:"success" bson:"success"`
	Guid        string `json:"guid" bson:"guid"`
	TotalRecord int    `json:"total_record" bson:"total_record"`
	TotalPage   int    `json:"total_page" bson:"total_page"`
}
