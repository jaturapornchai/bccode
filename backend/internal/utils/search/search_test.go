package search_test

import (
	"fmt"
	"smlcloudplatform/internal/utils/search"
	"testing"
)

type TestModel struct {
}

func TestSearchRepositories(t *testing.T) {
	xx := search.CreateTextFilter([]string{"name_1", "sx"}, "okลาก่อน")
	fmt.Println(xx)
}
