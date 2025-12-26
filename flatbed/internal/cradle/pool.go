package cradle

import (
	"context"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	"github.com/ratdaddy/blockcloset/flatbed/internal/requestid"
)

type Pool struct {
	mu     sync.Mutex
	conns  map[string]*grpc.ClientConn
	dialer func(context.Context, string) (*grpc.ClientConn, error)
}

func NewPool() *Pool {
	return &Pool{
		conns: make(map[string]*grpc.ClientConn),
		dialer: func(ctx context.Context, address string) (*grpc.ClientConn, error) {
			return grpc.NewClient(address,
				grpc.WithTransportCredentials(insecure.NewCredentials()),
				grpc.WithChainStreamInterceptor(requestIDStreamInterceptor()),
			)
		},
	}
}

func NewPoolWithDialer(dialer func(context.Context, string) (*grpc.ClientConn, error)) *Pool {
	return &Pool{
		conns:  make(map[string]*grpc.ClientConn),
		dialer: dialer,
	}
}

func (p *Pool) GetConn(ctx context.Context, address string) (*grpc.ClientConn, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if conn, ok := p.conns[address]; ok {
		return conn, nil
	}

	conn, err := p.dialer(ctx, address)
	if err != nil {
		return nil, err
	}

	p.conns[address] = conn
	return conn, nil
}

func (p *Pool) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, conn := range p.conns {
		if err := conn.Close(); err != nil {
			return err
		}
	}

	return nil
}

func requestIDStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if id := requestid.RequestIDFromContext(ctx); id != "" {
			ctx = metadata.AppendToOutgoingContext(ctx, "x-request-id", id)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
