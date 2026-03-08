package images

import (
	"bytes"
	"context"
	"mime/multipart"
	"smlcloudplatform/internal/images/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repo "smlcloudplatform/internal/product/productbarcode/repositories"
	slipimage_repo "smlcloudplatform/internal/slipimage/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

type IImagesService interface {
	UploadImage(shopId string, fh *multipart.FileHeader) (*models.Image, error)
	UploadImageToProduct(shopID string, fh *multipart.FileHeader) error
	GetImageByProductCode(shopid string, itemguid string, index int) (string, *bytes.Buffer, error)
	GetSlipImage(shopid string, posID string, docDate time.Time, docNo string) (string, *bytes.Buffer, error)
}

type ImagesService struct {
	persisterImage *microservice.PersisterImage
	invRepo        productbarcode_repo.IProductBarcodeRepository
	slipimageRepo  slipimage_repo.ISlipImageMongoRepository
	NewGUIDFn      func() string
	contextTimeout time.Duration
}

func NewImageService(persisterImage *microservice.PersisterImage,
	inventoryRepo productbarcode_repo.IProductBarcodeRepository,
	slipimageRepo slipimage_repo.ISlipImageMongoRepository,
) *ImagesService {

	contextTimeout := time.Duration(15) * time.Second
	// check config storage location

	return &ImagesService{
		persisterImage: persisterImage,
		invRepo:        inventoryRepo,
		slipimageRepo:  slipimageRepo,
		NewGUIDFn:      func() string { return utils.NewGUID() },
		contextTimeout: contextTimeout,
	}
}

func (svc ImagesService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ImagesService) UploadImage(shopId string, fh *multipart.FileHeader) (*models.Image, error) {

	fileUploadMetadataSlice := strings.Split(fh.Filename, ".")
	fileName := svc.NewGUIDFn() //fileUploadMetadataSlice[0]
	fileExtension := fileUploadMetadataSlice[len(fileUploadMetadataSlice)-1]

	fileName, err := svc.persisterImage.Upload(fh, shopId+"/"+fileName, fileExtension)

	if err != nil {
		return nil, err
	}

	// create thumbnail bla bla
	image := &models.Image{
		Uri: fileName,
	}

	return image, nil
}

func (svc ImagesService) UploadImageToProduct(shopID string, fh *multipart.FileHeader) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	fileUploadMetadataSlice := strings.Split(fh.Filename, ".")

	if len(fileUploadMetadataSlice) != 2 {
		return errors.New("invalid file name")
	}

	// find product by code
	barcodeFileName := fileUploadMetadataSlice[0]
	fileExtension := fileUploadMetadataSlice[len(fileUploadMetadataSlice)-1]

	findDoc, err := svc.invRepo.FindByBarcode(ctx, shopID, barcodeFileName)
	if err != nil {
		return err
	}

	if len(findDoc.Barcode) == 0 {
		return errors.New("not found product barcode")
	}

	uploadFileName, err := svc.persisterImage.Upload(fh, shopID+"/"+barcodeFileName, fileExtension)
	if err != nil {
		return err
	}

	dataDoc := findDoc

	if len(dataDoc.ImageURI) == 0 {
		dataDoc.ImageURI = uploadFileName
	}

	if dataDoc.Images == nil {
		dataDoc.Images = &[]productbarcode_models.ProductImage{}
	}

	imageXOrder := len(*dataDoc.Images) + 1

	// append image
	*dataDoc.Images = append(*dataDoc.Images, productbarcode_models.ProductImage{
		XOrder: imageXOrder,
		URI:    uploadFileName,
	})

	// save and return
	err = svc.invRepo.Update(context.Background(), shopID, dataDoc.GuidFixed, dataDoc)
	return err
}

func (svc ImagesService) GetImageByProductCode(shopid string, itemguid string, index int) (string, *bytes.Buffer, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.invRepo.FindByGuid(ctx, shopid, itemguid)

	if err != nil {
		return "", nil, err
	}

	if findDoc.Images == nil {
		return "", nil, errors.New("No Image")
	}

	productImage := *findDoc.Images
	inventoryImageLength := len(productImage)
	if inventoryImageLength < index {
		return "", nil, errors.New("Not Found Image")
	}

	imgFileUrl := productImage[index-1].URI

	imageUri, buffer, err := svc.persisterImage.FilePersister.LoadFile(imgFileUrl)
	if err != nil {
		return "", buffer, err
	}

	return imageUri, buffer, nil
}

func (svc ImagesService) GetSlipImage(shopid string, posID string, docDate time.Time, docNo string) (string, *bytes.Buffer, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.slipimageRepo.FindOne(ctx, shopid, bson.M{"posid": posID, "docdate": docDate, "docno": docNo})

	if err != nil {
		return "", nil, err
	}

	if findDoc.URI == "" {
		return "", nil, errors.New("no image")
	}

	imgFileUrl := findDoc.URI

	imageUri, buffer, err := svc.persisterImage.FilePersister.LoadFile(imgFileUrl)
	if err != nil {
		return "", buffer, err
	}

	return imageUri, buffer, nil
}
