module github.com/ratdaddy/blockcloset/gantry

go 1.24.6

require (
	github.com/ratdaddy/blockcloset/loggrpc v0.0.0
	github.com/ratdaddy/blockcloset/proto v0.0.0
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
)

require (
	github.com/lmittmann/tint v1.1.2 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
)

replace github.com/ratdaddy/blockcloset/loggrpc => ../loggrpc

replace github.com/ratdaddy/blockcloset/proto => ../proto
