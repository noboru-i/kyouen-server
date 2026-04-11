package datastore

import (
	"time"

	"cloud.google.com/go/datastore"
)

type KyouenPuzzleSummary struct {
	Count    int64     `datastore:"count"`
	LastDate time.Time `datastore:"lastDate"`
}

type KyouenPuzzle struct {
	StageNo    int64     `datastore:"stageNo"`
	Size       int64     `datastore:"size"`
	Stage      string    `datastore:"stage"`
	Creator    string    `datastore:"creator"`
	RegistDate time.Time `datastore:"registDate"`
}

type User struct {
	UserID          string `datastore:"userId"`          // Firebase UID
	ScreenName      string `datastore:"screenName"`      // Twitter screen name
	Image           string `datastore:"image"`           // Twitter profile image
	ClearStageCount int64  `datastore:"clearStageCount"` // Number of cleared stages
	TwitterUID      string `datastore:"twitterUid"`      // Twitter User ID (for reference)

	// TODO remove later (Legacy fields - not used in Firebase auth)
	AccessToken  string `datastore:"accessToken"`
	AccessSecret string `datastore:"accessSecret"`
	APIToken     string `datastore:"apiToken"`
}

type StageUser struct {
	StageKey  *datastore.Key `datastore:"stage"`
	UserKey   *datastore.Key `datastore:"user"`
	ClearDate time.Time      `datastore:"clearDate"`
}

type RegistModel struct {
	StageInfo  *datastore.Key `datastore:"stageInfo"`
	RegistDate time.Time      `datastore:"registDate"`
}

type UserMigration struct {
	OldKey      string    `datastore:"oldKey"`      // 旧キー名 (例: "KEY12345")
	NewKey      string    `datastore:"newKey"`      // 新キー名 (例: "KEYabc123def")
	TwitterUID  string    `datastore:"twitterUid"`  // Twitter UID
	FirebaseUID string    `datastore:"firebaseUid"` // Firebase UID
	MigratedAt  time.Time `datastore:"migratedAt"`  // マイグレーション実行日時
}
