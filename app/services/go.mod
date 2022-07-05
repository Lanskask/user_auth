module services

go 1.18

replace (
	configs => ../configs
	handlers => ../handlers
	model => ../model
)

require (
	configs v0.0.0-00010101000000-000000000000
	github.com/gofiber/fiber/v2 v2.34.1
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	handlers v0.0.0-00010101000000-000000000000
	model v0.0.0-00010101000000-000000000000
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.37.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/sys v0.0.0-20220227234510-4e6760a101f9 // indirect
)
