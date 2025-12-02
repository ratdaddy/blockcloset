module github.com/ratdaddy/blockcloset/cradle

go 1.25.0

require (
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3
	github.com/lmittmann/tint v1.1.2
	github.com/ratdaddy/blockcloset/loggrpc v0.0.0-20251130064613-168b3073df61
	github.com/ratdaddy/blockcloset/proto v0.0.0-00010101000000-000000000000
	google.golang.org/grpc v1.77.0
	google.golang.org/protobuf v1.36.10
)

require (
	golang.org/x/net v0.46.1-0.20251013234738-63d1a5100f82 // indirect
	golang.org/x/sys v0.37.0 // indirect
	golang.org/x/text v0.30.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20251022142026-3a174f9686a8 // indirect
)

replace github.com/ratdaddy/blockcloset/proto => ../proto

replace github.com/ratdaddy/blockcloset/loggrpc => ../loggrpc
