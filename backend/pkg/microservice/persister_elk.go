package microservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/config"
	"sync"

	elk "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

type IPersisterElk interface {
	Create(model interface{}) error
	CreateWithID(docID string, model interface{}) error
	Update(docID string, model interface{}) error
	Delete(docID string, model interface{}) error
}

type ElasticModel interface {
	IndexName() string
}

type PersisterElk struct {
	config  config.IPersisterElkConfig
	db      *elk.Client
	dbMutex sync.Mutex
}

func NewPersisterElk(config config.IPersisterElkConfig) *PersisterElk {
	return &PersisterElk{
		config: config,
	}
}

func (pst *PersisterElk) getClient() (*elk.Client, error) {

	if pst.db != nil {
		return pst.db, nil
	}

	pst.dbMutex.Lock()

	cfg := elk.Config{
		Username:  pst.config.Username(),
		Password:  pst.config.Password(),
		Addresses: pst.config.ElkAddress(),
	}

	es, err := elk.NewClient(cfg)

	if err != nil {
		return nil, err
	}

	pst.db = es

	pst.dbMutex.Unlock()

	return pst.db, nil
}

func (pst *PersisterElk) getIndexName(model interface{}) (string, error) {

	modelx, ok := model.(ElasticModel)

	if ok {
		return modelx.IndexName(), nil
	}
	return "", fmt.Errorf("struct is not implement IndexName()")
}

func (pst *PersisterElk) Create(model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}

	txtByte, err := json.Marshal(model)

	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index: indexName,
		Body:  bytes.NewReader(txtByte),
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterElk) CreateWithID(docID string, model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}

	txtByte, err := json.Marshal(model)

	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(txtByte),
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterElk) Update(docID string, model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}

	txtByte, err := json.Marshal(model)

	if err != nil {
		return err
	}

	req := esapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       bytes.NewReader(txtByte),
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterElk) Delete(docID string, model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}
	req := esapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}
