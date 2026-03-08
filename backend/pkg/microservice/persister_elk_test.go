package microservice_test

type ConfigElkTest struct{}

func (c *ConfigElkTest) ElkAddress() []string {

	return []string{
		"http://192.168.2.204:9200",
	}
}

func (c *ConfigElkTest) Username() string {
	return "elastic"
}

func (c *ConfigElkTest) Password() string {
	return "smlSoft2021"
}

type TestElkModel struct {
	Title       string `json:"title"`
	Description string `json:"description"`
}

func (*TestElkModel) IndexName() string {
	return "test"
}

// func TestElkCreate(t *testing.T) {
// 	pst := microservice.NewPersisterElk(&ConfigElkTest{})
// 	err := pst.Create(&TestElkModel{
// 		Title:       "test title",
// 		Description: "test description",
// 	})

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// }

// func TestElkUpdate(t *testing.T) {
// 	pst := microservice.NewPersisterElk(&ConfigElkTest{})
// 	err := pst.Update("eN9AU38BxwH1fQnY_Kq8", &TestElkModel{
// 		Title:       "test title",
// 		Description: "test description",
// 	})

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// }

// func TestElkDelete(t *testing.T) {
// 	pst := microservice.NewPersisterElk(&ConfigElkTest{})
// 	err := pst.Delete(ctx,"9-DSU38BxwH1fQnY-VzD", &TestElkModel{})

// 	if err != nil {
// 		t.Error(err)
// 		return
// 	}
// }
