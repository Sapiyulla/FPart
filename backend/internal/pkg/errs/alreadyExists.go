package errs

type ErrAlreadyExists struct {
	Resource string
}

func (e *ErrAlreadyExists) Error() string {
	return e.Resource + " already exists"
}
