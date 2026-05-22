package stage

import (
	"context"
	"errors"
	"math"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
	"golang.org/x/sync/errgroup"
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

type datastoreServiceIface interface {
	GetStages(ctx context.Context, startStageNo, limit int) ([]datastoreservice.KyouenPuzzle, []*datastore.Key, error)
	GetUserByID(ctx context.Context, userID string) (*datastoreservice.User, *datastore.Key, error)
	GetClearedStageKeyIDs(ctx context.Context, userKey *datastore.Key) (map[int64]time.Time, error)
	CheckStageExists(ctx context.Context, stageString string) (bool, error)
	CreateStage(ctx context.Context, stage datastoreservice.KyouenPuzzle) (*datastoreservice.KyouenPuzzle, error)
	GetStageByNo(ctx context.Context, stageNo int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error)
	GetOrCreateGuestUser(ctx context.Context) (*datastoreservice.User, *datastore.Key, error)
	CreateStageUser(ctx context.Context, stageKey *datastore.Key, userKey *datastore.Key) error
	HasStageUser(ctx context.Context, stageKey *datastore.Key, userKey *datastore.Key) (bool, error)
	GetClearedStagesByUser(ctx context.Context, userKey *datastore.Key) ([]datastoreservice.StageUser, error)
	GetStagesByKeys(ctx context.Context, keys []*datastore.Key) ([]datastoreservice.KyouenPuzzle, error)
	GetRecentActivities(ctx context.Context, limit int) ([]datastoreservice.StageUser, error)
	GetUsersByKeys(ctx context.Context, keys []*datastore.Key) ([]datastoreservice.User, error)
	DeleteUser(ctx context.Context, userID string) error
}

type ClearedStageResult struct {
	StageNo   int64
	ClearDate time.Time
}

type Service struct {
	datastoreService datastoreServiceIface
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

	var stage *datastoreservice.KyouenPuzzle
	var stageKeys []*datastore.Key
	var user *datastoreservice.User
	var userKey *datastore.Key

	// errgroup.WithContext を使い goroutine から実エラーを返す。
	// 片方の失敗で gctx がキャンセルされ、もう片方が context.Canceled で終わることがあるため、
	// context.Canceled の場合は stageErr/userErr を設定しない。
	// Wait 後に stage→mismatch→user の優先順で返すことでエラー選択を決定的にする。
	var stageErr, userErr error
	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		stage, stageKeys, err = s.datastoreService.GetStageByNo(gctx, stageNo)
		if err != nil {
			if !errors.Is(err, context.Canceled) {
				stageErr = ErrStageNotFound
			}
			return err
		}
		return nil
	})
	g.Go(func() error {
		var err error
		if auth.IsGuestUser(userUID) {
			// GetOrCreateGuestUser may run concurrently with the stage lookup and
			// could create the guest user even if the stage is not found. However,
			// the guest user always exists in production, so no write occurs here.
			user, userKey, err = s.datastoreService.GetOrCreateGuestUser(gctx)
		} else {
			user, userKey, err = s.datastoreService.GetUserByID(gctx, userUID)
		}
		if err != nil && !errors.Is(err, context.Canceled) {
			userErr = ErrUserNotFound
			return userErr
		}
		return nil
	})
	if err := g.Wait(); err != nil && stageErr == nil && userErr == nil {
		// stageErr/userErr が未設定のままの場合（context.Canceled 等の予期しないエラー）
		return nil, err
	}

	// stage エラーを優先して返すことで、エラー返却順序を決定的にする
	if stageErr != nil {
		return nil, stageErr
	}
	// userErr により gctx がキャンセルされ stage 取得が中断された場合、
	// stage は nil になり得る。その場合は userErr を優先して返す。
	if stage == nil {
		if userErr != nil {
			return nil, userErr
		}
		return nil, ErrStageNotFound
	}
	if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
		return nil, ErrStageMismatch
	}
	if userErr != nil {
		return nil, userErr
	}

	if err := s.datastoreService.CreateStageUser(ctx, stageKeys[0], userKey); err != nil {
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

	var users []datastoreservice.User
	var stages []datastoreservice.KyouenPuzzle

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		users, err = s.datastoreService.GetUsersByKeys(gctx, uUserKeys)
		return err
	})
	g.Go(func() error {
		var err error
		stages, err = s.datastoreService.GetStagesByKeys(gctx, uStageKeys)
		return err
	})
	if err := g.Wait(); err != nil {
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
