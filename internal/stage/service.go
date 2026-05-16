package stage

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"kyouen-server/internal/auth"
	datastoreservice "kyouen-server/internal/datastore"
	"kyouen-server/internal/generated/openapi"
	"kyouen-server/pkg/models"
)

var (
	ErrInsufficientStones = errors.New("stage must have 5 stones")
	ErrNoKyouen           = errors.New("sent stage don't have kyouen")
	ErrStageExists        = errors.New("sent stage is already exists")
	ErrInvalidKyouen      = errors.New("invalid kyouen")
	ErrStageNotFound      = errors.New("stage not found")
	ErrStageMismatch      = errors.New("stage mismatch")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidStageLength = errors.New("stage length must be size * size")
)

type ClearedStageResult struct {
	StageNo   int64
	ClearDate time.Time
}

type Service struct {
	datastoreService *datastoreservice.DatastoreService
	firebaseService  *datastoreservice.FirebaseService
}

func NewService(datastoreService *datastoreservice.DatastoreService, firebaseService *datastoreservice.FirebaseService) *Service {
	return &Service{
		datastoreService: datastoreService,
		firebaseService:  firebaseService,
	}
}

func (s *Service) GetStages(ctx context.Context, startStageNo, limit int, authUID string) ([]datastoreservice.KyouenPuzzle, []*datastore.Key, map[int64]time.Time, error) {
	stages, stageKeys, err := s.datastoreService.GetStages(ctx, startStageNo, limit)
	if err != nil {
		return nil, nil, nil, err
	}

	var clearedKeyIDs map[int64]time.Time
	if authUID != "" && !auth.IsGuestUser(authUID) {
		_, userKey, userErr := s.datastoreService.GetUserByID(ctx, authUID)
		if userErr == nil {
			clearedKeyIDs, _ = s.datastoreService.GetClearedStageKeyIDs(ctx, userKey)
		}
	}

	return stages, stageKeys, clearedKeyIDs, nil
}

func (s *Service) CreateStage(ctx context.Context, param openapi.NewStage, creatorName string) (*datastoreservice.KyouenPuzzle, error) {
	if len(param.Stage) != int(param.Size)*int(param.Size) {
		return nil, ErrInvalidStageLength
	}

	stage := *models.NewKyouenStage(int(param.Size), param.Stage)

	if stage.StoneCount() <= 4 {
		return nil, ErrInsufficientStones
	}

	kyouenData := stage.HasKyouen()
	if kyouenData == nil {
		return nil, ErrNoKyouen
	}

	exists, err := s.hasRegisteredStageAll(ctx, stage)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrStageExists
	}

	newStage := datastoreservice.KyouenPuzzle{
		Size:    param.Size,
		Stage:   param.Stage,
		Creator: creatorName,
	}

	return s.datastoreService.CreateStage(ctx, newStage)
}

func (s *Service) ClearStage(ctx context.Context, stageNo int, stageData string, userUID string) (*datastoreservice.User, error) {
	size := int(math.Sqrt(float64(len(stageData))))
	paramKyouenStage := models.NewKyouenStage(size, stageData)

	if !isKyouen(paramKyouenStage) {
		return nil, ErrInvalidKyouen
	}

	stage, stageKeys, err := s.datastoreService.GetStageByNo(ctx, stageNo)
	if err != nil {
		return nil, ErrStageNotFound
	}

	if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
		return nil, ErrStageMismatch
	}

	var user *datastoreservice.User
	var userKey *datastore.Key

	if auth.IsGuestUser(userUID) {
		user, userKey, err = s.datastoreService.GetOrCreateGuestUser(ctx)
		if err != nil {
			return nil, ErrUserNotFound
		}
	} else {
		user, userKey, err = s.datastoreService.GetUserByID(ctx, userUID)
		if err != nil {
			return nil, ErrUserNotFound
		}
	}

	err = s.datastoreService.CreateStageUser(ctx, stageKeys[0], userKey)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *Service) SyncStages(ctx context.Context, userUID string, clientClearedStages []openapi.ClearedStage) ([]ClearedStageResult, error) {
	_, userKey, err := s.datastoreService.GetUserByID(ctx, userUID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	for _, clearedStage := range clientClearedStages {
		_, stageKeys, err := s.datastoreService.GetStageByNo(ctx, int(clearedStage.StageNo))
		if err != nil {
			continue
		}

		exists, err := s.datastoreService.HasStageUser(ctx, stageKeys[0], userKey)
		if err != nil {
			continue
		}

		if !exists {
			err = s.datastoreService.CreateStageUser(ctx, stageKeys[0], userKey)
			if err != nil {
				continue
			}
		}
	}

	stageUsers, err := s.datastoreService.GetClearedStagesByUser(ctx, userKey)
	if err != nil {
		return nil, err
	}

	stageKeys := make([]*datastore.Key, len(stageUsers))
	for i, su := range stageUsers {
		stageKeys[i] = su.StageKey
	}
	stages, err := s.datastoreService.GetStagesByKeys(ctx, stageKeys)
	if err != nil {
		return nil, err
	}

	results := make([]ClearedStageResult, len(stageUsers))
	for i, su := range stageUsers {
		results[i] = ClearedStageResult{
			StageNo:   stages[i].StageNo,
			ClearDate: su.ClearDate,
		}
	}
	return results, nil
}

// Helper function to check if stage exists in all rotations and reflections
func (s *Service) hasRegisteredStageAll(ctx context.Context, stage models.KyouenStage) (bool, error) {
	for i := 0; i < 4; i++ {
		mirror := models.NewMirroredKyouenStage(stage)
		exists, err := s.datastoreService.CheckStageExists(ctx, mirror.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}

		stage = *models.NewRotatedKyouenStage(stage)
		exists, err = s.datastoreService.CheckStageExists(ctx, stage.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

type ActivityStage struct {
	StageNo   int64
	ClearDate time.Time
}

type ActivityUser struct {
	UserID        string
	ScreenName    string
	Image         string
	ClearedStages []ActivityStage
}

func (s *Service) GetActivities(ctx context.Context, limit int) ([]ActivityUser, error) {
	stageUsers, err := s.datastoreService.GetRecentActivities(ctx, limit)
	if err != nil {
		return nil, err
	}
	if len(stageUsers) == 0 {
		return []ActivityUser{}, nil
	}

	userKeys := make([]*datastore.Key, len(stageUsers))
	stageKeys := make([]*datastore.Key, len(stageUsers))
	for i, su := range stageUsers {
		userKeys[i] = su.UserKey
		stageKeys[i] = su.StageKey
	}
	uUserKeys, userIdx := uniqueKeys(userKeys)
	uStageKeys, stageIdx := uniqueKeys(stageKeys)

	users, err := s.datastoreService.GetUsersByKeys(ctx, uUserKeys)
	if err != nil {
		return nil, err
	}
	stages, err := s.datastoreService.GetStagesByKeys(ctx, uStageKeys)
	if err != nil {
		return nil, err
	}

	activityMap := make(map[string]*ActivityUser)
	var order []string
	for _, su := range stageUsers {
		u := users[userIdx[su.UserKey.String()]]
		if u.UserID == "" {
			continue
		}
		st := stages[stageIdx[su.StageKey.String()]]
		if st.StageNo == 0 {
			continue
		}
		if _, ok := activityMap[u.UserID]; !ok {
			activityMap[u.UserID] = &ActivityUser{
				UserID:     u.UserID,
				ScreenName: u.ScreenName,
				Image:      u.Image,
			}
			order = append(order, u.UserID)
		}
		activityMap[u.UserID].ClearedStages = append(
			activityMap[u.UserID].ClearedStages,
			ActivityStage{StageNo: st.StageNo, ClearDate: su.ClearDate},
		)
	}

	result := make([]ActivityUser, 0, len(order))
	for _, id := range order {
		result = append(result, *activityMap[id])
	}
	return result, nil
}

func uniqueKeys(keys []*datastore.Key) ([]*datastore.Key, map[string]int) {
	idx := make(map[string]int, len(keys))
	unique := make([]*datastore.Key, 0, len(keys))
	for _, k := range keys {
		s := k.String()
		if _, ok := idx[s]; ok {
			continue
		}
		idx[s] = len(unique)
		unique = append(unique, k)
	}
	return unique, idx
}

// DeleteAccount deletes a user account and all associated data
func (s *Service) DeleteAccount(ctx context.Context, userUID string) error {
	// TODO: Add audit log for account deletion request (required for compliance)

	err := s.datastoreService.DeleteUser(ctx, userUID)
	if err != nil {
		return err
	}

	err = s.firebaseService.DeleteUser(ctx, userUID)
	if err != nil {
		// Don't fail the entire operation since Datastore deletion succeeded
		// TODO: Add error log for Firebase Auth deletion failure (for manual cleanup)
	}

	return nil
}
