package models

type Pageable struct {
	Query string   `json:"q,omitempty"`
	Page  int      `json:"page"`
	Limit int      `json:"limit,omitempty"`
	Sorts []KeyInt `json:"sorts,omitempty"`
}

func (p *Pageable) GetOffest() int {
	if p.Page <= 1 {
		return 0
	}
	return (p.Page - 1) * p.Limit
}

type PageableStep struct {
	Query string   `json:"q,omitempty"`
	Skip  int      `json:"skip,omitempty"`
	Limit int      `json:"limit,omitempty"`
	Sorts []KeyInt `json:"sorts,omitempty"`
}

type KeyInt struct {
	Key   string
	Value int8
}
