package services_test

/*
type MockDocumentImageRepository struct {
	mock.Mock
}

func (m *MockDocumentImageRepository) Minus(a int, b int) (int, error) {
	args := m.Called(a, b)
	return args.Int(0), args.Error(1)
}

func (m *MockDocumentImageRepository) Create(ctx context.Context, doc models.DocumentImageDoc) (string, error) {
	args := m.Called(doc)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentImageRepository) Update(shopID string, guid string, doc models.DocumentImageDoc) error {
	args := m.Called(shopID, guid, doc)
	return args.Error(0)
}

func (m *MockDocumentImageRepository) DeleteByGuidfixed(shopID string, guid string, username string) error {
	args := m.Called(shopID, guid, username)
	return args.Error(0)
}

func (m *MockDocumentImageRepository) FindOne(shopID string, filters map[string]interface{}) (models.DocumentImageDoc, error) {
	args := m.Called(shopID, filters)
	return args.Get(0).(models.DocumentImageDoc), args.Error(1)
}

func (m *MockDocumentImageRepository) FindByGuid(shopID string, guid string) (models.DocumentImageDoc, error) {
	args := m.Called(shopID, guid)
	return args.Get(0).(models.DocumentImageDoc), args.Error(1)
}

func (m *MockDocumentImageRepository) FindPage(shopID string, searchInFields []string, pageable micromodels.Pageable)([]models.DocumentImageInfo, mongopagination.PaginationData, error) {
	args := m.Called(shopID, searchInFields, pageable)
	return args.Get(0).([]models.DocumentImageInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockDocumentImageRepository) FindPageFilterSort(shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageInfo, mongopagination.PaginationData, error) {
	args := m.Called(shopID, filters, searchInFields, pageables)
	return args.Get(0).([]models.DocumentImageInfo), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockDocumentImageRepository) SaveDocumentImageDocRefGroup(shopID string, docRef string, docImages []string) error {
	args := m.Called(shopID, docRef, docImages)
	return args.Error(0)
}

func (m *MockDocumentImageRepository) ListDocumentImageGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageGroup, mongopagination.PaginationData, error) {
	args := m.Called(shopID, filters, pageable)
	return args.Get(0).([]models.DocumentImageGroup), args.Get(1).(mongopagination.PaginationData), args.Error(2)
}

func (m *MockDocumentImageRepository) GetDocumentImageGroup(shopID string, docRef string) (models.DocumentImageGroup, error) {
	args := m.Called(shopID, docRef)
	return args.Get(0).(models.DocumentImageGroup), args.Error(1)
}

func (m *MockDocumentImageRepository) UpdateDocumentImageStatus(shopID string, guid string, docnoGUIDRef string, status int8) error {
	args := m.Called(shopID, guid, docnoGUIDRef, status)

	return args.Error(0)
}

func (m *MockDocumentImageRepository) UpdateDocumentImageStatusByDocumentRef(shopID string, docRef string, docnoGUIDRef string, status int8) error {
	args := m.Called(shopID, docRef, docnoGUIDRef, status)

	return args.Error(0)
}

type MockDocumentImageFilePersister struct {
	mock.Mock
}

func (m *MockDocumentImageFilePersister) Save(fh *multipart.FileHeader, fileName string, fileExtension string) (string, error) {
	args := m.Called(fh, fileName, fileExtension)
	return args.String(0), args.Error(1)
}

func (m *MockDocumentImageFilePersister) LoadFile(fileName string) (string, *bytes.Buffer, error) {
	args := m.Called(fileName)
	return args.String(0), args.Get(0).(*bytes.Buffer), args.Error(2)
}

func CreateImage() *image.RGBA {
	width := 200
	height := 100

	upLeft := image.Point{0, 0}
	lowRight := image.Point{width, height}

	img := image.NewRGBA(image.Rectangle{upLeft, lowRight})

	// Colors are defined by Red, Green, Blue, Alpha uint8 values.
	cyan := color.RGBA{100, 200, 200, 0xff}

	// Set color for each pixel.
	for x := 0; x < width; x++ {
		for y := 0; y < height; y++ {
			switch {
			case x < width/2 && y < height/2: // upper left quadrant
				img.Set(x, y, cyan)
			case x >= width/2 && y >= height/2: // lower right quadrant
				img.Set(x, y, color.White)
			default:
				// Use zero value.
			}
		}
	}

	// Encode as PNG.
	f, _ := os.Create("image.png")
	png.Encode(f, img)

	return img
}

func TestDocumentImageUploadService(t *testing.T) {

	giveShopId := "TESTSHOP"
	giveUserId := "TESTUSER"
	giveModuleName := "TESTMODULE"

	giveImageUploadExtension := "png"
	giveImageUploadSuccessUri := "http:/xxxxxx"
	activityTime := time.Now()
	giveNewGuid := utils.NewGUID()
	wantCreateDocumentImage := models.DocumentImageDoc{
		DocumentImageData: models.DocumentImageData{
			ShopIdentity: common.ShopIdentity{
				ShopID: giveShopId,
			},
			DocumentImageInfo: models.DocumentImageInfo{
				DocIdentity: common.DocIdentity{
					GuidFixed: giveNewGuid,
				},
				DocumentImage: models.DocumentImage{
					UploadedBy:  giveUserId,
					UploadedAt:  activityTime,
					ImageUri:    giveImageUploadSuccessUri,
					Module:      giveModuleName,
					DocumentRef: giveNewGuid,
				},
			},
		},
		ActivityDoc: common.ActivityDoc{
			CreatedAt: activityTime,
			CreatedBy: giveUserId,
		},
	}

	giveImageUploadFileNameWithShop := fmt.Sprintf("%s/%s", giveShopId, giveNewGuid)

	// body := new(bytes.Buffer)
	// writer := multipart.NewWriter(body)
	// writer.WriteField("module", giveModuleName)
	// part, _ := writer.CreateFormFile("file", "image.png")
	// err := png.Encode(part, CreateImage())
	// assert.Nil(t, err, "Failed On Give File to Process Test")
	// writer.Close()
	giveFileHeader := &multipart.FileHeader{
		Filename: "image.png",
	}

	mockRepo := new(MockDocumentImageRepository)
	mockRepo.On("Create", wantCreateDocumentImage).Return(wantCreateDocumentImage.GuidFixed, nil)

	mockFilePersister := new(MockDocumentImageFilePersister)
	mockFilePersister.On("Save", giveFileHeader, giveImageUploadFileNameWithShop, giveImageUploadExtension).Return(giveImageUploadSuccessUri, nil)

	svc := services.DocumentImageService{
		Repo:          mockRepo,
		FilePersister: mockFilePersister,
		NowFn: func() time.Time {
			return activityTime
		},
		NewGUIDFn: func() string {
			return giveNewGuid
		},
	}
	get, err := svc.UploadDocumentImage(giveShopId, giveUserId, giveModuleName, giveFileHeader)
	assert.Nil(t, err, fmt.Sprintf("Failed After Service Upload Document Image"))
	assert.Equal(t, get, &wantCreateDocumentImage.DocumentImageInfo, "Failed After Service Upload Document Image Are Not Equal Given Test Data.")
}
*/
