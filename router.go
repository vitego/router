package router

import (
	"encoding/json"
	"fmt"
	"github.com/ermos/annotation/parser"
	"github.com/gofiber/fiber/v2"
	"github.com/vitego/config"
	"github.com/vitego/router/manager"
	"github.com/vitego/router/response"
	"log"
	"os"
	"os/exec"
	"reflect"
	"strings"
)

// ListenAndServe listens on the TCP network address addr and then calls
// Serve with handler to handle requests on incoming connections.
// Accepted connections are configured to enable TCP keep-alives.
//
// The handler is typically nil, in which case the DefaultServeMux is used.
//
// ListenAndServe always returns a non-nil error.
func ListenAndServe(routes []byte, ch interface{}, eh exceptionHandler, mwh middlewareHandler) error {
	var (
		annotations []parser.API
		err         error
	)

	router := fiber.New()

	err = json.Unmarshal(routes, &annotations)
	if err != nil {
		log.Fatal(err)
	}

	for _, route := range annotations {
		for _, r := range route.Routes {
			routePath := config.Get("router.prefix")

			if route.Version != "" {
				routePath += fmt.Sprintf("/%s%s", route.Version, r.Route)
			} else {
				routePath += r.Route
			}

			switch strings.ToLower(r.Method) {
			case "get":
				router.Get(routePath, call(route, ch, mwh)...)
			case "post":
				router.Post(routePath, call(route, ch, mwh)...)
			case "put":
				router.Put(routePath, call(route, ch, mwh)...)
			case "patch":
				router.Patch(routePath, call(route, ch, mwh)...)
			case "delete":
				router.Delete(routePath, call(route, ch, mwh)...)
			}
		}
	}

	if config.Get("app.debug") == "true" {
		printHeader(config.Get("app.name"), config.Get("router.port"))
	}

	//if config.Get("router.cors") != "" {
	//	c := cors.New(cors.Options{
	//		AllowedOrigins:   strings.Split(config.Get("router.cors.allowedOrigins"), ","),
	//		AllowedHeaders:   strings.Split(config.Get("router.cors.allowedHeaders"), ","),
	//		AllowedMethods:   strings.Split(config.Get("router.cors.allowedMethods"), ","),
	//		AllowCredentials: config.Get("router.cors.allowCredentials") == "true",
	//		// Enable Debugging for testing, consider disabling in production
	//		Debug: config.Get("app.debug") == "true",
	//	})
	//
	//	handler = c.Handler(router)
	//}

	return router.Listen(fmt.Sprintf(":%s", config.Get("router.port")))
}

func printHeader(appName, port string) {
	var content string
	var separator string

	if appName != "" {
		content += fmt.Sprintf("%s's ", appName)
	}

	content += fmt.Sprintf("API currently running on port \033[1m%s\033[0m..", port)

	for i := 0; i < len(content)-6; i++ {
		separator += "-"
	}

	cmd := exec.Command("clear")
	cmd.Stdout = os.Stdout
	cmd.Run()

	fmt.Printf("%s--\n| ", separator)
	fmt.Print(content)
	fmt.Printf(" |\n--%s\n", separator)
}

func call(route parser.API, handler interface{}, mwh middlewareHandler) (handles []fiber.Handler) {
	build := reflect.ValueOf(handler).MethodByName(route.Controller)

	handles = append(handles, initRequest(route))

	for _, mw := range mwh.Defaults() {
		handles = append(handles, getMiddleware(mw))
	}

	mws := mwh.Middleware()
	rmws := route.Middleware

	for i, j := 0, len(rmws)-1; i < j; i, j = i+1, j-1 {
		rmws[i], rmws[j] = rmws[j], rmws[i]
	}

	for _, mw := range rmws {
		if mws[mw] == nil {
			log.Fatalf("%s middleware isn't specified", mw)
		}
		handles = append(handles, getMiddleware(mws[mw]))
	}

	handles = append(handles, getController(build))

	return
}

func getMiddleware(mw Middleware) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return mw.Run(c, c.Locals("manager").(*manager.Manager))
	}
}

func getController(build reflect.Value) fiber.Handler {
	return func(c *fiber.Ctx) error {
		ctr := make([]reflect.Value, 3)

		ctr = []reflect.Value{
			reflect.ValueOf(c),
			reflect.ValueOf(c.Context().Value("manager")),
		}

		_ = build.Call(ctr)

		return nil
	}
}

func initRequest(route parser.API) fiber.Handler {
	return func(c *fiber.Ctx) error {
		m, status, err := manager.New(route, c)
		if err != nil {
			response.Fatal(err, status)
		}

		c.Locals("manager", m)

		return c.Next()
	}
}
