package kafka

import (
	"fmt"
	"smlcloudplatform/internal/product/projection"
	"smlcloudplatform/internal/utils"
	"strings"
)

type MongoProductModel struct {
	HoldingCode  string `json:"holdingcode" bson:"holdingcode"`
	BusinessCode string `json:"businesscode" bson:"businesscode"`
	GuidFixed    string `json:"guidfixed" bson:"guidfixed"`
	Code         string `json:"code" bson:"code"`
	UnitCode     string `json:"unitcode" bson:"unitcode"`
	Names        []struct {
		Name *string `json:"name" bson:"name"`
	} `json:"names" bson:"names"`
	UnitNames []struct {
		Name *string `json:"name" bson:"name"`
	} `json:"unitnames" bson:"unitnames"`
	Projection *projection.Metadata `json:"_projection,omitempty" bson:"-"`
}

func normalizeProductSignal(p *MongoProductModel) error {
	p.HoldingCode = strings.TrimSpace(p.HoldingCode)
	p.BusinessCode = utils.NormalizeBusinessCode(p.BusinessCode)
	p.Code = utils.NormalizeBusinessCode(p.Code)
	if p.HoldingCode == "" || p.BusinessCode == "" || p.Code == "" {
		return fmt.Errorf("%w: holdingcode, businesscode and code are required for product signal", projection.ErrRejected)
	}
	return nil
}

// All three topics are signals to reconcile the same company-scoped row.
// A delayed delete must not delete a newly-created incarnation of that code.
func OnConsumeMessageProductCreateOrUpdate(msg string) error { return consumeProductSignal(msg) }
func OnConsumeMessageProductDelete(msg string) error         { return consumeProductSignal(msg) }
