# loggrpc

> Structured, context-aware logging for gRPC servers and clients in Go, built on the standard library `log/slog` package.

[![Go Reference](https://pkg.go.dev/badge/github.com/ratdaddy/loggrpc.svg)](https://pkg.go.dev/github.com/ratdaddy/loggrpc)
[![Go Report Card](https://goreportcard.com/badge/github.com/ratdaddy/loggrpc)](https://goreportcard.com/report/github.com/ratdaddy/loggrpc)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

`loggrpc` provides lightweight logging for Go gRPC servers and clients — inspired by [httplog](https://github.com/go-chi/httplog).

## Install

```bash
go get github.com/ratdaddy/loggrpc
```

## Overview

`loggrpc` makes it easy to add consistent, structured logs to gRPC services and clients.
It provides ready-to-use interceptors that capture request/response metadata and log them using Go’s [`slog`](https://pkg.go.dev/log/slog).

* **OTEL-friendly fields** — method, service, status code, latency, peer, trace/span IDs.
* **Drop-in interceptors** — works with both server and client.
* **Configurable** — choose log level, redact fields, sample requests, cap payload size.
* **slog-first** — built on the standard logger, no external logging dependency.

## Usage

Add the interceptor to your gRPC server:

```go
import (
    "google.golang.org/grpc"
    "github.com/ratdaddy/loggrpc"
)

logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

server := grpc.NewServer(
    grpc.UnaryInterceptor(loggrpc.UnaryServerInterceptor(logger)),
)

pb.RegisterMyServiceServer(server, &myService{})
```

On the client side:

```go
conn, _ := grpc.Dial(
    target,
    grpc.WithUnaryInterceptor(loggrpc.UnaryClientInterceptor(logger)),
)
defer conn.Close()
```

Example log (JSON):

```json
{
  "time": "2025-09-01T12:34:56Z",
  "level": "INFO",
  "grpc.service": "helloworld.Greeter",
  "grpc.method": "SayHello",
  "grpc.code": "OK",
  "duration_ms": 3.42,
  "peer.address": "127.0.0.1:54321",
  "msg": "gRPC request"
}
```

## Roadmap

* [x] Unary server interceptor
* [ ] Unary client interceptor
* [ ] Streaming interceptors
* [ ] More configurable fields & redaction

## License

TBD

