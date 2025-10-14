module github.com/ratdaddy/blockcloset/flatbed

go 1.24.6

require (
	github.com/go-chi/chi/v5 v5.2.2
	github.com/go-chi/httplog/v3 v3.2.2
	github.com/google/go-cmp v0.7.0
	github.com/lmittmann/tint v1.1.2
	github.com/oklog/ulid/v2 v2.1.1
	github.com/ratdaddy/blockcloset/pkg v0.0.0
	github.com/ratdaddy/blockcloset/proto v0.0.0
	google.golang.org/grpc v1.75.1
	google.golang.org/protobuf v1.36.9
)

require (
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)

replace github.com/ratdaddy/blockcloset/proto => ../proto

replace github.com/ratdaddy/blockcloset/pkg => ../pkg
