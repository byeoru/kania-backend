package grpcclient

import (
	"flag"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	metadataPb "github.com/byeoru/kania/grpc_client/metadata"
	updatesPb "github.com/byeoru/kania/grpc_client/updates"
)

var (
	grpcInit       sync.Once
	clientInstance *Client
	addr           = flag.String("addr", "localhost:50051", "the address to connect to")
)

type Client struct {
	Conn                  *grpc.ClientConn
	MapDataClient         metadataPb.MapDataClient
	RealtimeUpdatesClient updatesPb.RealtimeUpdatesClient
}

func NewClient() *Client {
	grpcInit.Do(func() {
		flag.Parse()
		// Set up a connection to the server.
		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		mc := metadataPb.NewMapDataClient(conn)
		rc := updatesPb.NewRealtimeUpdatesClient(conn)

		clientInstance = &Client{
			Conn:                  conn,
			MapDataClient:         mc,
			RealtimeUpdatesClient: rc,
		}
	})
	return clientInstance
}
