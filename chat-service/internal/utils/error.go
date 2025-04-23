package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type ServiceError struct {
	StatusCode int    `json:"code"`
	Err        string `json:"error"`
	Message    string `json:"message"`
}

func (err *ServiceError) Error() string {
	return fmt.Sprintf("status %d: err %v", err.StatusCode, err.Err)
}

func (err *ServiceError) Bytes() []byte {
	bytes, _ := json.Marshal(err)
	return bytes
}

func NewError(err string, code int) *ServiceError {
	return &ServiceError{
		StatusCode: code,
		Message:    http.StatusText(code),
		Err:        err,
	}
}
