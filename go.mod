module github.com/vitego/router

go 1.17

require (
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/ermos/annotation v0.0.0-20210309132609-a4e71ea8028f
	github.com/julienschmidt/httprouter v1.3.0
	github.com/rs/cors v1.8.2
	github.com/vitego/config v1.0.0
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/gofiber/fiber/v2 v2.29.0 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.34.0 // indirect
	github.com/valyala/fastjson v1.6.3 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
)

replace github.com/ermos/annotation v0.0.0-20210309132609-a4e71ea8028f => ../../ermos/annotation
