package datatransfer

import "smlcloudplatform/pkg/microservice"

type IDataTransferConnection interface {
	GetSourceConnection() microservice.IPersisterMongo
	GetTargetConnection() microservice.IPersisterMongo
	TestConnect() (bool, error)
}

type DataTransferConnection struct {
	sourceDatabase microservice.IPersisterMongo
	targetDatabase microservice.IPersisterMongo
}

func NewDataTransferConnection(sourceDatabase microservice.IPersisterMongo, targetDatabase microservice.IPersisterMongo) IDataTransferConnection {
	return &DataTransferConnection{
		sourceDatabase: sourceDatabase,
		targetDatabase: targetDatabase,
	}
}

func (dtc *DataTransferConnection) GetSourceConnection() microservice.IPersisterMongo {
	return dtc.sourceDatabase
}

func (dtc *DataTransferConnection) GetTargetConnection() microservice.IPersisterMongo {
	return dtc.targetDatabase
}

func (dtc *DataTransferConnection) TestConnect() (bool, error) {
	return true, nil
}
