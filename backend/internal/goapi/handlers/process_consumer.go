package handlers

// import (
// 	"smlcloudplatform/internal/goapi/logger"
// 	"context"
// 	"database/sql"
// 	mypg "smlcloudplatform/internal/goapi/mypg"
// 	build "smlcloudplatform/internal/goapi/process/build"
// 	processModel "smlcloudplatform/internal/goapi/process/models"
// 	processDoc "smlcloudplatform/internal/goapi/process/process-doc"
// 	processstock "smlcloudplatform/internal/goapi/process/process-stock"
// 	"strconv"
// 	"strings"
// )

// type IProcessDB interface {
// 	GetShopDB(holdingCode string) (*sql.DB, error)
// 	InsertDocList(data []processModel.DocStruct, docRefData []processModel.DocRefStruct)
// 	InsertStockWaitProcess(holdingCode string, barcodes []string) error
// 	InsertDocWaitProcess(holdingCode string, transFlag int, docNo string) error
// }

// type ProcessDB struct {
// 	db *sql.DB
// }

// type IProcessConsumer interface {
// 	DecodeDocDetail(msg string) processModel.ProcessMongoTransModel
// 	ListBarcode(data processModel.ProcessMongoTransModel) []string
// 	ProcessUpSertDocData(holdingCode string, data processModel.ProcessMongoTransModel) error
// 	ProcessDeleteDocData(holdingCode string, data processModel.ProcessMongoTransModel) error
// 	InvokeProcess(holdingcode string)
// }

// type ProcessConsumer struct {
// 	dbProcess IProcessDB
// }

// func NewPurchaseOrderConsumer(dbProcess IProcessDB) IProcessConsumer {
// 	return &ProcessConsumer{
// 		dbProcess: dbProcess,
// 	}
// }

// func (p *ProcessConsumer) DecodeDocDetail(msg string) processModel.ProcessMongoTransModel {
// 	decodeModel := build.DocDetailDecode(msg)
// 	return decodeModel
// }

// func (p *ProcessConsumer) ListBarcode(data processModel.ProcessMongoTransModel) []string {

// 	var barcodeList []string
// 	for _, item := range data.Details {
// 		barcodeList = append(barcodeList, item.Barcode)
// 	}
// 	return barcodeList
// }

// func (p *ProcessConsumer) ProcessUpSertDocData(holdingCode string, data processModel.ProcessMongoTransModel) error {

// 	var postgresData []processModel.DocStruct
// 	var postgresDocRefData []processModel.DocRefStruct

// 	postgresData = append(postgresData, build.MapDocStruct(data))
// 	// ถ้ามีการอ้างอิงเอกสาร ให้เพิ่มใน postgresDocRefData
// 	for _, refNo := range data.DocReferences {
// 		postgresDocRefData = append(postgresDocRefData, build.MapDocRefStruct(data.DocNo, data.TransFlag, refNo))
// 	}

// 	_, err := p.dbProcess.GetShopDB(holdingCode)
// 	if err != nil {
// 		return err
// 	}

// 	p.dbProcess.InsertDocList(postgresData, postgresDocRefData)

// 	barcodeList := p.ListBarcode(data)
// 	err = p.dbProcess.InsertStockWaitProcess(holdingCode, barcodeList)

// 	if err != nil {
// 		return err
// 	}

// 	err = p.dbProcess.InsertDocWaitProcess(holdingCode, data.TransFlag, data.DocNo)
// 	if err != nil {
// 		return err
// 	}

// 	p.InvokeProcess(holdingCode)
// 	return nil
// }

// func (p *ProcessConsumer) ProcessDeleteDocData(holdingCode string, data processModel.ProcessMongoTransModel) error {

// }

// func (d *ProcessConsumer) GetShopDB(holdingCode string) (*sql.DB, error) {
// 	postgresDB, err := mypg.FastConnect(holdingCode)
// 	if err != nil {
// 		logger.Info("Failed to connect to PostgreSQL: %v", err)
// 		return nil, err
// 	}
// 	return postgresDB, nil
// }

// func (p *ProcessConsumer) InvokeProcess(holdingcode string) {
// 	go processstock.ProcessStockCostAll(holdingcode)
// 	go processDoc.ProcessDocPurchaseAll(holdingcode)
// }

// func (p *ProcessDB) GetShopDB(holdingCode string) (*sql.DB, error) {
// 	postgresDB, err := mypg.FastConnect(holdingCode)
// 	if err != nil {
// 		logger.Info("Failed to connect to PostgreSQL: %v", err)
// 		return nil, err
// 	}

// 	p.db = postgresDB
// 	return postgresDB, nil

// }

// func (p *ProcessDB) InsertDocList(data []processModel.DocStruct, docRefData []processModel.DocRefStruct) {
// 	build.ProcessInsertDocList(context.Background(), p.db, data, docRefData)
// }

// func (p *ProcessDB) InsertStockWaitProcess(holdingCode string, barcodes []string) error {

// 	var barcodeInsertList []string
// 	for _, item := range barcodes {
// 		barcodeInsertList = append(barcodeInsertList, "('"+item+"')")
// 	}

// 	_, err := p.db.ExecContext(context.Background(), queryInsertStockWaitProcess)
// 	if err != nil {
// 		logger.Error("inserting into stockwaitprocess table: %v", err)
// 		return err
// 	}
// 	return nil
// }

// func (p *ProcessDB) InsertDocWaitProcess(holdingCode string, transFlag int, docNo string) error {
// 	queryInsertDocWaitProcess := "INSERT INTO docwaitprocess (transflag,docno) VALUES (" + strconv.Itoa(transFlag) + ",'" + docNo + "');"
// 	_, err := p.db.ExecContext(context.Background(), queryInsertDocWaitProcess)
// 	return err
// }
