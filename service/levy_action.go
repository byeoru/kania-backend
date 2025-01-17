package service

import (
	"database/sql"
	"sync"

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

func (s *LevyActionService) ExecuteTargetSectorActions(ctx *gin.Context, arg *db.FindTargetLevyActionsSortedByDateForUpdateParams, encampment int32) error {
	err := s.store.ExecTx(ctx, func(q *db.Queries) error {
		actions, err := s.store.FindTargetLevyActionsSortedByDateForUpdate(ctx, arg)
		if err != nil {
			return err
		}

		if len(actions) == 0 {
			return nil
		}

		currentOccupyLevies, err := s.store.FindStationedLevies(ctx, encampment)
		if err != nil {
			return err
		}

		// 점령중인 부대
		var defenders []*db.Levy = currentOccupyLevies
		var defenderIdx = 0
		bSectorOwnershipChanged := false

		for _, action := range actions {
			switch action.LeviesAction.ActionType {
			case util.Attack:
				defenderSize := len(defenders)
				// 방어 부대가 없는 경우
				if defenderSize == 0 {
					defenders = []*db.Levy{&action.Levy}
					bSectorOwnershipChanged = true
					continue
				}
				// 현재 졈령한 부대와 국가가 같은 경우
				if defenders[0].RealmMemberID == action.Levy.RealmMemberID {
					defenders = append(defenders, &action.Levy)
					continue
				}

				defenderAnnihilation := s.executeSingleBattleTurn(&action.Levy, defenders[defenderIdx])
				// 방어측이 전멸한 경우
				if defenderAnnihilation {
					defenderIdx++
					if defenderIdx > defenderSize-1 {
						// 뺏긴 영토가 방어측 국가의 마지막남은 영토인지 확인
						realmId, err := s.store.GetSectorRealmId(ctx, encampment)
						if err != nil {
							return nil
						}

						realmAndSectorCount, err := s.store.FindRealmAndSectorCount(ctx, realmId)
						if err != nil {
							return nil
						}

						// 모든 영토가 점령당한 경우
						if realmAndSectorCount.SectorCount <= 1 {
							// delete realm
							err := s.store.RemoveRealm(ctx, realmId)
							if err != nil {
								return nil
							}
							// 모든 수도가 점령당했을 경우 부흥 세력으로 전환()
						} else if len(realmAndSectorCount.Capitals) == 1 && realmAndSectorCount.Capitals[0] == encampment {
							// 공격부대의 국가가 없을 경우
							if !action.Levy.RealmID.Valid {
								// 주둔중인 부대가 없는 군의 경우 토착세력으로 변경(sector table 삭제)
								s.store.UpdateSectorToIndigenous(ctx, realmId)
							}
						}
						defenderIdx = 0
						defenders = []*db.Levy{&action.Levy}
						bSectorOwnershipChanged = true
					}
				}
			case util.Move:
				// TODO: 부대 주둔지 이동 구현
			}
			action.LeviesAction.Completed = true
		}

		// 최종적으로 점령한 국가 부대들의 주둔지 변경
		for _, levy := range defenders {
			levy.Encampment = encampment
		}

		// sector 소유권 변경
		if bSectorOwnershipChanged {
			realmId, err := s.store.GetRealmId(ctx, sql.NullInt64{Int64: defenders[0].RealmMemberID, Valid: true})
			if err != nil {
				return err
			}
			arg1 := db.UpdateSectorOwnershipParams{
				CellNumber: encampment,
				RealmID:    realmId,
			}
			err = s.store.UpdateSectorOwnership(ctx, &arg1)
			if err != nil {
				return nil
			}

			arg2 := db.AddRealmSectorJsonbParams{
				Key:     string(encampment),
				Value:   encampment,
				RealmID: realmId,
			}
			// 승리한 국가에 소유권 추가
			err = s.store.AddRealmSectorJsonb(ctx, &arg2)
			if err != nil {
				return err
			}

			prevRealmId, err := s.store.GetSectorRealmId(ctx, encampment)
			if err != nil {
				return err
			}
			arg3 := db.RemoveSectorJsonbParams{
				CellNumber: encampment,
				RealmID:    prevRealmId,
			}
			// 패배한 국가에 소유권 박탈
			err = s.store.RemoveSectorJsonb(ctx, &arg3)
			if err != nil {
				return err
			}
		}

		// 공격측 levy update and set action completed
		for _, action := range actions {
			arg1 := db.UpdateLevyParams{
				LevyID:        action.Levy.LevyID,
				Encampment:    action.Levy.Encampment,
				Swordmen:      action.Levy.Swordmen,
				ShieldBearers: action.Levy.ShieldBearers,
				Archers:       action.Levy.Archers,
				Lancers:       action.Levy.Lancers,
				SupplyTroop:   action.Levy.SupplyTroop,
				MovementSpeed: util.CalculateLevyAdvanceSpeed(&action.Levy),
			}
			// levy update
			err := s.store.UpdateLevy(ctx, &arg1)
			if err != nil {
				return err
			}

			arg2 := db.UpdateLevyActionCompletedParams{
				LevyActionID: action.LeviesAction.LevyActionID,
				Completed:    action.LeviesAction.Completed,
			}
			// action update
			err = s.store.UpdateLevyActionCompleted(ctx, &arg2)
			if err != nil {
				return err
			}
		}

		// 방어측 levy update
		for _, levy := range currentOccupyLevies {
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
			// levy update
			err := s.store.UpdateLevy(ctx, &arg)
			if err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *LevyActionService) executeSingleBattleTurn(attacker *db.Levy, defender *db.Levy) (defenderAnnihilation bool) {
	defenderAnnihilation = s.attack(attacker, defender)
	if defenderAnnihilation {
		return
	}
	s.attack(defender, attacker)
	return false
}

func (s *LevyActionService) attack(attacker *db.Levy, defender *db.Levy) (defenderAnnihilation bool) {
	// 궁병 공격 (궁병, 창기병, 검병, 방패병)
	annihilation := s.archersFirstAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 창기병 공격 (방패병, 창기병, 검병, 궁병)
	annihilation = s.lancersAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 검병 공격 (창기병, 검병, 방패병, 궁병)
	annihilation = s.swordmenAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 궁병 공격 (창기병, 궁병, 검병, 방패병)
	annihilation = s.archersAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	// 방패병 공격 (검병, 창기병, 방패병, 궁병)
	annihilation = s.shieldBearerAttack(attacker, defender)
	if annihilation {
		defenderAnnihilation = true
		return
	}
	return false
}

func (s *LevyActionService) archersFirstAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 궁병 -> 궁병 공격
	case defender.Archers > 0:
		if defender.ShieldBearers > 0 {
			defenderExtra := util.ExtraStat{
				AddedDefensive: 6,
			}
			defender.Archers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &defenderExtra)
		} else {
			defender.Archers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
		}
	// 궁병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 궁병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 궁병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func (s *LevyActionService) lancersAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 창기병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = s.inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 창기병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = s.inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 창기병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = s.inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 창기병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = s.inflictDamage(attacker.Lancers, unitStat.Lancer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func (s *LevyActionService) swordmenAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 검병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = s.inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 검병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = s.inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 검병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = s.inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 검병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = s.inflictDamage(attacker.Swordmen, unitStat.Swordman, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func (s *LevyActionService) archersAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 궁병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
		// 궁병 -> 궁병 공격
	case defender.Archers > 0:
		if defender.ShieldBearers > 0 {
			defenderExtra := util.ExtraStat{
				AddedDefensive: 6,
			}
			defender.Archers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &defenderExtra)
		} else {
			defender.Archers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
		}
	// 궁병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 궁병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = s.inflictDamage(attacker.Archers, unitStat.Archer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func (s *LevyActionService) shieldBearerAttack(attacker *db.Levy, defender *db.Levy) (annihilation bool) {
	unitStat := util.GetUnitStat()
	emptyExtra := util.ExtraStat{}
	switch {
	// 방패병 -> 검병 공격
	case defender.Swordmen > 0:
		defender.Swordmen = s.inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Swordmen, unitStat.Swordman, &emptyExtra)
	// 방패병 -> 창기병 공격
	case defender.Lancers > 0:
		defender.Lancers = s.inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Lancers, unitStat.Lancer, &emptyExtra)
	// 방패병 -> 방패병 공격
	case defender.ShieldBearers > 0:
		defender.ShieldBearers = s.inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.ShieldBearers, unitStat.ShieldBearer, &emptyExtra)
	// 방패병 -> 궁병 공격
	case defender.Archers > 0:
		defender.Archers = s.inflictDamage(attacker.ShieldBearers, unitStat.ShieldBearer, &emptyExtra, defender.Archers, unitStat.Archer, &emptyExtra)
	// 적 부대 전멸
	default:
		annihilation = true
		return
	}
	annihilation = false
	return
}

func (s *LevyActionService) inflictDamage(
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
