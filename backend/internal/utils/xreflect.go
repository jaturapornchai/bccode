package utils

import (
	"bytes"
	"fmt"
	"html/template"
	"reflect"

	"go.mongodb.org/mongo-driver/bson"
)

type ReflectTag struct {
	MainName string          `json:"mainname" bson:"mainname"`
	Child    ReflectTagChild `json:"child" bson:"child"`
	ReflectTagEmbed
}

type ReflectTagChild struct {
	ChildName string `json:"childname" bson:"childname"`
}

type ReflectTagEmbed struct {
	EmbedName string `json:"embedname" bson:"embedname"`
}

func GetReflectTagValue() {
	tagx := ReflectTag{}

	reflect.TypeOf(tagx).Field(0).Tag.Get("json")
	reflect.TypeOf(tagx).Field(0).Tag.Get("bson")

	fmt.Printf("Your type is %T\n", tagx)
	fmt.Println("reflect.TypeOf(tagx) = ", reflect.TypeOf(tagx))
	fmt.Println("reflect.TypeOf(tagx).Size() = ", reflect.TypeOf(tagx).Size())
	fmt.Println("reflect.TypeOf(tagx).Name() = ", reflect.TypeOf(tagx).Name())
	fmt.Println("reflect.TypeOf(tagx).NumField() = ", reflect.TypeOf(tagx).NumField())
	for i := 0; i < reflect.TypeOf(tagx).NumField(); i++ {
		fmt.Println(reflect.TypeOf(tagx).Field(i))
		fmt.Println("bson:: ", reflect.TypeOf(tagx).Field(i).Tag.Get("bson"))
	}
}

func ReplaceByTemplate() {

	data := map[string]interface{}{}

	data["Name"] = `"John'"`
	data["Street"] = `"Main Street"`

	rawText := `{
		"name": "{{.Name}}",
		"age":  30,
		"address": {
			"street": "{{.Street}}"
		}
	}`

	// Replace the placeholders with the variable values using a template
	tmpl, err := template.New("myTemplate").Parse(rawText)
	if err != nil {
		panic(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	replacedText := buf.String()

	var filter bson.M
	err = bson.UnmarshalExtJSON([]byte(replacedText), true, &filter)
	if err != nil {
		panic(err)
	}

	fmt.Printf("BSON data: %v\n", filter)
}
