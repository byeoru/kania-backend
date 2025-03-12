package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	db "github.com/byeoru/kania/db/repository"
	grpcclient "github.com/byeoru/kania/grpc_client"
	"github.com/byeoru/kania/types"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	levyActionServiceInit     sync.Once
	levyActionServiceInstance *LevyActionService
)

type LevyActionService struct {
	store      db.Store
	grpcClient *grpcclient.Client
}

func newLevyActionService(store db.Store, grpcClient *grpcclient.Client) *LevyActionService {
	levyActionServiceInit.Do(func() {
		levyActionServiceInstance = &LevyActionService{
			store,
			grpcClient,
		}
	})
	return levyActionServiceInstance
}

func (s *LevyActionService) FindOne(ctx *gin.Context, levyActionId int64) (*db.LeviesAction, error) {
	return s.store.FindLevyAction(ctx, levyActionId)
}

func (s *LevyActionService) ExecuteLevyAction(ctx *gin.Context, arg *db.CreateLevyActionParams) (*db.LeviesAction, error) {
	var rs *db.LeviesAction
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		action, err := s.store.CreateLevyAction(ctx, arg)
		if err != nil {
			return err
		}

		levyStatusUpdateParams := db.UpdateLevyStatusParams{
			LevyID:    arg.LevyID,
			Stationed: false,
		}
		err = s.store.UpdateLevyStatus(ctx, &levyStatusUpdateParams)
		if err != nil {
			return err
		}
		rs = action
		return nil
	})

	return rs, err
}

func (s *LevyActionService) FindLevyAction(ctx *gin.Context, levyId int64) (*db.LeviesAction, error) {
	return s.store.FindLevyActionByLevyId(ctx, levyId)
}

func (s *LevyActionService) FindOnGoingMyRealmActions(ctx *gin.Context, realmId int64) ([]*db.LeviesAction, error) {
	return s.store.FindOnGoingMyRealmActions(ctx, realmId)
}

func (s *LevyActionService) FindLevyActionByLevyId(ctx *gin.Context, arg *db.FindLevyActionCountByLevyIdParams) (int64, error) {
	return s.store.FindLevyActionCountByLevyId(ctx, arg)
}

type SyncClientType = string
type ActionResultArgs struct {
	sector     int32
	oldRealmId int64
	newRealmId int64
	actionType SyncClientType
	actionId   int64
}

const (
	AnnexIndigenousSector SyncClientType = "AnnexIndigenousSector"
	ReturnToEncampment    SyncClientType = "ReturnToEncampment"
	ReturnCompleted       SyncClientType = "ReturnCompleted"
	AnnihilateAttacker    SyncClientType = "AnnihilateAttacker"
	SurrenderToTarget     SyncClientType = "SurrenderToTarget"
	ReinforceTroops       SyncClientType = "ReinforceTroops"
	NationFall            SyncClientType = "NationFall"
	AnnexAndDisbandUnit   SyncClientType = "AnnexAndDisbandUnit"
	AnnexSector           SyncClientType = "AnnexSector"
	Recapture             SyncClientType = "Recapture"
	CapitalCaptured       SyncClientType = "CapitalCaptured"
	AllCapitalsCaptured   SyncClientType = "AllCapitalsCaptured"
	None                  SyncClientType = "None" // error
)

func syncClientSectorInfo(s *LevyActionService, arg *ActionResultArgs) {
	s.grpcClient.UpdateSectorOwnership(&grpcclient.SectorOwnership{
		Sector:     arg.sector,
		OldRealmID: arg.oldRealmId,
		NewRealmID: arg.newRealmId,
		ActionType: arg.actionType,
		ActionID:   arg.actionId,
	})
}

func (s *LevyActionService) ExecuteCronLevyActions(ctx *context.Context, currentWorldTime time.Time) error {
	actions, err := s.store.FindLevyActionsBeforeDate(*ctx, currentWorldTime)
	if err != nil {
		return err
	}

	for _, action := range actions {
		var actionResultArgs ActionResultArgs
		err := s.store.ExecTx(*ctx, func(q *db.Queries) error {
			bIndigenous := false
			var targetRealmId sql.NullInt64

			switch action.LeviesAction.ActionType {
			case util.Attack:
				var battleOutcome db.CreateBattleOutcomeParams
				var originalAttackerLevy db.Levy

				// 전투 전 공격부대 정보 저장
				originalAttackerLevy = action.Levy
				defenderInfo, err := q.FindSectorRealmForUpdate(*ctx, action.LeviesAction.TargetSector)
				if err != nil {
					if err != sql.ErrNoRows {
						return err
					} else {
						// 토착 세력인 경우
						bIndigenous = true
					}
				}

				// 방어측이 토착 세력인 경우
				if bIndigenous {
					indigenousUnit, err := q.FindIndigenousUnit(*ctx, action.LeviesAction.TargetSector)
					if err != nil {
						return err
					}

					// 공격 부대 소속 국가가 멸망한 상태인 경우
					if !action.Levy.RealmID.Valid {
						arg := db.UpdateIndigenousUnitsParams{
							SectorNumber: action.LeviesAction.TargetSector,
							Swordmen:     action.Levy.Swordmen + action.Levy.ShieldBearers + action.Levy.SupplyTroop,
							Archers:      action.Levy.Archers,
							Lancers:      action.Levy.Lancers,
						}
						// 향군에 투항
						err := q.UpdateIndigenousUnits(*ctx, &arg)
						if err != nil {
							return err
						}
						// 투항한 공격 부대 해체
						err = q.RemoveLevy(*ctx, action.Levy.LevyID)
						if err != nil {
							return err
						}
						break
					}

					arg1 := db.Levy{
						Swordmen:  indigenousUnit.Swordmen,
						Archers:   indigenousUnit.Archers,
						Lancers:   indigenousUnit.Lancers,
						Stationed: true,
					}

					originalIndigenousUnit := arg1

					// 전투
					bAttackerAnnihilation, bDefenderAnnihilation := executeSingleBattleTurn(&action.Levy, &arg1)
					if bDefenderAnnihilation { // 향군이 전멸한 경우

						err := annexSectorOfIndigenousLand(ctx, s, q, action)
						if err != nil {
							return err
						}
						// 부대 주둔
						stationInSector(&action.Levy, action.LeviesAction.TargetSector)

						// sync arg
						actionResultArgs = ActionResultArgs{
							sector:     action.LeviesAction.TargetSector,
							newRealmId: action.Levy.RealmID.Int64,
							actionType: AnnexIndigenousSector,
							actionId:   action.LeviesAction.LevyActionID,
						}
					} else if !bAttackerAnnihilation { // 공격, 수비 부대 둘 다 생존한 경우
						// 공격부대 주둔지로 복귀
						speed := util.CalculateLevyAdvanceSpeed(&action.Levy)
						err := returnToEncampment(ctx, q, speed, action)
						if err != nil {
							return err
						}

						// sync arg
						actionResultArgs = ActionResultArgs{
							sector:     action.LeviesAction.TargetSector,
							actionType: ReturnToEncampment,
							actionId:   action.LeviesAction.LevyActionID,
						}

					} else {
						// 공격 부대가 전멸한 경우
						action.Levy.Stationed = true

						// sync arg
						actionResultArgs = ActionResultArgs{
							actionType: AnnihilateAttacker,
							actionId:   action.LeviesAction.LevyActionID,
						}
					}

					// 전투 결과 생성
					attackerDelta, err := makeBattleOutcomeJsonb(&originalAttackerLevy, &action.Levy)
					if err != nil {
						return err
					}

					defenderDelta, err := makeBattleOutcomeJsonb(&originalIndigenousUnit, &arg1)
					if err != nil {
						return err
					}

					battleOutcome = db.CreateBattleOutcomeParams{
						LevyActionID: actionResultArgs.actionId,
						RealmID:      action.Levy.RealmID.Int64,
						Attacker:     attackerDelta,
						Defender:     defenderDelta,
					}

					// 전투 후 공격부대 정보 업데이트
					err = updateLevies(ctx, q, []*db.Levy{&action.Levy})
					if err != nil {
						return err
					}

					arg2 := db.UpdateIndigenousUnitsParams{
						SectorNumber: action.LeviesAction.TargetSector,
						Swordmen:     arg1.Swordmen,
						Archers:      arg1.Archers,
						Lancers:      arg1.Lancers,
					}
					// 전투 후 향군 업데이트
					err = q.UpdateIndigenousUnits(*ctx, &arg2)
					if err != nil {
						return err
					}
					// 전투 결과 생성
					err = q.CreateBattleOutcome(*ctx, &battleOutcome)
					if err != nil {
						return err
					}
					break
				}

				/* 방어측이 국가인 경우(부흥세력 포함) */
				arg := db.FindEncampmentLeviesParams{
					RealmID:    sql.NullInt64{Int64: defenderInfo.Realm.RealmID, Valid: true},
					Encampment: action.LeviesAction.TargetSector,
				}
				targetRealmId = arg.RealmID
				defenders, err := q.FindEncampmentLevies(*ctx, &arg)
				if err != nil {
					return err
				}

				defenderSize := len(defenders)

				// 공격 부대 소속 국가가 멸망한 상태인 경우
				if !action.Levy.RealmID.Valid {
					arg1 := db.CreateLevySurrenderParams{
						LevyID:                    action.Levy.LevyID,
						ReceivingRealmID:          defenderInfo.Realm.RealmID,
						SurrenderReason:           util.RealmAnnihilation,
						SurrenderedAt:             action.LeviesAction.ExpectedCompletionAt,
						SurrenderedSectorLocation: action.LeviesAction.TargetSector,
					}
					// 투항
					err := q.CreateLevySurrender(*ctx, &arg1)
					if err != nil {
						return err
					}

					// 부대 주둔
					stationInSector(&action.Levy, action.LeviesAction.TargetSector)
					// 투항부대 업데이트
					err = updateLevies(ctx, q, []*db.Levy{&action.Levy})
					if err != nil {
						return err
					}

					// sync arg
					actionResultArgs = ActionResultArgs{
						actionType: SurrenderToTarget,
						actionId:   action.LeviesAction.LevyActionID,
					}
					break
				}

				// 현재 졈령한 부대와 국가가 같은 경우
				if defenderInfo.Sector.RealmID == action.Levy.RealmID.Int64 {
					// 부대 주둔
					stationInSector(&action.Levy, action.LeviesAction.TargetSector)
					// 주둔지에 부대 추가
					defenders = append(defenders, &action.Levy)
					// 부대 정보 업데이트
					err := updateLevies(ctx, q, defenders)
					if err != nil {
						return err
					}

					// sync arg
					actionResultArgs = ActionResultArgs{
						actionType: ReinforceTroops,
						actionId:   action.LeviesAction.LevyActionID,
					}
					break
				}

				// 방어 부대가 없는 경우
				if defenderSize == 0 {
					r, err := resolveUndefendedTerritory(ctx, q, action, defenderInfo, &defenders)
					if err != nil {
						return err
					}

					actionResultArgs = *r
					// 공격 부대 정보 업데이트
					err = updateLevies(ctx, q, []*db.Levy{&action.Levy})
					if err != nil {
						return err
					}
					break
				}

				var validDefenderLevyIndex int = -1
				for idx, defender := range defenders {
					// 전투가 가능한 부대일 경우
					if !util.IsAnnihilated(defender) {
						validDefenderLevyIndex = idx
						break
					}
				}

				// 적진에 전투가능한 방어 부대가 없는 경우
				if validDefenderLevyIndex == -1 {
					r, err := resolveUndefendedTerritory(ctx, q, action, defenderInfo, &defenders)
					if err != nil {
						return err
					}
					actionResultArgs = *r

					// 전투 결과 생성
					attackerDelta, err := makeBattleOutcomeJsonb(&originalAttackerLevy, &action.Levy)
					if err != nil {
						return err
					}

					var emptyLevy db.Levy
					defenderDelta, err := makeBattleOutcomeJsonb(&emptyLevy, &emptyLevy)
					if err != nil {
						return err
					}

					battleOutcome = db.CreateBattleOutcomeParams{
						LevyActionID: actionResultArgs.actionId,
						RealmID:      action.Levy.RealmID.Int64,
						Attacker:     attackerDelta,
						Defender:     defenderDelta,
					}
					break
				} else {
					originalDefenderLevy := *defenders[validDefenderLevyIndex]

					// 전투
					bAttackerAnnihilation, bDefenderAnnihilation := executeSingleBattleTurn(&action.Levy, defenders[validDefenderLevyIndex])
					// 전투에 참여한 방어 부대가 전멸한 경우
					if bDefenderAnnihilation {
						// 더이상 방어할 부대가 없는 경우
						if validDefenderLevyIndex >= defenderSize-1 {
							r, err := resolveUndefendedTerritory(ctx, q, action, defenderInfo, &defenders)
							if err != nil {
								return err
							}
							actionResultArgs = *r
						} else {
							// 공격부대 주둔지로 복귀
							speed := util.CalculateLevyAdvanceSpeed(&action.Levy)
							err := returnToEncampment(ctx, q, speed, action)
							if err != nil {
								return err
							}

							// sync arg
							actionResultArgs = ActionResultArgs{
								sector:     action.LeviesAction.TargetSector,
								actionType: ReturnToEncampment,
								actionId:   action.LeviesAction.LevyActionID,
							}
						}
					} else if !bAttackerAnnihilation { // 공격, 수비 둘 다 생존한 경우
						// 공격부대 주둔지로 복귀
						speed := util.CalculateLevyAdvanceSpeed(&action.Levy)
						err := returnToEncampment(ctx, q, speed, action)
						if err != nil {
							return err
						}

						// sync arg
						actionResultArgs = ActionResultArgs{
							sector:     action.LeviesAction.TargetSector,
							actionType: ReturnToEncampment,
							actionId:   action.LeviesAction.LevyActionID,
						}
					} else {
						// 공격 부대가 전멸한 경우
						action.Levy.Stationed = true

						// sync arg
						actionResultArgs = ActionResultArgs{
							actionType: AnnihilateAttacker,
							actionId:   action.LeviesAction.LevyActionID,
						}
					}
					// 전투 결과 생성
					attackerDelta, err := makeBattleOutcomeJsonb(&originalAttackerLevy, &action.Levy)
					if err != nil {
						return err
					}

					defenderDelta, err := makeBattleOutcomeJsonb(&originalDefenderLevy, defenders[validDefenderLevyIndex])
					if err != nil {
						return err
					}

					battleOutcome = db.CreateBattleOutcomeParams{
						LevyActionID: actionResultArgs.actionId,
						RealmID:      action.Levy.RealmID.Int64,
						Attacker:     attackerDelta,
						Defender:     defenderDelta,
					}
				}

				// 공격, 방어 부대 정보 업데이트
				err = updateLevies(ctx, q, append(defenders, &action.Levy))
				if err != nil {
					return err
				}
				// 전투 결과 생성
				err = q.CreateBattleOutcome(*ctx, &battleOutcome)
				if err != nil {
					return err
				}
			case util.Move:
			case util.Recapture:

			case util.Return:
				// 본국이 멸망한 경우
				if !action.Levy.RealmID.Valid {
					// 부대 해체
					err = q.RemoveLevy(*ctx, action.Levy.LevyID)
					if err != nil {
						return err
					}
					break
				}

				realmId, err := q.GetSectorRealmId(*ctx, action.LeviesAction.TargetSector)
				if err != nil {
					return err
				}

				// 부대 주둔지가 함락 당한 경우
				/**
				부대 주둔지가 함락되었다는 것은 보급병이 전멸하였다는 뜻이기 때문에 다른 곳으로 이동할 군량이 없다고 설정한다
				때문에 주둔지를 되찾기 위해 공격을 시도함
				*/
				if realmId != action.Levy.RealmID.Int64 {
					// 주둔지 수복 시도(월드 시간 2시간 후)
					arg2 := db.CreateLevyActionParams{
						LevyID:               action.Levy.LevyID,
						OriginSector:         action.LeviesAction.TargetSector,
						TargetSector:         action.LeviesAction.TargetSector,
						ActionType:           util.Recapture,
						Completed:            false,
						StartedAt:            action.LeviesAction.ExpectedCompletionAt,
						ExpectedCompletionAt: action.LeviesAction.ExpectedCompletionAt.Add(time.Hour * 2),
					}
					_, err = q.CreateLevyAction(*ctx, &arg2)
					if err != nil {
						return err
					}

					// sync arg
					actionResultArgs = ActionResultArgs{
						actionType: Recapture,
						actionId:   action.LeviesAction.LevyActionID,
					}
					break
				}

				// 주둔지 복귀에 성공한 경우
				arg := db.UpdateLevyStatusParams{
					LevyID:    action.Levy.LevyID,
					Stationed: true,
				}
				err = q.UpdateLevyStatus(*ctx, &arg)
				if err != nil {
					return err
				}
				// sync arg
				actionResultArgs = ActionResultArgs{
					actionType: ReturnCompleted,
					actionId:   action.LeviesAction.LevyActionID,
				}
			}
			arg := db.UpdateLevyActionCompletedParams{
				LevyActionID:  action.LeviesAction.LevyActionID,
				Completed:     true,
				TargetRealmID: targetRealmId,
			}
			// action 완료 업데이트
			err := q.UpdateLevyActionCompleted(*ctx, &arg)
			if err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			return err
		}
		syncClientSectorInfo(s, &actionResultArgs)
	}

	return nil
}

func makeBattleOutcomeJsonb(beforeBattle *db.Levy, afterBattle *db.Levy) ([]byte, error) {
	return json.Marshal(types.DeltaTroops{
		Swordman:     fmt.Sprintf("%d %d", beforeBattle.Swordmen, afterBattle.Swordmen),
		Archer:       fmt.Sprintf("%d %d", beforeBattle.Archers, afterBattle.Archers),
		Lancer:       fmt.Sprintf("%d %d", beforeBattle.Lancers, afterBattle.Lancers),
		ShieldBearer: fmt.Sprintf("%d %d", beforeBattle.ShieldBearers, afterBattle.ShieldBearers),
		SupplyTroop:  fmt.Sprintf("%d %d", beforeBattle.SupplyTroop, afterBattle.SupplyTroop),
	})
}

func resolveUndefendedTerritory(
	ctx *context.Context,
	q *db.Queries,
	action *db.FindLevyActionsBeforeDateRow,
	defenderInfo *db.FindSectorRealmForUpdateRow,
	defenders *[]*db.Levy,
) (*ActionResultArgs, error) {
	// 합병
	annexResult, err := annexSectorOfRealm(ctx, q, action, &defenderInfo.Realm)
	if err != nil {
		return nil, err
	}
	// 공격부대 주둔
	stationInSector(&action.Levy, action.LeviesAction.TargetSector)

	// 방어진영의 국가가 멸망했을 경우
	if annexResult == NationFall || defenderInfo.Realm.PoliticalEntity == util.RestorationForces {
		arg := db.RemoveStationedLeviesParams{
			RealmID:    sql.NullInt64{Int64: defenderInfo.Realm.RealmID, Valid: true},
			Encampment: action.LeviesAction.TargetSector,
		}
		// 방어부대 모두 해체
		err = q.RemoveStationedLevies(*ctx, &arg)
		*defenders = []*db.Levy{}

	} else {
		// 모든 방어부대(전멸상태) 주둔지 수도로 변경
		for _, defender := range *defenders {
			defender.Encampment = defenderInfo.Realm.Capitals[0]
		}
	}
	// sync arg
	return &ActionResultArgs{
		sector:     action.LeviesAction.TargetSector,
		oldRealmId: defenderInfo.Realm.RealmID,
		newRealmId: action.Levy.RealmID.Int64,
		actionType: annexResult,
		actionId:   action.LeviesAction.LevyActionID,
	}, err
}

func annexSectorOfRealm(
	ctx *context.Context,
	q *db.Queries,
	action *db.FindLevyActionsBeforeDateRow,
	defenderRealm *db.Realm,
) (annexResult SyncClientType, error error) {
	SectorsNumber, err := q.GetNumberOfRealmSectors(*ctx, defenderRealm.RealmID)
	if err != nil {
		return None, err
	}
	numberOfCapitals := len(defenderRealm.Capitals)
	bCapital := util.Find(defenderRealm.Capitals, action.LeviesAction.TargetSector)

	// 현재 영토가 방어측의 마지막 영토라면
	if SectorsNumber <= 1 {
		// 국가 멸망
		err := q.RemoveRealm(*ctx, defenderRealm.RealmID)
		if err != nil {
			return None, err
		}
		annexResult = NationFall
	} else if bCapital { // 수도가 함락되었을 경우
		// 수도가 모두 함락되었을 경우
		if numberOfCapitals == 1 && defenderRealm.Capitals[0] == action.LeviesAction.TargetSector {
			attackerOwnerRmId, err := q.GetRealmOwnerRmId(*ctx, action.LeviesAction.RealmID)
			if err != nil {
				return None, err
			}
			// 패배한 국가의 지도자 내탕금 몰수
			arg1 := db.TransferPrivateCoffersParams{
				ReceiverRmID:  attackerOwnerRmId,
				SourceRmID:    defenderRealm.OwnerRmID,
				ReductionRate: util.PrivateCoffersReductionRate,
			}
			privateTransferResult, err := q.TransferPrivateCoffers(*ctx, &arg1)
			if err != nil {
				return None, err
			}

			// 패배한 국가의 국고 몰수
			arg2 := db.TransferStateCoffersParams{
				ReceiverRealmID: action.LeviesAction.RealmID,
				SourceRealmID:   defenderRealm.RealmID,
				ReductionRate:   util.StateCoffersReductionRate,
			}
			stateTransferResult, err := q.TransferStateCoffers(*ctx, &arg2)
			if err != nil {
				return None, err
			}

			// 내탕금 로그 기록
			arg3 := db.CreateBothPrivateCoffersLogParams{
				SourceRmID:           defenderRealm.OwnerRmID,
				SourceChangeAmount:   -privateTransferResult.Delta,
				SourceTotalCoffers:   privateTransferResult.SourcePrivateCoffers,
				SourceReason:         AllCapitalsCaptured,
				ReceiverRmID:         attackerOwnerRmId,
				ReceiverChangeAmount: privateTransferResult.Delta,
				ReceiverTotalCoffers: privateTransferResult.ReceiverPrivateCoffers,
				ReceiverReason:       AllCapitalsCaptured,
				WorldTimeAt:          action.LeviesAction.ExpectedCompletionAt,
			}
			err = q.CreateBothPrivateCoffersLog(*ctx, &arg3)
			if err != nil {
				return None, err
			}

			// 국고 로그 기록
			arg4 := db.CreateBothStateCoffersLogParams{
				SourceRealmID:        defenderRealm.RealmID,
				SourceChangeAmount:   -stateTransferResult.Delta,
				SourceTotalCoffers:   stateTransferResult.SourceStateCoffers,
				SourceReason:         AllCapitalsCaptured,
				ReceiverRealmID:      action.LeviesAction.RealmID,
				ReceiverChangeAmount: stateTransferResult.Delta,
				ReceiverTotalCoffers: stateTransferResult.ReceiverStateCoffers,
				ReceiverReason:       AllCapitalsCaptured,
				WorldTimeAt:          action.LeviesAction.ExpectedCompletionAt,
			}
			err = q.CreateBothStateCoffersLog(*ctx, &arg4)
			if err != nil {
				return None, err
			}

			arg5 := db.UpdateRealmPoliticalEntityAndRemoveCapitalParams{
				RealmID:              defenderRealm.RealmID,
				PoliticalEntity:      util.RestorationForces,
				RemoveCapital:        action.LeviesAction.TargetSector,
				PopulationGrowthRate: util.RestorationForcesPopulationGrowthRate,
			}
			// 수비측 국가 체제를 부흥 세력으로 전환 && 수도 삭제
			err = q.UpdateRealmPoliticalEntityAndRemoveCapital(*ctx, &arg5)
			if err != nil {
				return None, err
			}
			annexResult = AllCapitalsCaptured
		} else {
			// 아직 다른 수도가 남아있는 경우
			arg := db.RemoveCapitalParams{
				RemoveCapital: action.LeviesAction.TargetSector,
				RealmID:       defenderRealm.RealmID,
			}
			err := q.RemoveCapital(*ctx, &arg)
			if err != nil {
				return None, err
			}
			annexResult = CapitalCaptured
		}
	} else {
		// 일반 sector가 함락되었을 경우
		annexResult = AnnexSector
	}

	// sector 소유권 이전
	err = changeSectorOwnership(
		*ctx,
		q,
		action.LeviesAction.TargetSector,
		defenderRealm.RealmID,
		action.Levy.RealmID.Int64,
		action.Levy.RmID,
	)
	if err != nil {
		return None, err
	}

	return
}

func annexSectorOfIndigenousLand(
	ctx *context.Context,
	s *LevyActionService,
	q *db.Queries,
	action *db.FindLevyActionsBeforeDateRow,
) error {
	// get sector metadata
	sectorMetadata, err := s.grpcClient.GetSectorInfo(action.LeviesAction.TargetSector)
	if err != nil {
		return err
	}

	arg1 := db.CreateSectorParams{
		CellNumber:     action.LeviesAction.TargetSector,
		ProvinceNumber: sectorMetadata.Province,
		RealmID:        action.Levy.RealmID.Int64,
		RmID:           action.Levy.RmID,
		Population:     sectorMetadata.Population,
	}

	// 점령
	err = q.CreateSector(*ctx, &arg1)
	if err != nil {
		return err
	}

	arg2 := db.AddRealmSectorJsonbParams{
		Key:     fmt.Sprintf("{%d}", action.LeviesAction.TargetSector),
		Value:   action.LeviesAction.TargetSector,
		RealmID: action.Levy.RealmID.Int64,
	}
	// json 영토 추가
	err = q.AddRealmSectorJsonb(*ctx, &arg2)
	if err != nil {
		return err
	}
	return nil
}

func changeSectorOwnership(
	ctx context.Context,
	q *db.Queries,
	sectorNumber int32,
	prevOwnerRealmId int64,
	newOwnerRealmId int64,
	newRmId int64,
) error {
	arg1 := db.UpdateSectorOwnershipParams{
		CellNumber: sectorNumber,
		RealmID:    newOwnerRealmId,
		RmID:       newRmId,
	}
	// sector 소유권 이전
	err := q.UpdateSectorOwnership(ctx, &arg1)
	if err != nil {
		return err
	}

	arg2 := db.AddRealmSectorJsonbParams{
		Key:     fmt.Sprintf("{%d}", sectorNumber),
		Value:   sectorNumber,
		RealmID: newOwnerRealmId,
	}
	// 승리한 국가에 소유권 추가
	err = q.AddRealmSectorJsonb(ctx, &arg2)
	if err != nil {
		return err
	}

	arg3 := db.RemoveSectorJsonbParams{
		Key:     fmt.Sprintf("%d", sectorNumber),
		RealmID: prevOwnerRealmId,
	}
	// 패배한 국가에 소유권 박탈
	err = q.RemoveSectorJsonb(ctx, &arg3)
	if err != nil {
		return err
	}
	return nil
}

func stationInSector(levy *db.Levy, sectorNumber int32) {
	// 공격부대 주둔지 변경
	levy.Encampment = sectorNumber
	// 주둔 상태 업데이트
	levy.Stationed = true
}

func updateLevies(ctx *context.Context, q *db.Queries, levies []*db.Levy) error {
	for _, levy := range levies {
		arg := db.UpdateLevyParams{
			LevyID:        levy.LevyID,
			Encampment:    levy.Encampment,
			Swordmen:      levy.Swordmen,
			ShieldBearers: levy.ShieldBearers,
			Archers:       levy.Archers,
			Lancers:       levy.Lancers,
			SupplyTroop:   levy.SupplyTroop,
			Stationed:     levy.Stationed,
			MovementSpeed: util.CalculateLevyAdvanceSpeed(levy),
		}
		err := q.UpdateLevy(*ctx, &arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func returnToEncampment(ctx *context.Context, q *db.Queries, speed float64, action *db.FindLevyActionsBeforeDateRow) error {
	totalDurationHours := action.LeviesAction.Distance / speed
	cd := util.CalculateCompletionDate(action.LeviesAction.ExpectedCompletionAt, totalDurationHours)

	arg := db.CreateLevyActionParams{
		LevyID:               action.LeviesAction.LevyID,
		RealmID:              action.Levy.RealmID.Int64,
		OriginSector:         action.LeviesAction.TargetSector,
		TargetSector:         action.LeviesAction.OriginSector,
		Distance:             action.LeviesAction.Distance,
		ActionType:           util.Return,
		Completed:            false,
		StartedAt:            action.LeviesAction.ExpectedCompletionAt,
		ExpectedCompletionAt: cd,
	}

	_, err := q.CreateLevyAction(*ctx, &arg)
	return err
}

func executeSingleBattleTurn(attacker *db.Levy, defender *db.Levy) (attackerAnnihilation bool, defenderAnnihilation bool) {
	defenderAnnihilation = attack(attacker, defender, false)
	if defenderAnnihilation {
		attackerAnnihilation = false
		return
	}
	attackerAnnihilation = attack(defender, attacker, true)
	return
}

func attack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (bDefenderAnnihilation bool) {
	// 궁병 공격: 상대 주둔중 -> (궁병, 창기병, 검병, 방패병, 보급병), 상대 출병 -> 보급병
	annihilation := archersFirstAttack(attacker, defender, bCounterattack)
	if annihilation {
		bDefenderAnnihilation = true
		return
	}
	// 창기병 공격: 상대 주둔중 -> (방패병, 창기병, 검병, 궁병, 보급병), 상대 출병 -> 보급병
	annihilation = lancersAttack(attacker, defender, bCounterattack)
	if annihilation {
		bDefenderAnnihilation = true
		return
	}
	// 검병 공격: 상대 주둔중 -> (창기병, 검병, 방패병, 궁병, 보급병), 상대 출병 -> 보급병
	annihilation = swordmenAttack(attacker, defender, bCounterattack)
	if annihilation {
		bDefenderAnnihilation = true
		return
	}
	// 궁병 공격: 상대 주둔중 -> (창기병, 궁병, 검병, 방패병, 보급병), 상대 출병 -> 보급병
	annihilation = archersAttack(attacker, defender, bCounterattack)
	if annihilation {
		bDefenderAnnihilation = true
		return
	}
	// 방패병 공격: 상대 주둔중 -> (검병, 창기병, 방패병, 궁병, 보급병), 상대 출병 -> 보급병
	annihilation = shieldBearerAttack(attacker, defender, bCounterattack)
	if annihilation {
		bDefenderAnnihilation = true
		return
	}
	return false
}

func archersFirstAttack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed && !bCounterattack:
		// 궁병 -> 보급병 공격
		if defender.SupplyTroop > 0 {
			defender.SupplyTroop = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		} else {
			annihilation = true
			return
		}
	// 궁병 -> 궁병 공격
	case defender.Archers > 0:
		if defender.ShieldBearers > 0 {
			defenderExtra := util.ExtraStat{
				AddedDefensive: 6,
			}
			defender.Archers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &defenderExtra)
		} else {
			defender.Archers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
		}
	// 궁병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 궁병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 궁병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 궁병 -> 보급병 공격
	case defender.SupplyTroop > 0:
		defender.SupplyTroop = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func lancersAttack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed && !bCounterattack:
		// 창기병 -> 보급병 공격
		if defender.SupplyTroop > 0 {
			defender.SupplyTroop = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		} else {
			annihilation = true
			return
		}
	// 창기병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 창기병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 창기병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 창기병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 창기병 -> 보급병 공격
	case defender.SupplyTroop > 0:
		defender.SupplyTroop = inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func swordmenAttack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed && !bCounterattack:
		// 검병 -> 보급병 공격
		if defender.SupplyTroop > 0 {
			defender.SupplyTroop = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		} else {
			annihilation = true
			return
		}
	// 검병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 검병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 검병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 검병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 검병 -> 보급병 공격
	case defender.SupplyTroop > 0:
		defender.SupplyTroop = inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func archersAttack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed && !bCounterattack:
		// 궁병 -> 보급병 공격
		if defender.SupplyTroop > 0 {
			defender.SupplyTroop = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		} else {
			annihilation = true
			return
		}
	// 궁병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
		// 궁병 -> 궁병 공격
	case defender.Archers > 0:
		if defender.ShieldBearers > 0 {
			defenderExtra := util.ExtraStat{
				AddedDefensive: 6,
			}
			defender.Archers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &defenderExtra)
		} else {
			defender.Archers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
		}
	// 궁병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 궁병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 궁병 -> 보급병 공격
	case defender.SupplyTroop > 0:
		defender.SupplyTroop = inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func shieldBearerAttack(attacker *db.Levy, defender *db.Levy, bCounterattack bool) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed && !bCounterattack:
		// 방패병 -> 보급병 공격
		if defender.SupplyTroop > 0 {
			defender.SupplyTroop = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		} else {
			annihilation = true
			return
		}
	// 방패병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 방패병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 방패병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 방패병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 방패병 -> 보급병 공격
	case defender.SupplyTroop > 0:
		defender.SupplyTroop = inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.SupplyTroop, unitStat.SupplyTroop, &emptyExtra)
		// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func inflictDamage(
	attackerCount int32,
	attackerStat *util.UnitStat,
	attackerExtraStat *util.ExtraStat,
	defenderCount int32,
	defenderStat *util.UnitStat,
	defenderExtraStat *util.ExtraStat,
) (defendUnitCount int32) {
	defenderTotalHP := defenderCount * defenderStat.HP
	damage := attackerCount * ((attackerStat.Offensive + attackerExtraStat.AddedOffensive) - (defenderStat.Defensive + defenderExtraStat.AddedDefensive))
	remainingTotalHP := defenderTotalHP - damage
	if remainingTotalHP <= 0 {
		return 0
	}
	return int32(remainingTotalHP / defenderStat.HP)
}
