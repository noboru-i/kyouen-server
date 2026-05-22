package stage

import (
	"context"
	"errors"
	"testing"
	"time"

	"cloud.google.com/go/datastore"
	"kyouen-server/internal/auth"
	datastoreservice "kyouen-server/internal/datastore"
)

func makeKey(kind string, id int64) *datastore.Key {
	return &datastore.Key{Kind: kind, ID: id}
}

// mockDS はテスト用の datastoreServiceIface モック実装
type mockDS struct {
	getStageByNoFn         func(context.Context, int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error)
	getOrCreateGuestUserFn func(context.Context) (*datastoreservice.User, *datastore.Key, error)
	getUserByIDFn          func(context.Context, string) (*datastoreservice.User, *datastore.Key, error)
	createStageUserFn      func(context.Context, *datastore.Key, *datastore.Key) error
	createStageUserCalls   int
}

func (m *mockDS) GetStageByNo(ctx context.Context, stageNo int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
	return m.getStageByNoFn(ctx, stageNo)
}
func (m *mockDS) GetOrCreateGuestUser(ctx context.Context) (*datastoreservice.User, *datastore.Key, error) {
	return m.getOrCreateGuestUserFn(ctx)
}
func (m *mockDS) GetUserByID(ctx context.Context, userID string) (*datastoreservice.User, *datastore.Key, error) {
	return m.getUserByIDFn(ctx, userID)
}
func (m *mockDS) CreateStageUser(ctx context.Context, stageKey *datastore.Key, userKey *datastore.Key) error {
	m.createStageUserCalls++
	if m.createStageUserFn != nil {
		return m.createStageUserFn(ctx, stageKey, userKey)
	}
	return nil
}

// 使わないメソッドはpanicで実装（誤呼び出しをテストで検出する）
func (m *mockDS) GetStages(_ context.Context, _, _ int) ([]datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
	panic("not implemented")
}
func (m *mockDS) GetClearedStageKeyIDs(_ context.Context, _ *datastore.Key) (map[int64]time.Time, error) {
	panic("not implemented")
}
func (m *mockDS) CheckStageExists(_ context.Context, _ string) (bool, error) {
	panic("not implemented")
}
func (m *mockDS) CreateStage(_ context.Context, _ datastoreservice.KyouenPuzzle) (*datastoreservice.KyouenPuzzle, error) {
	panic("not implemented")
}
func (m *mockDS) HasStageUser(_ context.Context, _ *datastore.Key, _ *datastore.Key) (bool, error) {
	panic("not implemented")
}
func (m *mockDS) GetClearedStagesByUser(_ context.Context, _ *datastore.Key) ([]datastoreservice.StageUser, error) {
	panic("not implemented")
}
func (m *mockDS) GetStagesByKeys(_ context.Context, _ []*datastore.Key) ([]datastoreservice.KyouenPuzzle, error) {
	panic("not implemented")
}
func (m *mockDS) GetRecentActivities(_ context.Context, _ int) ([]datastoreservice.StageUser, error) {
	panic("not implemented")
}
func (m *mockDS) GetUsersByKeys(_ context.Context, _ []*datastore.Key) ([]datastoreservice.User, error) {
	panic("not implemented")
}
func (m *mockDS) DeleteUser(_ context.Context, _ string) error {
	panic("not implemented")
}

// 4x4グリッドの行0に白石4つ（直線共円）
const (
	validStageInput = "2222000000000000" // ユーザー入力（白石が直線共円）
	validStageDB    = "1111000000000000" // DBに保存された黒石パターン
)

// TestClearStage_StageErrorPriority: ステージ不在とユーザー不在が同時に起きると ErrStageNotFound が優先される
func TestClearStage_StageErrorPriority(t *testing.T) {
	ds := &mockDS{
		getStageByNoFn: func(_ context.Context, _ int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
			return nil, nil, errors.New("not found")
		},
		getOrCreateGuestUserFn: func(_ context.Context) (*datastoreservice.User, *datastore.Key, error) {
			return nil, nil, errors.New("not found")
		},
	}
	svc := &Service{datastoreService: ds}

	_, err := svc.ClearStage(context.Background(), 1, validStageInput, auth.GuestUID)
	if !errors.Is(err, ErrStageNotFound) {
		t.Errorf("expected ErrStageNotFound, got %v", err)
	}
}

// TestClearStage_StageMismatchPriority: ステージ不一致とユーザー不在が同時に起きると ErrStageMismatch が優先される
func TestClearStage_StageMismatchPriority(t *testing.T) {
	stageKey := makeKey("Stage", 1)
	ds := &mockDS{
		getStageByNoFn: func(_ context.Context, _ int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
			return &datastoreservice.KyouenPuzzle{Stage: "0000000000000000"}, []*datastore.Key{stageKey}, nil
		},
		getOrCreateGuestUserFn: func(_ context.Context) (*datastoreservice.User, *datastore.Key, error) {
			return nil, nil, errors.New("not found")
		},
	}
	svc := &Service{datastoreService: ds}

	_, err := svc.ClearStage(context.Background(), 1, validStageInput, auth.GuestUID)
	if !errors.Is(err, ErrStageMismatch) {
		t.Errorf("expected ErrStageMismatch, got %v", err)
	}
}

// TestClearStage_Success_CreateStageUserCalled: 正常ケースで CreateStageUser が1回呼ばれる
func TestClearStage_Success_CreateStageUserCalled(t *testing.T) {
	stageKey := makeKey("Stage", 1)
	userKey := makeKey("User", 0)
	guestUser := &datastoreservice.User{UserID: auth.GuestUID, ScreenName: "Guest"}

	ds := &mockDS{
		getStageByNoFn: func(_ context.Context, _ int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
			return &datastoreservice.KyouenPuzzle{Stage: validStageDB}, []*datastore.Key{stageKey}, nil
		},
		getOrCreateGuestUserFn: func(_ context.Context) (*datastoreservice.User, *datastore.Key, error) {
			return guestUser, userKey, nil
		},
	}
	svc := &Service{datastoreService: ds}

	user, err := svc.ClearStage(context.Background(), 1, validStageInput, auth.GuestUID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.UserID != auth.GuestUID {
		t.Errorf("expected guest user, got userID=%q", user.UserID)
	}
	if ds.createStageUserCalls != 1 {
		t.Errorf("expected CreateStageUser called once, got %d", ds.createStageUserCalls)
	}
}

// TestClearStage_StageNotFound_NoCreateStageUser: ステージ不在のとき CreateStageUser は呼ばれない
func TestClearStage_StageNotFound_NoCreateStageUser(t *testing.T) {
	userKey := makeKey("User", 0)
	guestUser := &datastoreservice.User{UserID: auth.GuestUID, ScreenName: "Guest"}

	ds := &mockDS{
		getStageByNoFn: func(_ context.Context, _ int) (*datastoreservice.KyouenPuzzle, []*datastore.Key, error) {
			return nil, nil, errors.New("not found")
		},
		getOrCreateGuestUserFn: func(_ context.Context) (*datastoreservice.User, *datastore.Key, error) {
			return guestUser, userKey, nil
		},
	}
	svc := &Service{datastoreService: ds}

	_, err := svc.ClearStage(context.Background(), 1, validStageInput, auth.GuestUID)
	if !errors.Is(err, ErrStageNotFound) {
		t.Errorf("expected ErrStageNotFound, got %v", err)
	}
	if ds.createStageUserCalls != 0 {
		t.Errorf("expected CreateStageUser not called, got %d calls", ds.createStageUserCalls)
	}
}

func TestUniqueKeys_Deduplication(t *testing.T) {
	k1 := makeKey("User", 1)
	k2 := makeKey("User", 2)
	keys := []*datastore.Key{k1, k2, k1, k2, k1}

	unique, idx := uniqueKeys(keys)

	if len(unique) != 2 {
		t.Errorf("Expected 2 unique keys, got %d", len(unique))
	}
	if len(idx) != 2 {
		t.Errorf("Expected idx length 2, got %d", len(idx))
	}
	if idx[k1.String()] != 0 {
		t.Errorf("Expected k1 at index 0, got %d", idx[k1.String()])
	}
	if idx[k2.String()] != 1 {
		t.Errorf("Expected k2 at index 1, got %d", idx[k2.String()])
	}
}

func TestUniqueKeys_Empty(t *testing.T) {
	unique, idx := uniqueKeys([]*datastore.Key{})

	if len(unique) != 0 {
		t.Errorf("Expected 0 unique keys, got %d", len(unique))
	}
	if len(idx) != 0 {
		t.Errorf("Expected empty idx, got %d", len(idx))
	}
}

func TestUniqueKeys_OrderPreserved(t *testing.T) {
	k1 := makeKey("Stage", 10)
	k2 := makeKey("Stage", 20)
	k3 := makeKey("Stage", 30)
	keys := []*datastore.Key{k3, k1, k2, k3, k1}

	unique, _ := uniqueKeys(keys)

	if len(unique) != 3 {
		t.Errorf("Expected 3 unique keys, got %d", len(unique))
	}
	if unique[0].ID != 30 || unique[1].ID != 10 || unique[2].ID != 20 {
		t.Errorf("Expected order [k3, k1, k2], got [%d, %d, %d]", unique[0].ID, unique[1].ID, unique[2].ID)
	}
}
