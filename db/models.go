package db

import (
	"time"
)

type KyouenPuzzleSummary struct {
	Count    int64
	LastDate time.Time
}
