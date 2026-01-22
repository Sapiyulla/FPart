package errs

type InternalError struct{}

func (e *InternalError) Error() string {
	return "internal server error"
}
