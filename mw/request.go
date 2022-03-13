package mw

import (
	"github.com/julienschmidt/httprouter"
	"github.com/vitego/router/manager"
	"log"
	"net/http"
)

type RequestMiddleware struct{}

func (RequestMiddleware) Run(m *manager.Manager, next httprouter.Handle, w http.ResponseWriter, r *http.Request) {
	if r.Method != "OPTIONS" {
		m.Set("ipAddr", getIPAddr(r))
		log.Printf("%s \"%s %s\"\n",
			m.Get("ipAddr").(string),
			r.Method,
			r.RequestURI,
		)
	}
	m.Next(next, w, r)
}

func getIPAddr(r *http.Request) string {
	IPAddress := r.Header.Get("X-Real-Ip")
	if IPAddress == "" {
		IPAddress = r.Header.Get("X-Real-Ip")
	}
	if IPAddress == "" {
		IPAddress = r.RemoteAddr
	}
	return IPAddress
}
