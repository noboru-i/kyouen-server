package datastore

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/datastore"
)

type DatastoreService struct {
	client *datastore.Client
}

func NewDatastoreService(projectID string) (*DatastoreService, error) {
	client, err := datastore.NewClient(context.Background(), projectID)
	if err != nil {
		return nil, fmt.Errorf("failed to create Datastore client: %w", err)
	}

	return &DatastoreService{client: client}, nil
}

func (s *DatastoreService) Close() error {
	return s.client.Close()
}

// GetClient returns the underlying Datastore client for compatibility
func (s *DatastoreService) GetClient() *datastore.Client {
	return s.client
}

// Statistics operations
func (s *DatastoreService) GetSummary(ctx context.Context) (*KyouenPuzzleSummary, error) {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)
	var summary KyouenPuzzleSummary

	err := s.client.Get(ctx, key, &summary)
	if err == datastore.ErrNoSuchEntity {
		summary = KyouenPuzzleSummary{
			Count:    0,
			LastDate: time.Now(),
		}
		_, err = s.client.Put(ctx, key, &summary)
		if err != nil {
			return nil, fmt.Errorf("failed to create initial summary: %w", err)
		}
	} else if err != nil {
		return nil, fmt.Errorf("failed to get summary: %w", err)
	}

	return &summary, nil
}

// Stages operations
func (s *DatastoreService) GetStages(ctx context.Context, startStageNo int, limit int) ([]KyouenPuzzle, []*datastore.Key, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").
		FilterField("stageNo", ">=", startStageNo).
		Order("stageNo").
		Limit(limit)

	keys, err := s.client.GetAll(ctx, query, &stages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stages: %w", err)
	}

	return stages, keys, nil
}

// GetClearedStageKeyIDs returns a map of Datastore key ID to clear date for stages cleared by the given user.
func (s *DatastoreService) GetClearedStageKeyIDs(ctx context.Context, userKey *datastore.Key) (map[int64]time.Time, error) {
	stageUsers, err := s.GetClearedStagesByUser(ctx, userKey)
	if err != nil {
		return nil, err
	}
	result := make(map[int64]time.Time, len(stageUsers))
	for _, su := range stageUsers {
		result[su.StageKey.ID] = su.ClearDate
	}
	return result, nil
}

func (s *DatastoreService) GetRecentStages(ctx context.Context, limit int) ([]KyouenPuzzle, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").
		Order("-stageNo").
		Limit(limit)

	_, err := s.client.GetAll(ctx, query, &stages)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent stages: %w", err)
	}

	return stages, nil
}

func (s *DatastoreService) GetStageByNo(ctx context.Context, stageNo int) (*KyouenPuzzle, []*datastore.Key, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").FilterField("stageNo", "=", stageNo).Limit(1)

	keys, err := s.client.GetAll(ctx, query, &stages)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get stage: %w", err)
	}

	if len(stages) == 0 {
		return nil, nil, fmt.Errorf("stage not found: %d", stageNo)
	}

	return &stages[0], keys, nil
}

func (s *DatastoreService) CreateStage(ctx context.Context, stage KyouenPuzzle) (*KyouenPuzzle, error) {
	nextStageNo, err := s.getNextStageNo(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get next stage number: %w", err)
	}

	stage.StageNo = nextStageNo
	stage.RegistDate = time.Now()

	key := datastore.IncompleteKey("KyouenPuzzle", nil)
	stageKey, err := s.client.Put(ctx, key, &stage)
	if err != nil {
		return nil, fmt.Errorf("failed to save stage: %w", err)
	}

	// Log error but don't fail the creation
	if err := s.createRegistModel(ctx, stageKey); err != nil {
		fmt.Printf("Warning: failed to create RegistModel: %v\n", err)
	}

	// Log error but don't fail the creation
	if err := s.updateSummary(ctx); err != nil {
		fmt.Printf("Warning: failed to update summary: %v\n", err)
	}

	return &stage, nil
}

func (s *DatastoreService) getNextStageNo(ctx context.Context) (int64, error) {
	var stages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").Order("-stageNo").Limit(1)

	_, err := s.client.GetAll(ctx, query, &stages)
	if err != nil {
		return 0, fmt.Errorf("failed to query latest stage: %w", err)
	}

	if len(stages) == 0 {
		return 1, nil
	}

	return stages[0].StageNo + 1, nil
}

func (s *DatastoreService) CheckStageExists(ctx context.Context, stageString string) (bool, error) {
	query := datastore.NewQuery("KyouenPuzzle").FilterField("stage", "=", stageString).Limit(1)
	count, err := s.client.Count(ctx, query)
	if err != nil {
		return false, fmt.Errorf("failed to check stage existence: %w", err)
	}
	return count > 0, nil
}

func (s *DatastoreService) updateSummary(ctx context.Context) error {
	key := datastore.IDKey("KyouenPuzzleSummary", 1, nil)

	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var summary KyouenPuzzleSummary
		err := tx.Get(key, &summary)
		if err == datastore.ErrNoSuchEntity {
			query := datastore.NewQuery("KyouenPuzzle").KeysOnly()
			count, err := s.client.Count(ctx, query)
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
func (s *DatastoreService) GetUserByID(ctx context.Context, userID string) (*User, *datastore.Key, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	var user User

	err := s.client.Get(ctx, key, &user)
	if err == datastore.ErrNoSuchEntity {
		return nil, nil, fmt.Errorf("user not found: %s", userID)
	} else if err != nil {
		return nil, nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, key, nil
}

func (s *DatastoreService) UpsertUser(ctx context.Context, user User, userID string) (*User, error) {
	key := datastore.NameKey("User", "KEY"+userID, nil)
	_, err := s.client.Put(ctx, key, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to save user: %w", err)
	}
	return &user, nil
}

// MigrateLegacyUser migrates a legacy Python-era user (keyed by Twitter UID) to a new Firebase UID-based key.
// It preserves clearStageCount, records the migration in UserMigration, and re-points StageUser records.
func (s *DatastoreService) MigrateLegacyUser(ctx context.Context, firebaseUID, screenName, image, twitterUID string) (*User, error) {
	oldKey := datastore.NameKey("User", "KEY"+twitterUID, nil)
	newKey := datastore.NameKey("User", "KEY"+firebaseUID, nil)

	var migratedUser User

	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var oldUser User
		if err := tx.Get(oldKey, &oldUser); err != nil {
			return fmt.Errorf("failed to get legacy user: %w", err)
		}

		migratedUser = User{
			UserID:          firebaseUID,
			ScreenName:      screenName,
			Image:           image,
			TwitterUID:      twitterUID,
			ClearStageCount: oldUser.ClearStageCount,
		}
		if _, err := tx.Put(newKey, &migratedUser); err != nil {
			return fmt.Errorf("failed to put migrated user: %w", err)
		}

		migration := UserMigration{
			OldKey:      "KEY" + twitterUID,
			NewKey:      "KEY" + firebaseUID,
			TwitterUID:  twitterUID,
			FirebaseUID: firebaseUID,
			MigratedAt:  time.Now(),
		}
		migrationKey := datastore.IncompleteKey("UserMigration", nil)
		if _, err := tx.Put(migrationKey, &migration); err != nil {
			return fmt.Errorf("failed to save UserMigration record: %w", err)
		}

		if err := tx.Delete(oldKey); err != nil {
			return fmt.Errorf("failed to delete legacy user: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to migrate legacy user: %w", err)
	}

	// StageUser レコードのキーを更新（トランザクション外: 25エンティティグループ制限のため）
	s.migrateStageUserRecords(ctx, oldKey, newKey)

	return &migratedUser, nil
}

// migrateStageUserRecords updates StageUser records that reference the old user key to point to the new key.
func (s *DatastoreService) migrateStageUserRecords(ctx context.Context, oldUserKey, newUserKey *datastore.Key) {
	query := datastore.NewQuery("StageUser").FilterField("user", "=", oldUserKey)
	var stageUsers []StageUser
	keys, err := s.client.GetAll(ctx, query, &stageUsers)
	if err != nil {
		fmt.Printf("Warning: failed to query StageUser records for migration: %v\n", err)
		return
	}

	for i := range stageUsers {
		stageUsers[i].UserKey = newUserKey
		if _, err := s.client.Put(ctx, keys[i], &stageUsers[i]); err != nil {
			fmt.Printf("Warning: failed to migrate StageUser record %v: %v\n", keys[i], err)
		}
	}
}

// CreateOrUpdateUserFromFirebase creates or updates a user from Firebase authentication data
func (s *DatastoreService) CreateOrUpdateUserFromFirebase(ctx context.Context, firebaseUID, screenName, image, twitterUID string) (*User, error) {
	existingUser, _, err := s.GetUserByID(ctx, firebaseUID)
	if err != nil {
		// Firebase UID でユーザーが見つからない場合、レガシーユーザー（Twitter UID キー）を検索
		if twitterUID != "" {
			legacyUser, _, legacyErr := s.GetUserByID(ctx, twitterUID)
			if legacyErr == nil && legacyUser != nil {
				// Python時代のユーザーが見つかった → マイグレーション実行
				return s.MigrateLegacyUser(ctx, firebaseUID, screenName, image, twitterUID)
			}
		}

		newUser := User{
			UserID:          firebaseUID,
			ScreenName:      screenName,
			Image:           image,
			TwitterUID:      twitterUID,
			ClearStageCount: 0,
		}
		return s.UpsertUser(ctx, newUser, firebaseUID)
	}

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
		return s.UpsertUser(ctx, *existingUser, firebaseUID)
	}

	return existingUser, nil
}

// GetOrCreateGuestUser gets or creates the guest user account (matches existing production data)
func (s *DatastoreService) GetOrCreateGuestUser(ctx context.Context) (*User, *datastore.Key, error) {
	guestUID := "0"

	existingUser, key, err := s.GetUserByID(ctx, guestUID)
	if err != nil {
		// Guest user doesn't exist, create new one with exact production data format
		guestUser := User{
			UserID:          guestUID,
			ScreenName:      "Guest",
			Image:           "https://kyouen.app/image/icon.png",
			TwitterUID:      "",
			ClearStageCount: 0,
		}

		user, err := s.UpsertUser(ctx, guestUser, guestUID)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create guest user: %w", err)
		}

		guestKey := datastore.NameKey("User", "KEY"+guestUID, nil)
		return user, guestKey, nil
	}

	return existingUser, key, nil
}

// StageUser operations
func (s *DatastoreService) CreateStageUser(ctx context.Context, stageKey *datastore.Key, userKey *datastore.Key) error {
	var stageUsers []StageUser
	query := datastore.NewQuery("StageUser").
		FilterField("stage", "=", stageKey).
		FilterField("user", "=", userKey).
		Limit(1)

	keys, err := s.client.GetAll(ctx, query, &stageUsers)
	if err != nil {
		return fmt.Errorf("failed to check existing StageUser: %w", err)
	}

	if len(stageUsers) == 0 {
		stageUser := StageUser{
			StageKey:  stageKey,
			UserKey:   userKey,
			ClearDate: time.Now(),
		}
		key := datastore.IncompleteKey("StageUser", nil)
		_, err = s.client.Put(ctx, key, &stageUser)
		if err != nil {
			return fmt.Errorf("failed to create StageUser: %w", err)
		}
	} else {
		stageUsers[0].ClearDate = time.Now()
		_, err = s.client.Put(ctx, keys[0], &stageUsers[0])
		if err != nil {
			return fmt.Errorf("failed to update StageUser: %w", err)
		}
	}

	return nil
}

// HasStageUser checks if a stage user relation exists
func (s *DatastoreService) HasStageUser(ctx context.Context, stageKey *datastore.Key, userKey *datastore.Key) (bool, error) {
	query := datastore.NewQuery("StageUser").
		FilterField("stage", "=", stageKey).
		FilterField("user", "=", userKey).
		Limit(1)

	var stageUsers []StageUser
	_, err := s.client.GetAll(ctx, query, &stageUsers)
	if err != nil {
		return false, fmt.Errorf("failed to check StageUser existence: %w", err)
	}

	return len(stageUsers) > 0, nil
}

// GetClearedStagesByUser gets all cleared stages for a user
func (s *DatastoreService) GetClearedStagesByUser(ctx context.Context, userKey *datastore.Key) ([]StageUser, error) {
	query := datastore.NewQuery("StageUser").
		FilterField("user", "=", userKey).
		Order("clearDate")

	var stageUsers []StageUser
	_, err := s.client.GetAll(ctx, query, &stageUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get cleared stages: %w", err)
	}

	return stageUsers, nil
}

func (s *DatastoreService) IncrementUserClearCount(ctx context.Context, userKey *datastore.Key) error {
	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
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
func (s *DatastoreService) DeleteUser(ctx context.Context, userID string) error {
	userKey := datastore.NameKey("User", "KEY"+userID, nil)
	var user User
	err := s.client.Get(ctx, userKey, &user)
	if err != nil {
		if err == datastore.ErrNoSuchEntity {
			return fmt.Errorf("user not found: %s", userID)
		}
		return fmt.Errorf("failed to get user: %w", err)
	}

	// TODO: Add audit log for user account deletion (required for compliance)

	_, err = s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		query := datastore.NewQuery("StageUser").FilterField("user", "=", userKey)
		keys, err := s.client.GetAll(ctx, query, &[]StageUser{})
		if err != nil {
			return fmt.Errorf("failed to get StageUser records: %w", err)
		}

		if len(keys) > 0 {
			err = tx.DeleteMulti(keys)
			if err != nil {
				return fmt.Errorf("failed to delete StageUser records: %w", err)
			}
		}

		// Anonymize creator field in KyouenPuzzle entities created by this user
		stageQuery := datastore.NewQuery("KyouenPuzzle").FilterField("creator", "=", user.ScreenName)
		var stages []KyouenPuzzle
		stageKeys, err := s.client.GetAll(ctx, stageQuery, &stages)
		if err != nil {
			return fmt.Errorf("failed to get user's stages: %w", err)
		}

		for i, stage := range stages {
			stage.Creator = "[deleted user]"
			_, err = tx.Put(stageKeys[i], &stage)
			if err != nil {
				return fmt.Errorf("failed to anonymize stage creator: %w", err)
			}
		}

		err = tx.Delete(userKey)
		if err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		return nil
	})

	return err
}

// GetRecentActivities gets recent user activities (stage completions)
func (s *DatastoreService) GetRecentActivities(ctx context.Context, limit int) ([]StageUser, error) {
	var stageUsers []StageUser
	query := datastore.NewQuery("StageUser").
		Order("-clearDate").
		Limit(limit)

	_, err := s.client.GetAll(ctx, query, &stageUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activities: %w", err)
	}

	return stageUsers, nil
}

// GetUserByKey gets a user by datastore key
func (s *DatastoreService) GetUserByKey(ctx context.Context, userKey *datastore.Key) (*User, error) {
	var user User
	err := s.client.Get(ctx, userKey, &user)
	if err != nil {
		return nil, fmt.Errorf("failed to get user by key: %w", err)
	}
	return &user, nil
}

// GetStageByKey gets a stage by datastore key
func (s *DatastoreService) GetStageByKey(ctx context.Context, stageKey *datastore.Key) (*KyouenPuzzle, error) {
	var stage KyouenPuzzle
	err := s.client.Get(ctx, stageKey, &stage)
	if err != nil {
		return nil, fmt.Errorf("failed to get stage by key: %w", err)
	}
	return &stage, nil
}

// GetStagesByKeys gets multiple stages by datastore keys.
// ErrNoSuchEntity entries are returned as zero-value KyouenPuzzle; other errors abort.
func (s *DatastoreService) GetStagesByKeys(ctx context.Context, keys []*datastore.Key) ([]KyouenPuzzle, error) {
	stages := make([]KyouenPuzzle, len(keys))
	err := s.client.GetMulti(ctx, keys, stages)
	if err == nil {
		return stages, nil
	}
	if merr, ok := err.(datastore.MultiError); ok {
		for _, e := range merr {
			if e != nil && e != datastore.ErrNoSuchEntity {
				return nil, fmt.Errorf("failed to get stages by keys: %w", err)
			}
		}
		return stages, nil
	}
	return nil, fmt.Errorf("failed to get stages by keys: %w", err)
}

// GetUsersByKeys gets multiple users by datastore keys.
// ErrNoSuchEntity entries are returned as zero-value User; other errors abort.
func (s *DatastoreService) GetUsersByKeys(ctx context.Context, keys []*datastore.Key) ([]User, error) {
	users := make([]User, len(keys))
	err := s.client.GetMulti(ctx, keys, users)
	if err == nil {
		return users, nil
	}
	if merr, ok := err.(datastore.MultiError); ok {
		for _, e := range merr {
			if e != nil && e != datastore.ErrNoSuchEntity {
				return nil, fmt.Errorf("failed to get users by keys: %w", err)
			}
		}
		return users, nil
	}
	return nil, fmt.Errorf("failed to get users by keys: %w", err)
}

// MigrateFirebaseUID migrates a user from an old Firebase UID to a new one.
// This handles the case where a Firebase Auth account was deleted and re-created,
// resulting in a new UID. The new-UID User entity must already exist in Datastore.
func (s *DatastoreService) MigrateFirebaseUID(ctx context.Context, oldUID, newUID string) (*User, error) {
	oldKey := datastore.NameKey("User", "KEY"+oldUID, nil)
	newKey := datastore.NameKey("User", "KEY"+newUID, nil)

	var migratedUser User

	_, err := s.client.RunInTransaction(ctx, func(tx *datastore.Transaction) error {
		var oldUser User
		if err := tx.Get(oldKey, &oldUser); err != nil {
			return fmt.Errorf("旧ユーザーの取得に失敗: %w", err)
		}

		var newUser User
		if err := tx.Get(newKey, &newUser); err != nil {
			return fmt.Errorf("新ユーザーの取得に失敗: %w", err)
		}

		// clearStageCount を旧ユーザーから引き継ぐ（screenName/image/twitterUid は新側を維持）
		newUser.ClearStageCount = oldUser.ClearStageCount
		migratedUser = newUser
		if _, err := tx.Put(newKey, &newUser); err != nil {
			return fmt.Errorf("新ユーザーの更新に失敗: %w", err)
		}

		migration := UserMigration{
			OldKey:      "KEY" + oldUID,
			NewKey:      "KEY" + newUID,
			TwitterUID:  oldUser.TwitterUID,
			FirebaseUID: newUID,
			MigratedAt:  time.Now(),
		}
		migrationKey := datastore.IncompleteKey("UserMigration", nil)
		if _, err := tx.Put(migrationKey, &migration); err != nil {
			return fmt.Errorf("UserMigration レコードの保存に失敗: %w", err)
		}

		if err := tx.Delete(oldKey); err != nil {
			return fmt.Errorf("旧ユーザーの削除に失敗: %w", err)
		}

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("Firebase UID 補正に失敗: %w", err)
	}

	// StageUser の UserKey を新キーに差し替える（トランザクション外: 25エンティティグループ制限のため）
	s.migrateStageUserRecords(ctx, oldKey, newKey)

	return &migratedUser, nil
}

// CountStageUsersByUserKey counts StageUser records that reference the given user key.
func (s *DatastoreService) CountStageUsersByUserKey(ctx context.Context, userKey *datastore.Key) (int, error) {
	query := datastore.NewQuery("StageUser").FilterField("user", "=", userKey).KeysOnly()
	keys, err := s.client.GetAll(ctx, query, &[]StageUser{})
	if err != nil {
		return 0, fmt.Errorf("StageUser 件数の取得に失敗: %w", err)
	}
	return len(keys), nil
}

// createRegistModel creates a RegistModel record for a given stage
func (s *DatastoreService) createRegistModel(ctx context.Context, stageKey *datastore.Key) error {
	registModel := RegistModel{
		StageInfo:  stageKey,
		RegistDate: time.Now(),
	}

	key := datastore.IncompleteKey("RegistModel", nil)
	_, err := s.client.Put(ctx, key, &registModel)
	if err != nil {
		return fmt.Errorf("failed to save RegistModel: %w", err)
	}

	return nil
}

// GetAndDeleteRegistModels fetches all pending RegistModel entries, deletes them,
// and returns the associated KyouenPuzzle keys.
// Returns nil if no entries exist (no new stages since last notification).
func (s *DatastoreService) GetAndDeleteRegistModels(ctx context.Context) ([]*datastore.Key, error) {
	var registModels []RegistModel
	query := datastore.NewQuery("RegistModel")

	keys, err := s.client.GetAll(ctx, query, &registModels)
	if err != nil {
		return nil, fmt.Errorf("failed to get RegistModels: %w", err)
	}

	if len(keys) == 0 {
		return nil, nil
	}

	if err := s.client.DeleteMulti(ctx, keys); err != nil {
		return nil, fmt.Errorf("failed to delete RegistModels: %w", err)
	}

	stageKeys := make([]*datastore.Key, len(registModels))
	for i, rm := range registModels {
		stageKeys[i] = rm.StageInfo
	}

	return stageKeys, nil
}
