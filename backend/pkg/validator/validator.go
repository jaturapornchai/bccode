package validator

import (
	"errors"
	"reflect"
	"strings"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translations "github.com/go-playground/validator/v10/translations/en"
)

type CustomValidator struct {
	Validator *validator.Validate
}

func NewCustomValidator() *CustomValidator {
	validate := validator.New()

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]

		if name == "-" {
			return ""
		}

		return name
	})

	customValidator := &CustomValidator{Validator: validate}
	return customValidator
}

func (cv *CustomValidator) Validate(i interface{}) error {
	en := en.New()
	uni := ut.New(en, en)

	trans, _ := uni.GetTranslator("en")
	en_translations.RegisterDefaultTranslations(cv.Validator, trans)

	if err := cv.Validator.Struct(i); err != nil {

		// return err
		// errs := err.(validator.ValidationErrors)

		// // returns a map with key = namespace & value = translated error
		// // NOTICE: 2 errors are returned and you'll see something surprising
		// // translations are i18n aware!!!!
		// // eg. '10 characters' vs '1 character'
		// fmt.Println(errs.Translate(trans))

		object, _ := err.(validator.ValidationErrors)

		for _, key := range object {
			return errors.New(key.Translate(trans))
		}
	}
	return nil
}
