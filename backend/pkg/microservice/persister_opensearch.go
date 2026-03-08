package microservice

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/config"
	"strings"
	"sync"

	"github.com/opensearch-project/opensearch-go"
	"github.com/opensearch-project/opensearch-go/opensearchapi"
)

type IPersisterOpenSearch interface {
	Create(model interface{}) error
	CreateWithID(docID string, model interface{}) error
	Update(docID string, model interface{}) error
	Delete(docID string, model interface{}) error
}

type OpenSearchModel interface {
	IndexName() string
}

type PersisterOpenSearch struct {
	config  config.IPersisterOpenSearchConfig
	db      *opensearch.Client
	dbMutex sync.Mutex
}

func NewPersisterOpenSearch(config config.IPersisterOpenSearchConfig) *PersisterOpenSearch {
	return &PersisterOpenSearch{
		config: config,
	}
}

func (pst *PersisterOpenSearch) getClient() (*opensearch.Client, error) {

	if pst.db != nil {
		return pst.db, nil
	}

	pst.dbMutex.Lock()

	cfg := opensearch.Config{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Username:  pst.config.Username(),
		Password:  pst.config.Password(),
		Addresses: pst.config.Address(),
	}

	es, err := opensearch.NewClient(cfg)

	if err != nil {
		return nil, err
	}

	pst.db = es

	pst.dbMutex.Unlock()

	return pst.db, nil
}

func (pst *PersisterOpenSearch) getIndexName(model interface{}) (string, error) {

	modelx, ok := model.(ElasticModel)

	if ok {
		return modelx.IndexName(), nil
	}
	return "", fmt.Errorf("struct is not implement IndexName() string")
}

func (pst *PersisterOpenSearch) Create(model interface{}) error {
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

	req := opensearchapi.IndexRequest{
		Index: indexName,
		Body:  bytes.NewReader(txtByte),
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterOpenSearch) CreateWithID(docID string, model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}

	txtByte, err := json.Marshal(model)
	document := strings.NewReader(string(txtByte))
	if err != nil {
		return err
	}

	req := opensearchapi.IndexRequest{
		Index:      indexName,
		DocumentID: docID,
		Body:       document,
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterOpenSearch) Update(docID string, model interface{}) error {
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

	req := opensearchapi.IndexRequest{
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

func (pst *PersisterOpenSearch) Delete(docID string, model interface{}) error {
	indexName, err := pst.getIndexName(model)
	if err != nil {
		return err
	}

	db, err := pst.getClient()

	if err != nil {
		return err
	}
	req := opensearchapi.DeleteRequest{
		Index:      indexName,
		DocumentID: docID,
	}

	_, err = req.Do(context.Background(), db)

	if err != nil {
		return err
	}

	return nil
}
