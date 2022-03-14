package mw

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/vitego/router/manager"
	"log"
	"net/http"
)

type RequestMiddleware struct{}

func (RequestMiddleware) Run(c *fiber.Ctx, m *manager.Manager) error {
	if c.Method() != "OPTIONS" {
		log.Printf("%s \"%s %s\"\n",
			c.IP(),
			c.Method(),
			c.Request().URI().String(),
		)
	}

	_ = c.Next()

	fmt.Println("ok")
	return nil
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
