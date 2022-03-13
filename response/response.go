package response

import (
	"encoding/json"
	"fmt"
	"github.com/vitego/config"
	"github.com/vitego/router/httperror"
	"net/http"
	"reflect"
)

type errorContent struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message"`
}

// Fatal stop and send error to error handler
func Fatal(err interface{}, statusAndCode ...int) {
	status := http.StatusInternalServerError
	code := 0

	if len(statusAndCode) > 0 {
		status = statusAndCode[0]
		if len(statusAndCode) > 1 {
			code = statusAndCode[1]
		}
	}

	panic(httperror.Error{
		Status: status,
		Code:   code,
		Value:  err,
	})
}

// Error send a error response
func Error(w http.ResponseWriter, status int, code int, err interface{}) bool {
	var (
		errorText  string
		errorValue string
	)

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if reflect.ValueOf(err).Type().String() == "string" {
		errorText = err.(string)
	} else {
		errorText = err.(error).Error()
	}

	errorValue = errorText
	if status == http.StatusInternalServerError {
		if config.Get("app.debug") != "true" {
			errorValue = config.Get("router.internalErrorMask")
		}
		fmt.Printf("[ Error ] %s\n", errorText)
	}

	failed := json.NewEncoder(w).Encode(&errorContent{Code: code, Message: errorValue})
	return failed != nil
}

// Success send success response
func Success(w http.ResponseWriter, status int, success interface{}) bool {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(status)

	if isNil(success) {
		_, _ = fmt.Fprintf(w, "[]")
		return true
	}

	failed := json.NewEncoder(w).Encode(success)
	return failed != nil
}

// NoContent send success response without body
func NoContent(w http.ResponseWriter) bool {
	return Success(w, http.StatusNoContent, nil)
}

func isNil(i interface{}) bool {
	if reflect.ValueOf(i).Kind() == reflect.Slice {
		return reflect.ValueOf(i).IsNil()
	}
	return false
}
