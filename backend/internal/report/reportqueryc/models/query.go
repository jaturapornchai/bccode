package models

type QueryParam struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Query struct {
	SQL    string       `json:"sql"`
	Params []QueryParam `json:"params"`
}

type QueryParamRequest struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}
