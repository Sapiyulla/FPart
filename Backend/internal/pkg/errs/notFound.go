package errs

type NotFoundError struct {
	Domain string
}

func (e *NotFoundError) Error() string {
	return e.Domain + " not found"
}
