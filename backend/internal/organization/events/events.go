package events

import (
	"context"
	"time"

	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const OutboxCollectionName = "organizationcrudoutbox"

type OutboxEventDoc struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	HoldingCode string             `json:"holdingcode" bson:"holdingcode"`
	GuidFixed   string             `json:"guidfixed" bson:"guidfixed"`
	Module      string             `json:"module" bson:"module"`
	Action      string             `json:"action" bson:"action"`
	Topic       string             `json:"topic" bson:"topic"`
	Key         string             `json:"key" bson:"key"`
	Payload     interface{}        `json:"payload" bson:"payload"`
	Status      string             `json:"status" bson:"status"`
	Error       string             `json:"error,omitempty" bson:"error,omitempty"`
	CreatedAt   time.Time          `json:"createdat" bson:"createdat"`
	UpdatedAt   time.Time          `json:"updatedat" bson:"updatedat"`
}

func (OutboxEventDoc) CollectionName() string {
	return OutboxCollectionName
}

func PublishOrOutbox(ctx context.Context, pst microservice.IPersisterMongo, prod microservice.IProducer, holdingCode string, module string, action string, topic string, key string, payload interface{}) (string, error) {
	if err := prod.SendMessage(topic, key, payload); err != nil {
		now := time.Now()
		outbox := OutboxEventDoc{
			HoldingCode: holdingCode,
			GuidFixed:   primitive.NewObjectID().Hex(),
			Module:      module,
			Action:      action,
			Topic:       topic,
			Key:         key,
			Payload:     payload,
			Status:      "pending",
			Error:       err.Error(),
			CreatedAt:   now,
			UpdatedAt:   now,
		}
		if _, outboxErr := pst.Create(ctx, OutboxEventDoc{}, outbox); outboxErr != nil {
			return "failed", outboxErr
		}
		return "outbox", nil
	}
	return "published", nil
}
