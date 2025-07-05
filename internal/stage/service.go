package stage

import (
	"errors"
	"math"
	"strings"

	"kyouen-server/internal/datastore"
	"kyouen-server/pkg/models"
	"kyouen-server/internal/generated/openapi"
)

var (
	ErrInsufficientStones = errors.New("stage must have 5 stones")
	ErrNoKyouen          = errors.New("sent stage don't have kyouen")
	ErrStageExists       = errors.New("sent stage is already exists")
	ErrInvalidKyouen     = errors.New("invalid kyouen")
	ErrStageNotFound     = errors.New("stage not found")
	ErrStageMismatch     = errors.New("stage mismatch")
	ErrUserNotFound      = errors.New("user not found")
)

type Service struct {
	datastoreService *datastore.DatastoreService
}

func NewService(datastoreService *datastore.DatastoreService) *Service {
	return &Service{
		datastoreService: datastoreService,
	}
}

func (s *Service) CreateStage(param openapi.NewStage, creatorName string) (*datastore.KyouenPuzzle, error) {
	// Validate stage using existing business logic
	stage := *models.NewKyouenStage(int(param.Size), param.Stage)
	
	// Check stone count
	if stage.StoneCount() <= 4 {
		return nil, ErrInsufficientStones
	}
	
	// Check if stage has kyouen
	kyouenData := stage.HasKyouen()
	if kyouenData == nil {
		return nil, ErrNoKyouen
	}
	
	// Check if stage already exists (including rotations and reflections)
	exists, err := s.hasRegisteredStageAll(stage)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrStageExists
	}
	
	// Create stage with authenticated user as creator
	newStage := datastore.KyouenPuzzle{
		Size:    param.Size,
		Stage:   param.Stage,
		Creator: creatorName,
	}
	
	return s.datastoreService.CreateStage(newStage)
}

func (s *Service) ClearStage(stageNo int, stageData string, userUID string) (*datastore.User, error) {
	// Validate clear stage using existing business logic
	size := int(math.Sqrt(float64(len(stageData))))
	paramKyouenStage := models.NewKyouenStage(size, stageData)
	
	if !isKyouen(paramKyouenStage) {
		return nil, ErrInvalidKyouen
	}
	
	// Get stage from database
	stage, stageKeys, err := s.datastoreService.GetStageByNo(stageNo)
	if err != nil {
		return nil, ErrStageNotFound
	}
	
	// Verify stage matches
	if stage.Stage != strings.Replace(paramKyouenStage.ToString(), "2", "1", -1) {
		return nil, ErrStageMismatch
	}
	
	// Get user from database
	user, userKey, err := s.datastoreService.GetUserByID(userUID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	
	// Create stage user relation to record the clear
	err = s.datastoreService.CreateStageUser(stageKeys[0], userKey)
	if err != nil {
		return nil, err
	}
	
	return user, nil
}

func (s *Service) SyncStages(userUID string, clientClearedStages []openapi.ClearedStage) ([]datastore.StageUser, error) {
	// Get user from database
	_, userKey, err := s.datastoreService.GetUserByID(userUID)
	if err != nil {
		return nil, ErrUserNotFound
	}
	
	// For each client cleared stage, create stage user relation if not exists
	for _, clearedStage := range clientClearedStages {
		// Get stage by stage number
		_, stageKeys, err := s.datastoreService.GetStageByNo(int(clearedStage.StageNo))
		if err != nil {
			// Skip stages that don't exist
			continue
		}
		
		// Check if stage user relation already exists
		exists, err := s.datastoreService.HasStageUser(stageKeys[0], userKey)
		if err != nil {
			continue
		}
		
		if !exists {
			// Create stage user relation
			err = s.datastoreService.CreateStageUser(stageKeys[0], userKey)
			if err != nil {
				continue
			}
		}
	}
	
	// Get all cleared stages for this user from server
	return s.datastoreService.GetClearedStagesByUser(userKey)
}

// Helper function to check if stage exists in all rotations and reflections
func (s *Service) hasRegisteredStageAll(stage models.KyouenStage) (bool, error) {
	for i := 0; i < 4; i++ {
		mirror := models.NewMirroredKyouenStage(stage)
		exists, err := s.datastoreService.CheckStageExists(mirror.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
		
		stage = *models.NewRotatedKyouenStage(stage)
		exists, err = s.datastoreService.CheckStageExists(stage.ToString())
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	return false, nil
}

// DeleteAccount deletes a user account and all associated data
func (s *Service) DeleteAccount(userUID string) error {
	return s.datastoreService.DeleteUser(userUID)
}

