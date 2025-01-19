package service

import (
	"context"
	"database/sql"
	"sync"
	"time"

	db "github.com/byeoru/kania/db/repository"
	"github.com/byeoru/kania/util"
	"github.com/gin-gonic/gin"
)

var (
	levyActionServiceInit     sync.Once
	levyActionServiceInstance *LevyActionService
)

type LevyActionService struct {
	store db.Store
}

func newLevyActionService(store db.Store) *LevyActionService {
	levyActionServiceInit.Do(func() {
		levyActionServiceInstance = &LevyActionService{
			store,
		}
	})
	return levyActionServiceInstance
}

func (s *LevyActionService) ExecuteLevyAction(ctx *gin.Context, arg *db.CreateLevyActionParams) error {
	return s.store.ExecTx(ctx, func(q *db.Queries) error {
		err := s.store.CreateLevyAction(ctx, arg)
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

		return nil
	})
}

func (s *LevyActionService) FindLevyAction(ctx *gin.Context, arg *db.FindLevyActionParams) (*db.LeviesAction, error) {
	return s.store.FindLevyAction(ctx, arg)
}

func (s *LevyActionService) FindLevyActionByLevyId(ctx *gin.Context, arg *db.FindLevyActionCountByLevyIdParams) (int64, error) {
	return s.store.FindLevyActionCountByLevyId(ctx, arg)
}

func (s *LevyActionService) ExecuteLevyActions(ctx *context.Context, currentWorldTime time.Time) error {
	actions, err := s.store.FindLevyActionsBeforeDate(*ctx, currentWorldTime)
	if err != nil {
		return err
	}

	for _, action := range actions {
		s.store.ExecTx(*ctx, func(q *db.Queries) error {
			bIndigenous := false

			switch action.LeviesAction.ActionType {
			case util.Attack:
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
						Swordmen: indigenousUnit.Swordmen,
						Archers:  indigenousUnit.Archers,
						Lancers:  indigenousUnit.Lancers,
					}
					bAttackerAnnihilation, bDefenderAnnihilation := executeSingleBattleTurn(&action.Levy, &arg1)

					if bDefenderAnnihilation { // 향군이 전멸한 경우
						err := annexSectorOfIndigenousLand(ctx, q, action)
						if err != nil {
							return err
						}
						// 부대 주둔
						stationInSector(&action.Levy, action.LeviesAction.TargetSector)
					} else if !bAttackerAnnihilation { // 공격, 수비 부대 둘 다 생존한 경우
						// 공격부대 주둔지로 복귀
						err = returnToEncampment(ctx, q, action)
						if err != nil {
							return err
						}
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

					// 공격 부대가 전멸한 경우 다른 처리를 하지 않는다.
					break
				}

				/* 방어측이 국가인 경우(부흥세력 포함) */
				arg := db.FindEncampmentLeviesParams{
					RealmID:    sql.NullInt64{Int64: defenderInfo.Realm.RealmID, Valid: true},
					Encampment: action.LeviesAction.TargetSector,
				}
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
					break
				}

				var validDefenderLevyIndex int
				for idx, defender := range defenders {
					// 방어 부대에 벙사가 없을 경우
					if util.IsAnnihilated(defender) {

					} else {
						// 전투가 가능한 부대일 경우
						validDefenderLevyIndex = idx
						break
					}
				}

				// 전투
				bAttackerAnnihilation, bDefenderAnnihilation := executeSingleBattleTurn(&action.Levy, defenders[validDefenderLevyIndex])
				// 전투에 참여한 방어 부대가 전멸한 경우
				if bDefenderAnnihilation {
					// 더이상 방어할 부대가 없는 경우
					if validDefenderLevyIndex >= defenderSize-1 {
						// 합병
						annexResult, err := annexSectorOfRealm(ctx, q, action, &defenderInfo.Realm)
						if err != nil {
							return err
						}
						// 공격부대 주둔
						stationInSector(&action.Levy, action.LeviesAction.TargetSector)

						// 방어진영의 국가가 멸망했거나 부흥세력일 경우
						if annexResult == util.Annihilation || defenderInfo.Realm.PoliticalEntity == util.RestorationForces {
							arg := db.RemoveStationedLeviesParams{
								RealmID:    sql.NullInt64{Int64: defenderInfo.Realm.RealmID, Valid: true},
								Encampment: action.LeviesAction.TargetSector,
							}
							// 방어부대 모두 해체
							err := q.RemoveStationedLevies(*ctx, &arg)
							if err != nil {
								return err
							}
							defenders = []*db.Levy{}
						} else {
							// 모든 방어부대(전멸상태) 주둔지 수도로 변경
							for _, defender := range defenders {
								defender.Encampment = defenderInfo.Realm.Capitals[0]
							}
						}
					} else {
						// 공격부대 주둔지로 복귀
						err := returnToEncampment(ctx, q, action)
						if err != nil {
							return err
						}
					}
				} else if !bAttackerAnnihilation { // 공격, 수비 둘 다 생존한 경우
					// 공격부대 주둔지로 복귀
					err = returnToEncampment(ctx, q, action)
					if err != nil {
						return err
					}
				}
				// 공격, 방어 부대 정보 업데이트
				err = updateLevies(ctx, q, append(defenders, &action.Levy))
				if err != nil {
					return err
				}
			case util.Move:

			}

			arg := db.UpdateLevyActionCompletedParams{
				LevyActionID: action.LeviesAction.LevyActionID,
				Completed:    true,
			}
			// action 업데이트
			err := q.UpdateLevyActionCompleted(*ctx, &arg)
			if err != nil {
				return err
			}

			return nil
		})
	}

	return nil
}

func annexSectorOfRealm(
	ctx *context.Context,
	q *db.Queries,
	action *db.FindLevyActionsBeforeDateRow,
	defenderRealm *db.Realm,
) (annexResult util.AnnexResult, error error) {
	SectorsNumber, err := q.GetNumberOfRealmSectors(*ctx, defenderRealm.RealmID)
	if err != nil {
		return util.Error, err
	}

	numberOfCapitals := len(defenderRealm.Capitals)
	bCapital := util.Find(defenderRealm.Capitals, action.LeviesAction.TargetSector)
	// 현재 영토가 방어측의 마지막 영토라면
	if SectorsNumber <= 1 {
		// 국가 멸망
		err := q.RemoveRealm(*ctx, defenderRealm.RealmID)
		if err != nil {
			return util.Error, err
		}
		annexResult = util.Annihilation
	} else if bCapital { // 수도가 함락되었을 경우
		// 수도가 모두 함락되었을 경우
		if numberOfCapitals == 1 && defenderRealm.Capitals[0] == action.LeviesAction.TargetSector {
			arg1 := db.TransferSectorOwnershipToAttackersParams{
				AttackerRealmID: action.Levy.RealmID.Int64,
				DefenderRealmID: defenderRealm.RealmID,
			}
			// 주둔중인 부대가 없는 수비측 영토는 모두 공격측 영토로 변경
			err := q.TransferSectorOwnershipToAttackers(*ctx, &arg1)
			if err != nil {
				return util.Error, err
			}

			arg2 := db.UpdateRealmPoliticalEntityAndRemoveCapitalParams{
				RealmID:         defenderRealm.RealmID,
				PoliticalEntity: util.RestorationForces,
				RemoveCapital:   action.LeviesAction.TargetSector,
			}
			// 수비측 국가 체제를 부흥 세력으로 전환 && 수도 삭제
			err = q.UpdateRealmPoliticalEntityAndRemoveCapital(*ctx, &arg2)
			if err != nil {
				return util.Error, err
			}
			annexResult = util.AllCapitalsCaptured
		} else {
			// 아직 다른 수도가 남아있는 경우
			arg := db.RemoveCapitalParams{
				RemoveCapital: action.LeviesAction.TargetSector,
				RealmID:       defenderRealm.RealmID,
			}
			err := q.RemoveCapital(*ctx, &arg)
			if err != nil {
				return util.Error, err
			}
			annexResult = util.CapitalCaptured
		}
	} else {
		// 일반 sector가 함락되었을 경우
		annexResult = util.Captured
	}

	// sector 소유권 이전
	err = changeSectorOwnership(*ctx, q, action.LeviesAction.TargetSector, defenderRealm.RealmID, action.Levy.RealmID.Int64)
	if err != nil {
		return util.Error, err
	}

	return
}

func annexSectorOfIndigenousLand(
	ctx *context.Context,
	q *db.Queries,
	action *db.FindLevyActionsBeforeDateRow,
) error {
	// TODO: 웹소켓 서버에서 ProvinceNumber, Population 받아오기
	arg1 := db.CreateSectorParams{
		CellNumber:     action.LeviesAction.TargetSector,
		ProvinceNumber: 1,
		RealmID:        action.Levy.RealmID.Int64,
		Population:     1000,
	}

	// 점령
	err := q.CreateSector(*ctx, &arg1)
	if err != nil {
		return err
	}

	arg2 := db.AddRealmSectorJsonbParams{
		Key:     string(action.LeviesAction.TargetSector),
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
) error {
	arg1 := db.UpdateSectorOwnershipParams{
		CellNumber: sectorNumber,
		RealmID:    newOwnerRealmId,
	}
	// sector 소유권 이전
	err := q.UpdateSectorOwnership(ctx, &arg1)
	if err != nil {
		return nil
	}

	arg2 := db.AddRealmSectorJsonbParams{
		Key:     string(sectorNumber),
		Value:   sectorNumber,
		RealmID: newOwnerRealmId,
	}
	// 승리한 국가에 소유권 추가
	err = q.AddRealmSectorJsonb(ctx, &arg2)
	if err != nil {
		return err
	}

	arg3 := db.RemoveSectorJsonbParams{
		CellNumber: sectorNumber,
		RealmID:    prevOwnerRealmId,
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
			MovementSpeed: util.CalculateLevyAdvanceSpeed(levy),
		}
		err := q.UpdateLevy(*ctx, &arg)
		if err != nil {
			return err
		}
	}
	return nil
}

func returnToEncampment(ctx *context.Context, q *db.Queries, action *db.FindLevyActionsBeforeDateRow) error {
	completionAt := action.LeviesAction.ExpectedCompletionAt.Add(
		action.LeviesAction.ExpectedCompletionAt.Sub(action.LeviesAction.StartedAt),
	)

	arg := db.CreateLevyActionParams{
		LevyID:               action.LeviesAction.LevyID,
		OriginSector:         action.LeviesAction.TargetSector,
		TargetSector:         action.LeviesAction.OriginSector,
		ActionType:           util.Return,
		Completed:            false,
		StartedAt:            action.LeviesAction.ExpectedCompletionAt,
		ExpectedCompletionAt: completionAt,
	}

	return q.CreateLevyAction(*ctx, &arg)
}

func executeSingleBattleTurn(attacker *db.Levy, defender *db.Levy) (attackerAnnihilation bool, defenderAnnihilation bool) {
	defenderAnnihilation = attack(attacker, defender)
	if defenderAnnihilation {
		attackerAnnihilation = false
		return
	}
	attackerAnnihilation = attack(defender, attacker)
	return
}

func attack(attacker *db.Levy, defender *db.Levy) (defenderAnnihilation bool) {
	// 궁병 공격: 상대 주둔중 -> (궁병, 창기병, 검병, 방패병), 상대 출병 -> 보급병
	annihilation := archersFirstAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 창기병 공격: 상대 주둔중 -> (방패병, 창기병, 검병, 궁병), 상대 출병 -> 보급병
	annihilation = lancersAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 검병 공격: 상대 주둔중 -> (창기병, 검병, 방패병, 궁병), 상대 출병 -> 보급병
	annihilation = swordmenAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 궁병 공격: 상대 주둔중 -> (창기병, 궁병, 검병, 방패병), 상대 출병 -> 보급병
	annihilation = archersAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 방패병 공격: 상대 주둔중 -> (검병, 창기병, 방패병, 궁병), 상대 출병 -> 보급병
	annihilation = shieldBearerAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	return false
}

func archersFirstAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed:
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
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func lancersAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed:
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
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func swordmenAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed:
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
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func archersAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed:
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
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func shieldBearerAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 수비측 출전 상태
	case !defender.Stationed:
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
