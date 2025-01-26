package grpcclient

import (
	"flag"
	"log"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/byeoru/kania/grpc_client/metadata"
)

var (
	grpcInit       sync.Once
	clientInstance *Client
	addr           = flag.String("addr", "localhost:50051", "the address to connect to")
)

type Client struct {
	Conn          *grpc.ClientConn
	MapDataClient pb.MapDataClient
}

func NewClient() *Client {
	grpcInit.Do(func() {
		flag.Parse()
		// Set up a connection to the server.
		conn, err := grpc.NewClient(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		c := pb.NewMapDataClient(conn)

		clientInstance = &Client{
			Conn:          conn,
			MapDataClient: c,
		}
	})
	return clientInstance
}
