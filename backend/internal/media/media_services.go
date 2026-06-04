package media

import (
	"mime"
	"mime/multipart"
	"path/filepath"
	"smlcloudplatform/internal/media/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
)

type IMediaService interface {
	UploadImage(holdingCode string, fh *multipart.FileHeader) (*models.Media, error)
	UploadVideo(holdingCode string, fh *multipart.FileHeader) (*models.Media, error)
}

type MediaService struct {
	filePersister microservice.IPersisterFile
	NewGUIDFn     func() string
}

func NewMediaService(filePersister microservice.IPersisterFile, newGuidFn func() string) *MediaService {

	return &MediaService{
		filePersister: filePersister,
		NewGUIDFn:     newGuidFn,
	}
}

func InitMediaService() IMediaService {

	filePersister := microservice.NewFilePersister()
	fnGenGuid := func() string { return utils.NewGUID() }
	return NewMediaService(filePersister, fnGenGuid)
}

func (svc MediaService) UploadImage(holdingCode string, fh *multipart.FileHeader) (*models.Media, error) {

	// fileUploadMetadataSlice := strings.Split(fh.Filename, ".")
	// fileName := svc.NewGUIDFn() //fileUploadMetadataSlice[0]

	// imageUri, err := svc.filePersister.Save(fh, fileName, fileUploadMetadataSlice[1])
	// if err != nil {
	// 	return nil, err
	// }

	// return &models.Media{
	// 	HoldingCode: holdingCode,
	// 	Uri:    imageUri,
	// }, nil

	return nil, nil
}

func (svc MediaService) UploadVideo(holdingCode string, fh *multipart.FileHeader) (*models.Media, error) {

	fileExt := ""

	contentType := fh.Header.Get("Content-Type")
	extensions, err := mime.ExtensionsByType(contentType)
	if err != nil {
		return nil, err
	}

	if len(extensions) > 0 {
		fileExt = extensions[0]
	}

	// If no extension was found, fall back to using the original file name
	ext := filepath.Ext(fh.Filename)
	if ext != "" {
		fileExt = strings.ToLower(ext)
	}

	fileExt = strings.TrimPrefix(fileExt, ".")
	fileName := svc.NewGUIDFn()

	uploadFileName, err := svc.filePersister.Save(fh, holdingCode+"/"+fileName, fileExt)

	if err != nil {
		return nil, err
	}

	return &models.Media{
		Uri:       uploadFileName,
		Size:      fh.Size,
		Extension: fileExt,
	}, nil

}
