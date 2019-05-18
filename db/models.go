package db

import (
	"time"
)

type KyouenPuzzleSummary struct {
	Count    int64
	LastDate time.Time
}

type KyouenPuzzle struct {
	StageNo    int64
	Size       int64
	Stage      string
	Creator    string
	RegistDate time.Time
}
