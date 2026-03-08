package datatransfer

import "os"

type SourceDatabaseConfig struct{}

func (SourceDatabaseConfig) MongodbURI() string {
	return os.Getenv("MONGODB_SOURCE_CONNECTION")
}

func (SourceDatabaseConfig) DB() string {
	return os.Getenv("MONGODB_SOURCE_DATABASE")
}

func (SourceDatabaseConfig) Debug() bool {
	return false
}

type DestinationDatabaseConfig struct{}

func (DestinationDatabaseConfig) MongodbURI() string {
	return os.Getenv("MONGODB_DESTINATION_CONNECTION")
}

func (DestinationDatabaseConfig) DB() string {
	return os.Getenv("MONGODB_DESTINATION_DATABASE")
}

func (DestinationDatabaseConfig) Debug() bool {
	return false
}
