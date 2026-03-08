package repositories

import (
	"mime/multipart"
	"smlcloudplatform/pkg/microservice"
)

type ISlipImageStorageImageRepository interface {
	Upload(fh *multipart.FileHeader, fileName string, fileExt string) (string, error)
}

type SlipImageStorageImageRepository struct {
	pst *microservice.PersisterImage
}

func NewSlipImageStorageImageRepository(pst *microservice.PersisterImage) *SlipImageStorageImageRepository {

	insRepo := &SlipImageStorageImageRepository{
		pst: pst,
	}

	return insRepo
}

func (repo SlipImageStorageImageRepository) Upload(fh *multipart.FileHeader, fileName string, fileExt string) (string, error) {
	return repo.pst.Upload(fh, fileName, fileExt)
}
