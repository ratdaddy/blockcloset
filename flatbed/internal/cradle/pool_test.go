package cradle

import (
	"context"
	"testing"

	"google.golang.org/grpc"
)

func TestPool(t *testing.T) {
	t.Parallel()

	type tc struct {
		name           string
		getCalls       []string // addresses to call GetConn with
		wantSameConn   bool     // whether first and second call should return same conn
		wantNilConn    bool     // whether connection should be nil
	}

	cases := []tc{
		{
			name:        "first call creates new connection",
			getCalls:    []string{"localhost:9444"},
			wantNilConn: false,
		},
		{
			name:         "reuses existing connection",
			getCalls:     []string{"localhost:9444", "localhost:9444"},
			wantSameConn: true,
		},
		{
			name:         "different addresses get different connections",
			getCalls:     []string{"localhost:9444", "localhost:9445"},
			wantSameConn: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			pool := NewPool()
			defer pool.Close()

			ctx := context.Background()

			var conns []*grpc.ClientConn
			for _, addr := range c.getCalls {
				conn, err := pool.GetConn(ctx, addr)
				if err != nil {
					t.Fatalf("GetConn(%q): %v", addr, err)
				}
				conns = append(conns, conn)
			}

			if len(conns) > 0 {
				if c.wantNilConn {
					if conns[0] != nil {
						t.Fatal("GetConn returned non-nil connection, want nil")
					}
				} else {
					if conns[0] == nil {
						t.Fatal("GetConn returned nil connection")
					}
				}
			}

			if len(conns) >= 2 {
				if c.wantSameConn {
					if conns[0] != conns[1] {
						t.Fatal("GetConn returned different connection pointers, want same connection reused")
					}
				} else {
					if conns[0] == conns[1] {
						t.Fatal("GetConn returned same connection pointer, want different connections")
					}
				}
			}
		})
	}
}

func TestPool_Close(t *testing.T) {
	t.Parallel()

	type tc struct {
		name      string
		getCalls  []string // addresses to call GetConn with before Close
		wantError bool
	}

	cases := []tc{
		{
			name:      "Close cleans up all connections",
			getCalls:  []string{"localhost:9444", "localhost:9445", "localhost:9446"},
			wantError: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()

			pool := NewPool()
			ctx := context.Background()

			// Create connections
			var conns []*grpc.ClientConn
			for _, addr := range c.getCalls {
				conn, err := pool.GetConn(ctx, addr)
				if err != nil {
					t.Fatalf("GetConn(%q): %v", addr, err)
				}
				conns = append(conns, conn)
			}

			// Close the pool
			err := pool.Close()
			if c.wantError && err == nil {
				t.Fatal("Close returned nil error, want error")
			}
			if !c.wantError && err != nil {
				t.Fatalf("Close returned error: %v", err)
			}

			// Verify all connections are closed by checking their state
			for i, conn := range conns {
				state := conn.GetState()
				// A closed connection should be in Shutdown state
				if state.String() != "SHUTDOWN" {
					t.Errorf("connection %d (addr=%q) state: got %v, want SHUTDOWN", i, c.getCalls[i], state)
				}
			}
		})
	}
}
