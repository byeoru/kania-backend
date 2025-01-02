package errors

import "fmt"

type TxError struct {
	Code    int
	Message string
}

func (e *TxError) Error() string {
	return fmt.Sprintf("code: %d, message: %s", e.Code, e.Message)
}

func NewTxError(code int, message string) *TxError {
	return &TxError{
		Code:    code,
		Message: message,
	}
}
