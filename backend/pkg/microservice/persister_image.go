package microservice

import (
	"mime/multipart"
)

type IPersisterImage interface {
	CreateThumbnail() error
	Upload(fh *multipart.FileHeader) (string, error)
}

type PersisterImage struct {
	FilePersister IPersisterFile
}

func NewPersisterImage(filePersister IPersisterFile) *PersisterImage {
	return &PersisterImage{
		FilePersister: filePersister,
	}
}

func (pst *PersisterImage) Upload(fh *multipart.FileHeader, fileName string, fileExt string) (string, error) {

	//
	// u, err := url.Parse(pst.FilePersister.StoreDataUri)
	// if err != nil {
	// 	return "", err
	// }

	//uniqueId := uuid.New()
	//filename := strings.Replace(uniqueId.String(), "-", "", -1)
	//fileExt := strings.Split(fh.Filename, ".")[1]
	//imageFileName := fmt.Sprintf("%s.%s", filename, fileExt)

	imageUri, err := pst.FilePersister.Save(fh, fileName, fileExt)
	if err != nil {
		return "", err
	}
	// u.Path = path.Join(u.Path, imageFileName)
	// imageUri := u.String()

	return imageUri, nil
}
