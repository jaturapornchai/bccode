package microservice

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"smlcloudplatform/internal/config"
	"strings"
)

type IPersisterFile interface {
	Save(file *multipart.FileHeader, fileName string, fileExtension string) (string, error)
	LoadFile(fileName string) (string, *bytes.Buffer, error)
}

type PersisterFile struct {
	StoreFilePath string
	StoreDataUri  string
}

type File struct {
	FileName    string
	ContentType string
	Data        []byte
	Size        int
}

func NewPersisterFile(cfg *config.StorageFileConfig) *PersisterFile {
	return &PersisterFile{
		StoreFilePath: cfg.StorageDataPath(),
		StoreDataUri:  cfg.StorageUriAtlas(),
	}
}

func (pst *PersisterFile) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {

	u, err := url.Parse(pst.StoreDataUri)
	if err != nil {
		return "", err
	}

	imageFileName := fmt.Sprintf("%s.%s", fileName, fileExtension)
	file, err := fh.Open()
	if err != nil {
		return "", err
	}
	defer file.Close()

	// upload Now
	tempFile, err := os.CreateTemp(pst.StoreFilePath, "upload-*."+fileExtension)
	if err != nil {
		return "", err
	}
	defer tempFile.Close()

	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}
	// write this byte array to our temporary file
	tempFile.Write(fileBytes)

	// rename
	tmpFileName := tempFile.Name()
	tempFile.Close()

	if fileName != "" {
		//fmt.Println(fileName)
		// dir, _ := filepath.Split(pst.StoreFilePath)
		uploadFileName := filepath.Join(pst.StoreFilePath, imageFileName)
		//fmt.Println(uploadFileName)
		_ = os.Rename(tmpFileName, uploadFileName)
		// if err != nil {
		// 	fmt.Println(err.Error())
		// }
		fileName = uploadFileName
	}

	u.Path = path.Join(u.Path, imageFileName)
	imageUri := u.String()

	return imageUri, nil

	//storeFilePath := filepath.Join(pst.StoreFilePath, fileName)
	// return os.WriteFile(storeFilePath, file.Data, 0600)
}

func (pst *PersisterFile) LoadFile(fileName string) (string, *bytes.Buffer, error) {

	imgFileName := strings.Replace(fileName, pst.StoreDataUri, "", -1)
	storateFileName := filepath.Join(pst.StoreFilePath, imgFileName)

	return storateFileName, nil, nil
}
