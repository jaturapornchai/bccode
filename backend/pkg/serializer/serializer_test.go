package serializer_test

// import (
// 	"smlcloudplatform/pkg/serializer"
// 	"testing"
// 	"time"

// 	"github.com/stretchr/testify/assert"
// )

// type TestStructStd struct {
// 	Str     string  `json:"str"`
// 	Int     int     `json:"int"`
// 	Float64 float64 `json:"float64"`
// 	Bool    bool    `json:"bool"`
// }

// type TestStructObj struct {
// 	Date time.Time `json:"date"`
// }

// type TestStructParent struct {
// 	ID       int               `json:"id"`
// 	Name     string            `json:"name"`
// 	Child    TestStructChild   `json:"child"`
// 	Children []TestStructChild `json:"children"`
// }

// type TestStructChild struct {
// 	ID   int    `json:"id"`
// 	Name string `json:"name"`
// }

// func TestDecodeStdFromText(t *testing.T) {
// 	rawText := `{"str":"test","int":1,"float64":1.1,"bool":true}`
// 	var testStruct TestStructStd
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, "test", testStruct.Str)
// 	assert.Equal(t, 1, testStruct.Int)
// 	assert.Equal(t, 1.1, testStruct.Float64)
// 	assert.Equal(t, true, testStruct.Bool)
// }

// func TestDecodeStdFromTextDefault(t *testing.T) {
// 	rawText := `{}`
// 	var testStruct TestStructStd
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, "", testStruct.Str)
// 	assert.Equal(t, 0, testStruct.Int)
// 	assert.Equal(t, 0.0, testStruct.Float64)
// 	assert.Equal(t, false, testStruct.Bool)
// }

// func TestDecodeObjFromText(t *testing.T) {
// 	rawText := `{"date":"2020-01-01T00:00:00Z"}`
// 	var testStruct TestStructObj
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC), testStruct.Date)
// }

// func TestDecodeObjFromTextEmpty(t *testing.T) {
// 	rawText := `{}`
// 	var testStruct TestStructObj
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// }

// func TestDecodeNestestObj(t *testing.T) {
// 	rawText := `{"id":1,"name":"test","child":{"id":1,"name":"test"},"children":[{"id":1,"name":"test"}]}`
// 	var testStruct TestStructParent
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, 1, testStruct.ID)
// 	assert.Equal(t, "test", testStruct.Name)
// 	assert.Equal(t, 1, testStruct.Child.ID)
// 	assert.Equal(t, "test", testStruct.Child.Name)
// 	assert.Equal(t, 1, testStruct.Children[0].ID)
// 	assert.Equal(t, "test", testStruct.Children[0].Name)
// }

// func TestDecodeNestestObjEmpty(t *testing.T) {
// 	rawText := `{}`
// 	var testStruct TestStructParent
// 	err := serializer.Unmarshal([]byte(rawText), &testStruct)

// 	assert.Nil(t, err)
// }

// func TestEncodeStdToText(t *testing.T) {
// 	testStruct := TestStructStd{
// 		Str:     "test",
// 		Int:     1,
// 		Float64: 1.1,
// 		Bool:    true,
// 	}
// 	rawText, err := serializer.Marshal(testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"str":"test","int":1,"float64":1.1,"bool":true}`, string(rawText))
// }

// func TestEncodeObjToText(t *testing.T) {
// 	testStruct := TestStructObj{
// 		Date: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
// 	}
// 	rawText, err := serializer.Marshal(testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"date":"2020-01-01T00:00:00Z"}`, string(rawText))
// }

// func TestEncodeNestestObj(t *testing.T) {
// 	testStruct := TestStructParent{
// 		ID:   1,
// 		Name: "test",
// 		Child: TestStructChild{
// 			ID:   1,
// 			Name: "test",
// 		},
// 		Children: []TestStructChild{
// 			{
// 				ID:   1,
// 				Name: "test",
// 			},
// 		},
// 	}
// 	rawText, err := serializer.Marshal(testStruct)

// 	assert.Nil(t, err)
// 	assert.Equal(t, `{"id":1,"name":"test","child":{"id":1,"name":"test"},"children":[{"id":1,"name":"test"}]}`, string(rawText))
// }
