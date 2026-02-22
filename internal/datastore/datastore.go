package datastore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
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
func (s *DatastoreService) GetSummary() (*KyouenPuzzleSummary, error) {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)
	var summary KyouenPuzzleSummary

	err := s.client.Get(s.ctx, key, &summary)
	if err == datastore.ErrNoSuchEntity {
		// Create initial summary if it doesn't exist
		summary = KyouenPuzzleSummary{
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
func (s *DatastoreService) GetStages(startStageNo int, limit int) ([]KyouenPuzzle, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").
		FilterField("stageNo", ">=", startStageNo).
		Order("stageNo").
		Limit(limit)

	_, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return nil, fmt.Errorf("failed to get stages: %w", err)
	}

	return stages, nil
}

func (s *DatastoreService) GetRecentStages(limit int) ([]KyouenPuzzle, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").
		Order("-stageNo").
		Limit(limit)

	_, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent stages: %w", err)
	}

	return stages, nil
}

func (s *DatastoreService) GetStageByNo(stageNo int) (*KyouenPuzzle, []*datastore.Key, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").FilterField("stageNo", "=", stageNo).Limit(1)

	keys, err := s.client.GetAll(s.ctx, query, &stages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stage: %w", err)
	}

	if len(stages) == 0 {
		return nil, nil, fmt.Errorf("stage not found: %d", stageNo)
	}

	return &stages[0], keys, nil
}

func (s *DatastoreService) CreateStage(stage KyouenPuzzle) (*KyouenPuzzle, error) {
	// Get next stage number
	nextStageNo, err := s.getNextStageNo()
	if err != nil {
		return nil, fmt.Errorf("failed to get next stage number: %w", err)
	}

	stage.StageNo = nextStageNo
	stage.RegistDate = time.Now()

	key := datastore.IncompleteKey("KyouenPuzzle", nil)
	stageKey, err := s.client.Put(s.ctx, key, &stage)
	if err != nil {
		return nil, fmt.Errorf("failed to save stage: %w", err)
	}

	// Create RegistModel record
	if err := s.createRegistModel(stageKey); err != nil {
		// Log error but don't fail the creation
		fmt.Printf("Warning: failed to create RegistModel: %v\n", err)
	}

	// Update summary
	if err := s.updateSummary(); err != nil {
		// Log error but don't fail the creation
		fmt.Printf("Warning: failed to update summary: %v\n", err)
	}

	return &stage, nil
}

func (s *DatastoreService) getNextStageNo() (int64, error) {
	var stages []KyouenPuzzle
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
	query := datastore.NewQuery("KyouenPuzzle").FilterField("stage", "=", stageString).Limit(1)
	count, err := s.client.Count(s.ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to check stage existence: %w", err)
	}
	return count > 0, nil
}

func (s *DatastoreService) updateSummary() error {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)

	_, err := s.client.RunInTransaction(s.ctx, func(tx *datastore.Transaction) error {
		var summary KyouenPuzzleSummary
		err := tx.Get(key, &summary)
		if err == datastore.ErrNoSuchEntity {
			// Count all stages
			query := datastore.NewQuery("KyouenPuzzle").KeysOnly()
			count, err := s.client.Count(s.ctx, query)
			if err != nil {
				return fmt.Errorf("failed to count stages: %w", err)
			}
			summary = KyouenPuzzleSummary{
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
func (s *DatastoreService) GetUserByID(userID string) (*User, *datastore.Key, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	var user User

	err := s.client.Get(s.ctx, key, &user)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil, fmt.Errorf("user not found: %s", userID)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, key, nil
}

func (s *DatastoreService) UpsertUser(user User, userID string) (*User, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	_, err := s.client.Put(s.ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	return &user, nil
}

// CreateOrUpdateUserFromFirebase creates or updates a user from Firebase authentication data
func (s *DatastoreService) CreateOrUpdateUserFromFirebase(firebaseUID, screenName, image, twitterUID string) (*User, error) {
	// Try to get existing user
	existingUser, _, err := s.GetUserByID(firebaseUID)
	if err != nil {
		// User doesn't exist, create new one
		newUser := User{
			UserID:          firebaseUID,
			ScreenName:      screenName,
			Image:           image,
			TwitterUID:      twitterUID,
			ClearStageCount: 0,
		}
		return s.UpsertUser(newUser, firebaseUID)
	}

	// User exists, update information if needed
	updated := false
	if existingUser.ScreenName != screenName {
		existingUser.ScreenName = screenName
		updated = true
	}
	if existingUser.Image != image {
		existingUser.Image = image
		updated = true
	}
	if existingUser.TwitterUID != twitterUID && twitterUID != "" {
		existingUser.TwitterUID = twitterUID
		updated = true
	}

	if updated {
		return s.UpsertUser(*existingUser, firebaseUID)
	}

	return existingUser, nil
}

// GetOrCreateGuestUser gets or creates the guest user account (matches existing production data)
func (s *DatastoreService) GetOrCreateGuestUser() (*User, *datastore.Key, error) {
	guestUID := "0"

	// Try to get existing guest user
	existingUser, key, err := s.GetUserByID(guestUID)
	if err != nil {
		// Guest user doesn't exist, create new one with exact production data format
		guestUser := User{
			UserID:          guestUID,
			ScreenName:      "Guest",
			Image:           "https://kyouen.app/image/icon.png",
			TwitterUID:      "",
			ClearStageCount: 0,
		}

		user, err := s.UpsertUser(guestUser, guestUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create guest user: %w", err)
		}

		// Return with the key
		guestKey := datastore.NameKey("User", "KEY"+guestUID, nil)
		return user, guestKey, nil
	}

	return existingUser, key, nil
}

// StageUser operations
func (s *DatastoreService) CreateStageUser(stageKey *datastore.Key, userKey *datastore.Key) error {
	// Check if already exists
	var stageUsers []StageUser
	query := datastore.NewQuery("StageUser").
		FilterField("stage", "=", stageKey).
		FilterField("user", "=", userKey).
		Limit(1)

	keys, err := s.client.GetAll(s.ctx, query, &stageUsers)
	if err != nil {
		return fmt.Errorf("failed to check existing StageUser: %w", err)
	}

	if len(stageUsers) == 0 {
		// Create new StageUser
		stageUser := StageUser{
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

// HasStageUser checks if a stage user relation exists
func (s *DatastoreService) HasStageUser(stageKey *datastore.Key, userKey *datastore.Key) (bool, error) {
	query := datastore.NewQuery("StageUser").
		FilterField("stage", "=", stageKey).
		FilterField("user", "=", userKey).
		Limit(1)

	var stageUsers []StageUser
	_, err := s.client.GetAll(s.ctx, query, &stageUsers)
	if err != nil {
		return false, fmt.Errorf("failed to check StageUser existence: %w", err)
	}

	return len(stageUsers) > 0, nil
}

// GetClearedStagesByUser gets all cleared stages for a user
func (s *DatastoreService) GetClearedStagesByUser(userKey *datastore.Key) ([]StageUser, error) {
	query := datastore.NewQuery("StageUser").
		FilterField("user", "=", userKey).
		Order("clearDate")

	var stageUsers []StageUser
	_, err := s.client.GetAll(s.ctx, query, &stageUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleared stages: %w", err)
	}

	return stageUsers, nil
}

func (s *DatastoreService) IncrementUserClearCount(userKey *datastore.Key) error {
	_, err := s.client.RunInTransaction(s.ctx, func(tx *datastore.Transaction) error {
		var user User
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

// DeleteUser deletes a user and all associated data from Datastore
func (s *DatastoreService) DeleteUser(userID string) error {
	// Get user key and entity
	userKey := datastore.NameKey("User", "KEY"+userID, nil)
	var user User
	err := s.client.Get(s.ctx, userKey, &user)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return fmt.Errorf("user not found: %s", userID)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: Add audit log for user account deletion (required for compliance)

	// Run deletion in transaction
	_, err = s.client.RunInTransaction(s.ctx, func(tx *datastore.Transaction) error {
		// Delete all StageUser records for this user
		query := datastore.NewQuery("StageUser").FilterField("user", "=", userKey)
		keys, err := s.client.GetAll(s.ctx, query, &[]StageUser{})
		if err != nil {
			return fmt.Errorf("failed to get StageUser records: %w", err)
		}

		// Delete StageUser records
		if len(keys) > 0 {
			err = tx.DeleteMulti(keys)
			if err != nil {
				return fmt.Errorf("failed to delete StageUser records: %w", err)
			}
		}

		// Anonymize creator field in KyouenPuzzle entities created by this user
		stageQuery := datastore.NewQuery("KyouenPuzzle").FilterField("creator", "=", user.ScreenName)
		var stages []KyouenPuzzle
		stageKeys, err := s.client.GetAll(s.ctx, stageQuery, &stages)
		if err != nil {
			return fmt.Errorf("failed to get user's stages: %w", err)
		}

		// Anonymize creator field
		for i, stage := range stages {
			stage.Creator = "[deleted user]"
			_, err = tx.Put(stageKeys[i], &stage)
			if err != nil {
				return fmt.Errorf("failed to anonymize stage creator: %w", err)
			}
		}

		// Delete the user entity
		err = tx.Delete(userKey)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})

	return err
}

// GetRecentActivities gets recent user activities (stage completions)
func (s *DatastoreService) GetRecentActivities(limit int) ([]StageUser, error) {
	var stageUsers []StageUser
	query := datastore.NewQuery("StageUser").
		Order("-clearDate").
		Limit(limit)

	_, err := s.client.GetAll(s.ctx, query, &stageUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}

	return stageUsers, nil
}

// GetUserByKey gets a user by datastore key
func (s *DatastoreService) GetUserByKey(userKey *datastore.Key) (*User, error) {
	var user User
	err := s.client.Get(s.ctx, userKey, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by key: %w", err)
	}
	return &user, nil
}

// GetStageByKey gets a stage by datastore key
func (s *DatastoreService) GetStageByKey(stageKey *datastore.Key) (*KyouenPuzzle, error) {
	var stage KyouenPuzzle
	err := s.client.Get(s.ctx, stageKey, &stage)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage by key: %w", err)
	}
	return &stage, nil
}

// createRegistModel creates a RegistModel record for a given stage
func (s *DatastoreService) createRegistModel(stageKey *datastore.Key) error {
	registModel := RegistModel{
		StageInfo:  stageKey,
		RegistDate: time.Now(),
	}

	key := datastore.IncompleteKey("RegistModel", nil)
	_, err := s.client.Put(s.ctx, key, &registModel)
	if err != nil {
		return fmt.Errorf("failed to save RegistModel: %w", err)
	}

	return nil
}
