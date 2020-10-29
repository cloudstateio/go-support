package eventsourced

import (
	"context"
	"log"
	"net"
	"testing"

	"github.com/cloudstateio/go-support/cloudstate"
	"github.com/cloudstateio/go-support/cloudstate/eventsourced"
	"github.com/cloudstateio/go-support/cloudstate/protocol"
	shoppingcart2 "github.com/cloudstateio/go-support/example/shoppingcart"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type server struct {
	t              *testing.T
	server         *cloudstate.CloudState
	conn           *grpc.ClientConn
	lis            *bufconn.Listener
	teardownServer func()
	teardownClient func()
	serviceName    string
}

func newServer(t *testing.T) *server {
	t.Helper()
	s := server{t: t}
	if s.t == nil {
		panic("not test context defined")
	}
	s.t.Helper()
	server, err := cloudstate.New(protocol.Config{
		ServiceName:    "shopping-cart",
		ServiceVersion: "9.9.8",
	})
	if err != nil {
		s.t.Fatal(err)
	}
	s.server = server
	err = server.RegisterEventSourced(&eventsourced.Entity{
		ServiceName:   "com.example.shoppingcart.ShoppingCart",
		PersistenceID: "ShoppingCart",
		EntityFunc:    shoppingcart2.NewShoppingCart,
		SnapshotEvery: 1,
	}, protocol.DescriptorConfig{
		Service: "shoppingcart.proto",
	}.AddDomainDescriptor("domain.proto"))
	if err != nil {
		s.t.Fatal(err)
	}
	s.lis = bufconn.Listen(1024 * 1024)
	s.teardownServer = func() {
		s.server.Stop()
	}
	go func() {
		if err := server.RunWithListener(s.lis); err != nil {
			log.Fatalf("Server exited with error: %v", err)
		}
	}()
	return &s
}

func (s *server) newClientConn() {
	if s.conn != nil && s.teardownClient != nil {
		s.teardownClient()
	}
	// client
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return s.lis.Dial()
	}), grpc.WithInsecure())
	s.conn = conn
	if err != nil {
		s.t.Fatalf("Failed to dial bufnet: %v", err)
	}
	s.teardownClient = func() {
		s.conn.Close()
	}
}

func (s *server) teardown() {
	s.teardownClient()
	s.teardownServer()
}
