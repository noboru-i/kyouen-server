package db

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
	UserID          string `datastore:"userId"`
	ScreenName      string `datastore:"screenName"`
	Image           string `datastore:"image"`
	ClearStageCount int64  `datastore:"clearStageCount"`

	// TODO remove later (Don't use it now.)
	AccessToken  string `datastore:"accessToken"`
	AccessSecret string `datastore:"accessSecret"`
	APIToken     string `datastore:"apiToken"`
}

type StageUser struct {
	StageKey  *datastore.Key `datastore:"stage"`
	UserKey   *datastore.Key `datastore:"user"`
	ClearDate time.Time      `datastore:"clearDate"`
}
