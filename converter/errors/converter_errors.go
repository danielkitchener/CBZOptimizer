package errors

type PageIgnoredError struct {
	s string
}

func (e *PageIgnoredError) Error() string {
	return e.s
}

func NewPageIgnored(text string) error {
	return &PageIgnoredError{text}
}
