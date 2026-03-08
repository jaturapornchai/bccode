package utils_test

import (
	"smlcloudplatform/internal/utils"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestPage(t *testing.T) {
	page, limit := utils.GetPageParam(getParam)

	assert.Equal(t, 2, page)
	assert.Equal(t, 20, limit)
}

func getParam(p string) string {
	if p == "page" {
		return "2"
	} else if p == "limit" {
		return "20"
	}
	return ""
}
