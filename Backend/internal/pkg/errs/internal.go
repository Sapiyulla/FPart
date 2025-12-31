package errs

type InternalError struct {
	Domain string
}

func (e *InternalError) Error() string {
	return e.Domain
}
