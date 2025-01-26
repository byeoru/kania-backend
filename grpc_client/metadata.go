package grpcclient

import (
	"context"
	"time"

	pb "github.com/byeoru/kania/grpc_client/metadata"
)

func (c *Client) GetDistance(originSector int32, targetSector int32) (distance float64, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.MapDataClient.GetDistance(ctx, &pb.GetDistanceRequest{Origin: originSector, Target: targetSector})
	return r.Distance, err
}

type SectorMetadata struct {
	Province   int32
	Population int32
}

func (c *Client) GetSectorInfo(sector int32) (sectorMetadata *SectorMetadata, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.MapDataClient.GetSectorInfo(ctx, &pb.GetSectorInfoRequest{Sector: sector})
	return &SectorMetadata{Province: r.Province, Population: r.Population}, err
}
