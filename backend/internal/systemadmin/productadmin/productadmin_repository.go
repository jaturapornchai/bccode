package productadmin

import (
	"context"
	"smlcloudplatform/internal/product/productbarcode/models"
	productBarcodeModel "smlcloudplatform/internal/product/productbarcode/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"

	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductAdminMongoRepository interface {
	FindProductBarcodeByHoldingCode(ctx context.Context, holdingCode string) ([]productBarcodeModel.ProductBarcodeDoc, error)
	FindProductAndBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeDoc, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeDoc, mongopagination.PaginationData, error)
	DeleteProductBarcodeByHoldingCode(ctx context.Context, holdingCode string, userName string, ids []string) error
}

type ProductAdminMongoRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.ProductBarcodeDoc]
}

func NewProductAdminMongoRepository(pst microservice.IPersisterMongo) IProductAdminMongoRepository {
	return &ProductAdminMongoRepository{
		pst:              pst,
		SearchRepository: repositories.NewSearchRepository[models.ProductBarcodeDoc](pst),
	}
}

func (r ProductAdminMongoRepository) FindProductBarcodeByHoldingCode(ctx context.Context, holdingCode string) ([]productBarcodeModel.ProductBarcodeDoc, error) {

	docList := []productBarcodeModel.ProductBarcodeDoc{}
	err := r.pst.Find(ctx, &productBarcodeModel.ProductBarcodeDoc{}, bson.M{"holdingcode": holdingCode}, &docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r ProductAdminMongoRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ProductBarcodeDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r ProductAdminMongoRepository) DeleteProductBarcodeByHoldingCode(ctx context.Context, holdingCode string, userName string, ids []string) error {
	// err := r.pst.Delete(&productBarcodeModel.ProductBarcodeDoc{}, bson.M{"holdingcode": holdingCode})
	// if err != nil {
	// 	return err
	// }

	// err := r.pst.DeleteByID(&productBarcodeModel.ProductBarcodeDoc{}, bson.M{"holdingcode": holdingCode})
	// if err != nil {
	// 	return err
	// }

	err := r.pst.SoftBatchDeleteByID(ctx, &productBarcodeModel.ProductBarcodeDoc{}, userName, ids)
	if err != nil {
		return err
	}

	return nil
}

func (r ProductAdminMongoRepository) FindProductAndBarcode(ctx context.Context, holdingCode string, barcode string) (models.ProductBarcodeDoc, error) {

	var doc models.ProductBarcodeDoc
	err := r.pst.FindOne(ctx, &models.ProductBarcodeDoc{}, bson.M{"holdingcode": holdingCode, "barcode": barcode}, &doc)
	if err != nil {
		return models.ProductBarcodeDoc{}, err
	}

	return doc, nil
}
