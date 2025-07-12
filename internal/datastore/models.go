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
