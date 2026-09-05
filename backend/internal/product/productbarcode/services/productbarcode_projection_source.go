package services

import (
	"context"

	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"smlcloudplatform/pkg/microservice"
)

// The consumer's source lookup must not inherit a secondary read preference:
// an older replica snapshot could overwrite an already reconciled projection.
type primaryProjectionReads struct {
	microservice.IPersisterMongo
}

func (p primaryProjectionReads) Find(ctx context.Context, model, filter, decode interface{}, opts ...*options.FindOptions) error {
	collection, err := p.Exec(ctx, model)
	if err != nil {
		return err
	}
	collection, err = collection.Clone(options.Collection().SetReadPreference(readpref.Primary()))
	if err != nil {
		return err
	}
	cursor, err := collection.Find(ctx, filter, opts...)
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)
	return cursor.All(ctx, decode)
}
