package validator

import (
	"github.com/go-playground/validator/v10"

	"listenly-backend/pkg/apperror"
)

type CustomValidator struct {
	validator *validator.Validate
}

func New() *CustomValidator {
	return &CustomValidator{validator: validator.New()}
}

func (cv *CustomValidator) Validate(i interface{}) error {
	if err := cv.validator.Struct(i); err != nil {
		return TranslateError(err)
	}
	return nil
}

func TranslateError(err error) error {
	validationErrors, ok := err.(validator.ValidationErrors)
	if !ok {
		return err
	}

	errs := make(map[string]string)
	for _, fe := range validationErrors {
		errs[fe.Field()] = translateTag(fe)
	}

	return &apperror.ValidationError{Errors: errs}
}

func translateTag(fe validator.FieldError) string {
	field := fe.Field()

	switch fe.Tag() {
	case "required":
		return field + " is required"
	case "email":
		return field + " must be a valid email"
	case "min":
		return field + " must be at least " + fe.Param() + " characters"
	case "max":
		return field + " must be at most " + fe.Param() + " characters"
	case "oneof":
		return field + " must be one of: " + fe.Param()
	default:
		return field + " is invalid"
	}
}
