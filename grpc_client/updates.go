package grpcclient

import (
	"context"
	"time"

	pb "github.com/byeoru/kania/grpc_client/updates"
)

type SectorOwnership struct {
	Sector     int32
	OldRealmID int64
	NewRealmID int64
	ActionType string
	ActionID   int64
}

func (c *Client) UpdateSectorOwnership(arg *SectorOwnership) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	c.RealtimeUpdatesClient.UpdateSectorOwnership(ctx, &pb.UpdateSectorOwnershipRequest{
		Sector:     arg.Sector,
		OldRealmId: arg.OldRealmID,
		NewRealmId: arg.NewRealmID,
		ActionType: arg.ActionType,
		ActionId:   arg.ActionID,
	})
}
