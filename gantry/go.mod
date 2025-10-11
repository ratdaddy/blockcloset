module github.com/ratdaddy/blockcloset/gantry

go 1.24.6

require (
	github.com/google/uuid v1.6.0
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.2
	github.com/lmittmann/tint v1.1.2
	github.com/ratdaddy/blockcloset/loggrpc v0.0.0
	github.com/ratdaddy/blockcloset/pkg v0.0.0
	github.com/ratdaddy/blockcloset/proto v0.0.0
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.8
	modernc.org/sqlite v1.39.0
)

require (
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/ncruces/go-strftime v0.1.9 // indirect
	github.com/remyoudompheng/bigfft v0.0.0-20230129092748-24d4a6f8daec // indirect
	golang.org/x/exp v0.0.0-20250620022241-b7579e27df2b // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.34.0 // indirect
	golang.org/x/text v0.26.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250707201910-8d1bb00bc6a7 // indirect
	modernc.org/libc v1.66.3 // indirect
	modernc.org/mathutil v1.7.1 // indirect
	modernc.org/memory v1.11.0 // indirect
)

replace github.com/ratdaddy/blockcloset/loggrpc => ../loggrpc

replace github.com/ratdaddy/blockcloset/proto => ../proto

replace github.com/ratdaddy/blockcloset/pkg => ../pkg
