package models

type ResultModel struct {
	Success     bool   `json:"success" bson:"success"`
	Guid        string `json:"guid" bson:"guid"`
	TotalRecord int    `json:"totalrecord" bson:"totalrecord"`
	TotalPage   int    `json:"totalpage" bson:"totalpage"`
}
