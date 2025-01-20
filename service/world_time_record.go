package service

import (
	"context"
	"sync"

	db "github.com/byeoru/kania/db/repository"
)

var (
	worldTimeRecordServiceInit sync.Once
	worldTimeRecordInstance    *WorldTimeRecordService
)

type WorldTimeRecordService struct {
	store db.Store
}

func newWorldTimeRecordService(store db.Store) *WorldTimeRecordService {
	worldTimeRecordServiceInit.Do(func() {
		worldTimeRecordInstance = &WorldTimeRecordService{
			store,
		}
	})
	return worldTimeRecordInstance
}

func (s *WorldTimeRecordService) CreateWorldTimeRecord(ctx *context.Context, arg *db.CreateWorldTimeRecordParams) error {
	return s.store.CreateWorldTimeRecord(*ctx, arg)
}
