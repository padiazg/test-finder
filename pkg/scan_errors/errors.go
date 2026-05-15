package scanerrors

import "errors"

type ScanError struct {
	Err error
}

func (e *ScanError) Error() string {
	return e.Err.Error()
}

func (e *ScanError) Unwrap() error {
	return e.Err
}

type ScanWarning struct {
	Err error
}

func (e *ScanWarning) Error() string {
	return e.Err.Error()
}

func (e *ScanWarning) Unwrap() error {
	return e.Err
}

type ScanTimeout struct {
	Err error
}

func (e *ScanTimeout) Error() string {
	return e.Err.Error()
}

func (e *ScanTimeout) Unwrap() error {
	return e.Err
}

func IsError(err error) bool {
	var (
		to *ScanTimeout
		se *ScanError
	)

	return errors.As(err, &to) || errors.As(err, &se)
}
