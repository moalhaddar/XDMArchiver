package twitter

import "fmt"

type ErrNot200 struct {
	OriginalError error
	StatusCode    int
}

func (err *ErrNot200) Error() string {
	return fmt.Sprintf("http status code %d", err.StatusCode)
}

func (e *ErrNot200) Unwrap() error {
	return e.OriginalError
}
