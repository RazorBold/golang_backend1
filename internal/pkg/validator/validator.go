package validator

import "github.com/go-playground/validator/v10"

var v = validator.New()

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func Validate(s any) []FieldError {
	err := v.Struct(s)
	if err == nil {
		return nil
	}

	var errs []FieldError
	for _, e := range err.(validator.ValidationErrors) {
		errs = append(errs, FieldError{
			Field:   e.Field(),
			Message: fieldMessage(e),
		})
	}
	return errs
}

func fieldMessage(e validator.FieldError) string {
	switch e.Tag() {
	case "required":
		return "field is required"
	case "email":
		return "invalid email format"
	case "min":
		return "value too short (min " + e.Param() + ")"
	case "max":
		return "value too long (max " + e.Param() + ")"
	case "uuid4":
		return "invalid UUID format"
	default:
		return "invalid value"
	}
}
