package services_test

import (
	"smlcloudplatform/internal/warehouse/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreditPaymentTransactionPhaser(t *testing.T) {

	giveInput := `{
		"guid_fixed": "guid001",
		"shopid": "shop001",
		"code": "code001",
		"names": [
			{
				"code": "en",
				"name": "warehouse name 001"
			}
		],
		"location": [
			{
				"code": "loc001",
				"names": [
					{
						"code": "en",
						"name": "loc name 001"
					}
				],
				"shelf": [
					{
						"code": "shelf001",
						"name": "shelf name 001"
					}
				]
			}
		]
		}`

	phaser := services.WarehousePhaser{}
	got, err := phaser.PhaseSingleDoc(giveInput)

	assert.Nil(t, err)
	assert.Equal(t, "guid001", got.GuidFixed)
	assert.Equal(t, "shop001", got.ShopID)
	assert.Equal(t, "code001", got.Code)
	assert.Equal(t, "warehouse name 001", *(got.Names)[0].Name)
	assert.Equal(t, "loc001", got.Location[0].Code)
	assert.Equal(t, "loc name 001", *(*got.Location[0].Names)[0].Name)
	assert.Equal(t, "shelf001", got.Location[0].Shelf[0].Code)
	assert.Equal(t, "shelf name 001", got.Location[0].Shelf[0].Name)

}
