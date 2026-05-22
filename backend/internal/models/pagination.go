package models

type Pagination struct {
	Total int `json:"total"`
	Page int `json:"page"`
	PerPage int `json:"per_page"`
	TotalPage int `json:"total_page"`
}
