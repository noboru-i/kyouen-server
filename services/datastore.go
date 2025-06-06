package services

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
	"kyouen-server/db"
)

type DatastoreService struct {
	client *datastore.Client
	ctx    context.Context
}

func NewDatastoreService(projectID string) (*DatastoreService, error) {
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Datastore client: %w", err)
	}

	return &DatastoreService{
		client: client,
		ctx:    ctx,
	}, nil
}

func (s *DatastoreService) Close() error {
	return s.client.Close()
}

// GetClient returns the underlying Datastore client for compatibility
func (s *DatastoreService) GetClient() *datastore.Client {
	return s.client
}

// Statistics operations
func (s *DatastoreService) GetSummary() (*db.KyouenPuzzleSummary, error) {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)
	var summary db.KyouenPuzzleSummary
	
	err := s.client.Get(s.ctx, key, &summary)
	if err == datastore.ErrNoSuchEntity {
		// Create initial summary if it doesn't exist
		summary = db.KyouenPuzzleSummary{
			Count:    0,
			LastDate: time.Now(),
		}
		_, err = s.client.Put(s.ctx, key, &summary)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial summary: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}
	
	return &summary, nil
}

// Stages operations
func (s *DatastoreService) GetStages(startStageNo int, limit int) ([]db.KyouenPuzzle, error) {
	var stages []db.KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").
		Filter("stageNo >=", startStageNo).
		Order("stageNo").
		Limit(limit)
	
	_, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}
	
	return stages, nil
}

func (s *DatastoreService) GetStageByNo(stageNo int) (*db.KyouenPuzzle, []*datastore.Key, error) {
	var stages []db.KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").Filter("stageNo =", stageNo).Limit(1)
	
	keys, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stage: %w", err)
	}
	
	if len(stages) == 0 {
		return nil, nil, fmt.Errorf("stage not found: %d", stageNo)
	}
	
	return &stages[0], keys, nil
}

func (s *DatastoreService) CreateStage(stage db.KyouenPuzzle) (*db.KyouenPuzzle, error) {
	// Get next stage number
	nextStageNo, err := s.getNextStageNo()
	if err != nil {
		return nil, fmt.Errorf("failed to get next stage number: %w", err)
	}
	
	stage.StageNo = nextStageNo
	stage.RegistDate = time.Now()
	
	key := datastore.IncompleteKey("KyouenPuzzle", nil)
	_, err = s.client.Put(s.ctx, key, &stage)
	if err != nil {
		return nil, fmt.Errorf("failed to save stage: %w", err)
	}
	
	// Update summary
	if err := s.updateSummary(); err != nil {
		// Log error but don't fail the creation
		fmt.Printf("Warning: failed to update summary: %v\n", err)
	}
	
	return &stage, nil
}

func (s *DatastoreService) getNextStageNo() (int64, error) {
	var stages []db.KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").Order("-stageNo").Limit(1)
	
	_, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return 0, fmt.Errorf("failed to query latest stage: %w", err)
	}
	
	if len(stages) == 0 {
		return 1, nil
	}
	
	return stages[0].StageNo + 1, nil
}

func (s *DatastoreService) CheckStageExists(stageString string) (bool, error) {
	query := datastore.NewQuery("KyouenPuzzle").Filter("stage =", stageString).Limit(1)
	count, err := s.client.Count(s.ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to check stage existence: %w", err)
	}
	return count > 0, nil
}

func (s *DatastoreService) updateSummary() error {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)
	
	_, err := s.client.RunInTransaction(s.ctx, func(tx *datastore.Transaction) error {
		var summary db.KyouenPuzzleSummary
		err := tx.Get(key, &summary)
		if err == datastore.ErrNoSuchEntity {
			// Count all stages
			query := datastore.NewQuery("KyouenPuzzle").KeysOnly()
			count, err := s.client.Count(s.ctx, query)
			if err != nil {
				return fmt.Errorf("failed to count stages: %w", err)
			}
			summary = db.KyouenPuzzleSummary{
				Count:    int64(count),
				LastDate: time.Now(),
			}
		} else if err != nil {
			return fmt.Errorf("failed to get summary: %w", err)
		} else {
			summary.Count++
			summary.LastDate = time.Now()
		}
		
		_, err = tx.Put(key, &summary)
		return err
	})
	
	return err
}

// Users operations
func (s *DatastoreService) GetUserByID(userID string) (*db.User, *datastore.Key, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	var user db.User
	
	err := s.client.Get(s.ctx, key, &user)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil, fmt.Errorf("user not found: %s", userID)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}
	
	return &user, key, nil
}

func (s *DatastoreService) UpsertUser(user db.User, userID string) (*db.User, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	_, err := s.client.Put(s.ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	return &user, nil
}

// StageUser operations
func (s *DatastoreService) CreateStageUser(stageKey *datastore.Key, userKey *datastore.Key) error {
	// Check if already exists
	var stageUsers []db.StageUser
	query := datastore.NewQuery("StageUser").
		Filter("stage =", stageKey).
		Filter("user =", userKey).
		Limit(1)
	
	keys, err := s.client.GetAll(s.ctx, query, &stageUsers)
	if err != nil {
		return fmt.Errorf("failed to check existing StageUser: %w", err)
	}
	
	if len(stageUsers) == 0 {
		// Create new StageUser
		stageUser := db.StageUser{
			StageKey:  stageKey,
			UserKey:   userKey,
			ClearDate: time.Now(),
		}
		key := datastore.IncompleteKey("StageUser", nil)
		_, err = s.client.Put(s.ctx, key, &stageUser)
		if err != nil {
			return fmt.Errorf("failed to create StageUser: %w", err)
		}
	} else {
		// Update existing StageUser
		stageUsers[0].ClearDate = time.Now()
		_, err = s.client.Put(s.ctx, keys[0], &stageUsers[0])
		if err != nil {
			return fmt.Errorf("failed to update StageUser: %w", err)
		}
	}
	
	return nil
}

func (s *DatastoreService) IncrementUserClearCount(userKey *datastore.Key) error {
	_, err := s.client.RunInTransaction(s.ctx, func(tx *datastore.Transaction) error {
		var user db.User
		err := tx.Get(userKey, &user)
		if err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}
		
		user.ClearStageCount++
		_, err = tx.Put(userKey, &user)
		return err
	})
	
	return err
}