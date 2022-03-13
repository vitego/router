package router

import (
	"github.com/julienschmidt/httprouter"
	"github.com/vitego/router/manager"
	"net/http"
)

type middlewareHandler interface {
	Defaults() []Middleware
	Middleware() map[string]Middleware
}

type Middleware interface {
	Run(m *manager.Manager, next httprouter.Handle, w http.ResponseWriter, r *http.Request)
}

//func Handle(m *manager.Manager, next httprouter.Handle) httprouter.Handle {
//
//}
