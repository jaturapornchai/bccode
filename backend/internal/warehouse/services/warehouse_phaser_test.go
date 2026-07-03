package services_test

import (
	"smlcloudplatform/internal/warehouse/services"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreditPaymentTransactionPhaser(t *testing.T) {

	giveInput := `{
		"guidfixed": "guid001",
		"holdingcode": "shop001",
		"code": "code001",
		"names": [
			{
				"code": "en",
				"name": "warehouse name 001"
			}
		],
		"latitude": 13.75,
		"longitude": 100.5
		}`

	phaser := services.WarehousePhaser{}
	got, err := phaser.PhaseSingleDoc(giveInput)

	assert.Nil(t, err)
	assert.Equal(t, "guid001", got.GuidFixed)
	assert.Equal(t, "shop001", got.HoldingCode)
	assert.Equal(t, "code001", got.Code)
	assert.Equal(t, "warehouse name 001", *(got.Names)[0].Name)
	assert.Equal(t, 13.75, got.Latitude)
	assert.Equal(t, 100.5, got.Longitude)

}
