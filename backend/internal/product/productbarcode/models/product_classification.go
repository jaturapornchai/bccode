package models

import "fmt"

const (
	ItemTypeStock int8 = iota
	ItemTypeService
	ItemTypeSet
	ItemTypeNotStock
)

const (
	MaterialTypeGeneral int8 = iota
	MaterialTypeMaterial
	MaterialTypeSemiFinished
	MaterialTypeSet
	MaterialTypeAgricultural
)

func IsValidItemType(value int8) bool {
	return value >= ItemTypeStock && value <= ItemTypeNotStock
}

func IsValidMaterialType(value int8) bool {
	return value >= MaterialTypeGeneral && value <= MaterialTypeAgricultural
}

func ValidateProductClassification(itemType int8, materialType int8) error {
	if !IsValidItemType(itemType) {
		return fmt.Errorf("invalid itemtype %d: allowed values are 0=Stock, 1=Service, 2=Set, 3=Not Stock", itemType)
	}
	if !IsValidMaterialType(materialType) {
		return fmt.Errorf("invalid materialtype %d: allowed values are 0=General, 1=Material, 2=Semi-Finished, 3=Set, 4=Agricultural", materialType)
	}
	if itemType == ItemTypeSet && materialType != MaterialTypeSet {
		return fmt.Errorf("invalid materialtype %d for itemtype 2 (Set): expected materialtype 3 (Set)", materialType)
	}
	return nil
}
