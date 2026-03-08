package models

type PartStatus int8

const (
	PartStatusPending PartStatus = iota
	PartStatusDone
	PartStatusError
	PartStatusNotFound
)

func (s PartStatus) String() string {
	return [...]string{"pending", "processing", "done", "error"}[s]
}
