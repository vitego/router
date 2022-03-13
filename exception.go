package router

import "net/http"

type exceptionHandler interface {
	Render(w http.ResponseWriter, r *http.Request, err interface{})
}
