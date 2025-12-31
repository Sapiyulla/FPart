package errs

type DuplicateError struct{}

func (e *DuplicateError) Error() string {
	return "email duplication error"
}
