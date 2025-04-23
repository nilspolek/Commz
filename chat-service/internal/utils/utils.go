package utils

import (
	"encoding/json"
	"net/http"
)

func SendJsonResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

func SendJsonError(w http.ResponseWriter, err error) {
	// dont send empty error
	if err == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	customError, ok := err.(*ServiceError)
	if ok {
		// marshal error into json and send status from error as statuscode
		w.WriteHeader(customError.StatusCode)
		w.Write(customError.Bytes())
		return
	}

	customError = NewError("Internal server error", http.StatusInternalServerError)
	w.WriteHeader(customError.StatusCode)
	w.Write(customError.Bytes())
}
