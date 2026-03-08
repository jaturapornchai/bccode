package servicetools

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/pkg/microservice"
	"time"

	commonModel "smlcloudplatform/internal/models"
	shopModel "smlcloudplatform/internal/shop/models"
	accountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	chartofaccountrepositories "smlcloudplatform/internal/vfgl/chartofaccount/repositories"

	journalBookModel "smlcloudplatform/internal/vfgl/journalbook/models"
	journalbookRepo "smlcloudplatform/internal/vfgl/journalbook/repositories"
)

type IServiceTools interface {
	RegisterHttp(ms *microservice.Microservice, pathPrefix string)
	InitMasterCenterShop() error
	InitChartOfAccountMasterCenter() error
	InitJournalBookMasterCenter(ctx microservice.IContext) error
}

type ServiceTools struct {
	logger           logger.ILogger
	masterShop       shopModel.ShopDoc
	cfg              config.IConfig
	persisterMongodb microservice.IPersisterMongo
	timeoutDuration  time.Duration
}

func NewServiceTools(logger logger.ILogger, cfg config.IConfig, persisterMongodb microservice.IPersisterMongo) IServiceTools {
	return &ServiceTools{
		logger:           logger,
		masterShop:       MasterShop(),
		cfg:              cfg,
		persisterMongodb: persisterMongodb,
		timeoutDuration:  time.Duration(30) * time.Second,
	}
}

func (t *ServiceTools) RegisterHttp(ms *microservice.Microservice, pathPrefix string) {
	ms.POST(pathPrefix+"/servicetools/initMasterJournalBookCenter", t.InitJournalBookMasterCenter)
}

func (t *ServiceTools) InitMasterCenterShop() error {

	shopRepo := shop.NewShopRepository(t.persisterMongodb)
	shopObject, err := shopRepo.FindByGuid(context.TODO(), t.masterShop.GuidFixed)
	if err != nil {
		t.logger.Error(" Error Find Master Shop ", err)
		return err
	}

	if shopObject.GuidFixed == "" {

		t.logger.Info("Create Master Shop")

		_, err := shopRepo.Create(context.Background(), t.masterShop)
		if err != nil {
			t.logger.Error(" Error Create Shop ", err)
			return err
		}
	} else {
		t.logger.Info("Master Shop is Already")
	}

	return nil
}

func (t *ServiceTools) InitChartOfAccountMasterCenter() error {

	charts := ListMasterCharts()

	chartRepo := chartofaccountrepositories.NewChartOfAccountRepository(t.persisterMongodb)
	for i := 0; i < len(charts); i++ {
		t.logger.Infof("Process Chart %s:%s", charts[i].AccountCode, charts[i].AccountName)

		findAccount, err := chartRepo.FindByGuid(context.TODO(), t.masterShop.GuidFixed, charts[i].AccountCode)
		if err != nil {
			t.logger.Errorf("Error Find Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
		}

		if findAccount.GuidFixed == "" {
			chartDoc := accountModel.ChartOfAccountDoc{
				ChartOfAccountData: accountModel.ChartOfAccountData{
					ShopIdentity: models.ShopIdentity{
						ShopID: t.masterShop.GuidFixed,
					},
					ChartOfAccountInfo: accountModel.ChartOfAccountInfo{
						DocIdentity: models.DocIdentity{
							GuidFixed: charts[i].AccountCode,
						},
						ChartOfAccount: accountModel.ChartOfAccount{
							AccountCode:            charts[i].AccountCode,
							AccountName:            charts[i].AccountName,
							AccountCategory:        charts[i].AccountCategory,
							AccountBalanceType:     charts[i].AccountBalanceType,
							AccountGroup:           charts[i].AccountGroup,
							AccountLevel:           charts[i].AccountLevel,
							ConsolidateAccountCode: charts[i].ConsolidateAccountCode,
							// ISCenterChart:          true,
						},
					},
				},
			}

			_, err := chartRepo.Create(context.Background(), chartDoc)
			if err != nil {
				t.logger.Errorf("Error Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			} else {
				t.logger.Infof("Create Account %s:%s", charts[i].AccountCode, charts[i].AccountName)
			}

		} else {
			//logger.Infof("Account %s:%s is Already", charts[i].AccountCode, charts[i].AccountName)
		}
	}

	return nil
}

func (t *ServiceTools) InitJournalBookMasterCenter(ctx microservice.IContext) error {

	books := MasterJournalBook()
	bookRepo := journalbookRepo.NewJournalBookMongoRepository(t.persisterMongodb)

	for i := 0; i < len(books); i++ {
		findBook, err := bookRepo.FindByGuid(context.TODO(), t.masterShop.GuidFixed, books[i].Code)
		if err != nil {
			return err
		}

		if findBook.GuidFixed == "" {
			bookRepo.Create(context.Background(),
				journalBookModel.JournalBookDoc{
					JournalBookData: journalBookModel.JournalBookData{
						ShopIdentity: models.ShopIdentity{
							ShopID: t.masterShop.GuidFixed,
						},
						JournalBookInfo: journalBookModel.JournalBookInfo{
							DocIdentity: models.DocIdentity{
								GuidFixed: books[i].Code,
							},
							JournalBook: books[i],
						},
					},
				},
			)
		}
	}

	ctx.Response(http.StatusOK, commonModel.ResponseSuccess{
		Success: true,
	})
	return nil
}

func MasterShop() shopModel.ShopDoc {
	return shopModel.ShopDoc{
		ShopInfo: shopModel.ShopInfo{
			DocIdentity: models.DocIdentity{
				GuidFixed: "999999999",
			},
			Shop: shopModel.Shop{
				Name1: "Master Shop",
			},
		},
	}
}

func ListMasterCharts() []accountModel.ChartOfAccount {

	file, _ := ioutil.ReadFile("./cmd/create-master-center/account_codes.json")
	accounts := []accountModel.ChartOfAccount{}

	_ = json.Unmarshal([]byte(file), &accounts)

	return accounts
}

func MasterJournalBook() []journalBookModel.JournalBook {

	book1 := journalBookModel.JournalBook{
		Code: "1",
		Name: commonModel.Name{
			Name1: "สมุดรายวันทั่วไป",
		},
		// ISCenterBook: true,
	}

	book2 := journalBookModel.JournalBook{
		Code: "2",
		Name: commonModel.Name{
			Name1: "สมุดเงินสดรับ",
		},
		// ISCenterBook: true,
	}

	book3 := journalBookModel.JournalBook{
		Code: "3",
		Name: commonModel.Name{
			Name1: "สมุดเงินสดจ่าย",
		},
		// ISCenterBook: true,
	}

	book4 := journalBookModel.JournalBook{
		Code: "4",
		Name: commonModel.Name{
			Name1: "สมุดรายวันขาย",
		},
		// ISCenterBook: true,
	}

	book5 := journalBookModel.JournalBook{
		Code: "5",
		Name: commonModel.Name{
			Name1: "สมุดรายวันซื้อ",
		},
		// ISCenterBook: true,
	}

	journalbooks := []journalBookModel.JournalBook{
		book1,
		book2,
		book3,
		book4,
		book5,
	}

	return journalbooks
}
