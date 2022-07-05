module user_auth

go 1.18

replace (
	configs => ./configs
	handlers => ./handlers
	middleware => ./middleware
	model => ./model
	router => ./router
	services => ./services
)

require (
	configs v0.0.0-00010101000000-000000000000
	github.com/gofiber/fiber/v2 v2.34.1
	middleware v0.0.0-00010101000000-000000000000
	model v0.0.0-00010101000000-000000000000
	router v0.0.0-00010101000000-000000000000
	services v0.0.0-00010101000000-000000000000
)

require (
	github.com/andybalholm/brotli v1.0.4 // indirect
	github.com/klauspost/compress v1.15.0 // indirect
	github.com/lib/pq v1.10.6 // indirect
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.37.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sys v0.0.0-20220520151302-bc2c85ada10a // indirect
	gopkg.in/yaml.v3 v3.0.0 // indirect
	handlers v0.0.0-00010101000000-000000000000 // indirect
)
