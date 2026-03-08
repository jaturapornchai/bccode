package microservice

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/storage/azblob"
	"github.com/apex/log"
)

// implement PersisterFile Interface

type PersisterAzureBlob struct{}

func NewPersisterAzureBlob() *PersisterAzureBlob {
	return &PersisterAzureBlob{}
}

func (p *PersisterAzureBlob) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {

	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	imageFileName := fmt.Sprintf("%s.%s", fileName, fileExtension)

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	AZURE_STORAGE_ACCOUNT_NAME := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	AZURE_STORAGE_ACCOUNT_KEY := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	AZURE_STORAGE_CONTAINER_NAME := os.Getenv("AZURE_STORAGE_CONTAINER_NAME")

	url := fmt.Sprintf("https://%s.blob.core.windows.net/", AZURE_STORAGE_ACCOUNT_NAME)

	cred, err := azblob.NewSharedKeyCredential(AZURE_STORAGE_ACCOUNT_NAME, AZURE_STORAGE_ACCOUNT_KEY)
	if err != nil {
		log.Fatal("Invalid credentials with error: " + err.Error())
	}

	client, err := azblob.NewClientWithSharedKeyCredential(url, cred, nil)
	if err != nil {
		log.Fatal("Invalid Service Client with error: " + err.Error())
	}

	context := context.TODO()
	_, err = client.UploadBuffer(context, AZURE_STORAGE_CONTAINER_NAME, imageFileName, fileBytes, nil)
	if err != nil {
		return "", err
	}

	blobUri := fmt.Sprintf("%s%s/%s", url, AZURE_STORAGE_CONTAINER_NAME, imageFileName) //.URL()

	return blobUri, nil
}

func (p *PersisterAzureBlob) LoadFile(fileName string) (string, *bytes.Buffer, error) {
	// imgFileName := strings.Replace(fileName, pst.StoreDataUri, "", -1)
	// storageFileName := filepath.Join(pst.StoreFilePath, imgFileName)

	if strings.HasPrefix(fileName, "http") {
		resp, err := http.Get(fileName)
		if err != nil {
			return "", nil, err
		}
		defer resp.Body.Close()

		downloadedData := &bytes.Buffer{}
		downloadedData.ReadFrom(resp.Body)

		return fileName, downloadedData, nil
	}

	// download blob from azure
	AZURE_STORAGE_ACCOUNT_NAME := os.Getenv("AZURE_STORAGE_ACCOUNT_NAME")
	AZURE_STORAGE_ACCOUNT_KEY := os.Getenv("AZURE_STORAGE_ACCOUNT_KEY")
	AZURE_STORAGE_CONTAINER_NAME := os.Getenv("AZURE_STORAGE_CONTAINER_NAME")

	url := fmt.Sprintf("https://%s.blob.core.windows.net/", AZURE_STORAGE_ACCOUNT_NAME)

	credential, err := azblob.NewSharedKeyCredential(AZURE_STORAGE_ACCOUNT_NAME, AZURE_STORAGE_ACCOUNT_KEY)
	if err != nil {
		return "", nil, errors.New("Invalid credentials with error: " + err.Error())
	}

	client, err := azblob.NewClientWithSharedKeyCredential(url, credential, nil)
	if err != nil {
		return "", nil, errors.New("Invalid Service Client with error: " + err.Error())
	}

	context := context.TODO()
	get, err := client.DownloadStream(context, AZURE_STORAGE_CONTAINER_NAME, fileName, nil)
	if err != nil {
		return "", nil, errors.New(err.Error())
	}
	downloadedData := &bytes.Buffer{}
	_, err = downloadedData.ReadFrom(get.Body)
	if err != nil {
		return "", nil, errors.New(err.Error())
	}

	return fileName, downloadedData, nil
}
