package services

import (
	"encoding/json"
	"smlcloudplatform/internal/warehouse/models"
)

type WarehousePhaser struct{}

// implement all method in ITransactionPhaser
func (c WarehousePhaser) PhaseSingleDoc(input string) (*models.WarehousePG, error) {

	doc := models.WarehouseDoc{}
	err := json.Unmarshal([]byte(input), &doc)
	if err != nil {
		return nil, err
	}

	pgDoc, err := c.PhaseWarehouseDoc(doc)
	if err != nil {
		return nil, err
	}

	return pgDoc, nil
}

func (c WarehousePhaser) PhaseMultipleDoc(input string) (*[]models.WarehousePG, error) {

	docs := []models.WarehouseDoc{}
	err := json.Unmarshal([]byte(input), &docs)
	if err != nil {
		return nil, err
	}

	pgDocs := []models.WarehousePG{}
	for _, doc := range docs {
		transaction, err := c.PhaseWarehouseDoc(doc)
		if err != nil {
			return nil, err
		}
		pgDocs = append(pgDocs, *transaction)
	}

	return &pgDocs, nil
}

func (c *WarehousePhaser) PhaseWarehouseDoc(doc models.WarehouseDoc) (*models.WarehousePG, error) {

	dataDoc := models.WarehousePG{}

	dataDoc.ShopID = doc.ShopID
	dataDoc.GuidFixed = doc.GuidFixed
	dataDoc.Code = doc.Code
	dataDoc.Names = *doc.Names
	dataDoc.Location = *doc.Location

	return &dataDoc, nil
}
