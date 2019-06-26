package db

import (
	"time"
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
