package router

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ermos/annotation/parser"
	"github.com/julienschmidt/httprouter"
	"github.com/rs/cors"
	"github.com/vitego/config"
	"github.com/vitego/router/manager"
	"github.com/vitego/router/response"
	"log"
	"net/http"
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
		handler     interface{}
		err         error
	)

	router := httprouter.New()

	handler = router
	router.PanicHandler = eh.Render

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
				router.GET(routePath, call(route, ch, mwh))
			case "post":
				router.POST(routePath, call(route, ch, mwh))
			case "put":
				router.PUT(routePath, call(route, ch, mwh))
			case "patch":
				router.PATCH(routePath, call(route, ch, mwh))
			case "delete":
				router.DELETE(routePath, call(route, ch, mwh))
			}
		}
	}

	if config.Get("app.debug") == "true" {
		printHeader(config.Get("app.name"), config.Get("router.port"))
	}

	if config.Get("router.cors") != "" {
		c := cors.New(cors.Options{
			AllowedOrigins:   strings.Split(config.Get("router.cors.allowedOrigins"), ","),
			AllowedHeaders:   strings.Split(config.Get("router.cors.allowedHeaders"), ","),
			AllowedMethods:   strings.Split(config.Get("router.cors.allowedMethods"), ","),
			AllowCredentials: config.Get("router.cors.allowCredentials") == "true",
			// Enable Debugging for testing, consider disabling in production
			Debug: config.Get("app.debug") == "true",
		})

		handler = c.Handler(router)
	}

	return http.ListenAndServe(fmt.Sprintf(":%s", config.Get("router.port")), handler.(http.Handler))
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

func call(route parser.API, handler interface{}, mwh middlewareHandler) (handle httprouter.Handle) {
	build := reflect.ValueOf(handler).MethodByName(route.Controller)

	handle = getController(build)
	for _, mw := range mwh.Defaults() {
		handle = getMiddleware(mw, handle)
	}

	mws := mwh.Middleware()
	for _, mw := range route.Middleware {
		if mws[mw] == nil {
			log.Fatalf("%s middleware isn't specified", mw)
		}
		handle = getMiddleware(mws[mw], handle)
	}

	return initRequest(route, handle)
}

func getMiddleware(mw Middleware, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		mw.Run(r.Context().Value("manager").(*manager.Manager), next, w, r)
	}
}

func getController(build reflect.Value) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		c := make([]reflect.Value, 3)
		c = []reflect.Value{
			reflect.ValueOf(context.Background()),
			reflect.ValueOf(r.Context().Value("manager")),
			reflect.ValueOf(w),
		}
		_ = build.Call(c)
	}
}

func initRequest(route parser.API, next httprouter.Handle) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		m, status, err := manager.New(route, r, ps)
		if err != nil {
			response.Fatal(err, status)
		}

		ctx := context.WithValue(r.Context(), "manager", m)

		next(w, r.WithContext(ctx), ps)
	}
}
