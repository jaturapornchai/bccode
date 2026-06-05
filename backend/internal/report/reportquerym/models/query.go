package models

type QueryParam struct {
	Name  string      `json:"name"`
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

type Query struct {
	Collection string       `json:"collection"`
	Filter     string       `json:"filter"`
	Fields     []string     `json:"fields"`
	Params     []QueryParam `json:"params"`
}

type QueryParamRequest struct {
	Name  string      `json:"name"`
	Value interface{} `json:"value"`
}

type DynamicCollection struct {
	Collection string
}

func (d *DynamicCollection) SetCollectionName(collectionName string) {
	d.Collection = collectionName
}

func (d *DynamicCollection) CollectionName() string {
	return d.Collection
}
