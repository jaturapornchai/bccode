package mogoutil

import (
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

func AggregatePageDecode[T any](paginatedData *mongopagination.PaginatedData) ([]T, error) {
	var aggList []T
	for _, raw := range paginatedData.Data {
		var item *T
		marshallErr := bson.Unmarshal(raw, &item)

		if marshallErr != nil {
			return aggList, marshallErr
		}
		aggList = append(aggList, *item)
	}

	if aggList == nil {
		aggList = make([]T, 0)
	}

	return aggList, nil
}
