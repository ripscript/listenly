package apperror

type ValidationError struct {
	Errors map[string]string
}

func (v *ValidationError) Error() string {
	return "validation failed"
}
