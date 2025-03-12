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
		// Self-signed 인증서를 로드
		// certPool := x509.NewCertPool()
		// certPath := "certificate/cert.pem" // 인증서 경로
		// cert, err := os.ReadFile(certPath)
		// if err != nil {
		// 	log.Fatalf("인증서 읽기 오류: %v", err)
		// }

		// 인증서 인증 풀에 추가
		// ok := certPool.AppendCertsFromPEM(cert)
		// if !ok {
		// 	log.Fatalf("인증서 풀에 추가 실패")
		// }

		// TLS 인증 설정
		// creds := credentials.NewTLS(&tls.Config{
		// 	RootCAs: certPool, // 인증서 풀 사용
		// })

		// // 인증서 로드 (서버 인증서 사용)
		// creds, err := credentials.NewClientTLSFromFile("certificate/cert.pem", "")
		// if err != nil {
		// 	log.Fatalf("TLS 인증서 로드 실패: %v", err)
		// }

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
