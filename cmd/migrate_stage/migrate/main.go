package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

const productionProjectID = "my-android-server"

type KyouenPuzzle struct {
	StageNo    int64     `datastore:"stageNo"`
	Size       int64     `datastore:"size"`
	Stage      string    `datastore:"stage"`
	Creator    string    `datastore:"creator"`
	RegistDate time.Time `datastore:"registDate"`
}

type PaddedStageRecord struct {
	StageNo       int64     `json:"stageNo"`
	Size          int64     `json:"size"`
	OriginalStage string    `json:"originalStage"`
	TrimmedStage  string    `json:"trimmedStage"`
	Creator       string    `json:"creator"`
	RegistDate    time.Time `json:"registDate"`
}

func main() {
	dryRun := true
	if len(os.Args) >= 2 && os.Args[1] == "--apply" {
		dryRun = false
	}

	if dryRun {
		log.Println("[DRY-RUN] 実際のデータは変更しません。--apply を指定すると実行されます。")
	} else {
		log.Println("[APPLY] 本番Datastoreのデータを更新します。")
	}

	backupPath := "cmd/migrate_stage/data/padded_stages_before.json"

	ctx := context.Background()
	client, err := datastore.NewClient(ctx, productionProjectID)
	if err != nil {
		log.Fatalf("Datastoreクライアント作成失敗: %v", err)
	}
	defer client.Close()

	log.Println("本番Datastoreからステージデータを取得中...")

	var allStages []KyouenPuzzle
	var allKeys []*datastore.Key
	query := datastore.NewQuery("KyouenPuzzle").Order("stageNo")
	allKeys, err = client.GetAll(ctx, query, &allStages)
	if err != nil {
		log.Fatalf("ステージ取得失敗: %v", err)
	}

	log.Printf("取得完了: %d ステージ\n", len(allStages))

	// 補正対象の抽出
	type target struct {
		key   *datastore.Key
		stage KyouenPuzzle
	}
	var targets []target
	var records []PaddedStageRecord

	for i, s := range allStages {
		if s.Size <= 0 {
			continue
		}
		expectedLen := int(s.Size * s.Size)
		if len(s.Stage) <= expectedLen {
			continue
		}
		padding := s.Stage[expectedLen:]
		if strings.Count(padding, "0") != len(padding) {
			continue
		}
		// ゼロパディングパターン
		trimmed := s.Stage[:expectedLen]
		records = append(records, PaddedStageRecord{
			StageNo:       s.StageNo,
			Size:          s.Size,
			OriginalStage: s.Stage,
			TrimmedStage:  trimmed,
			Creator:       s.Creator,
			RegistDate:    s.RegistDate,
		})
		fixed := s
		fixed.Stage = trimmed
		targets = append(targets, target{key: allKeys[i], stage: fixed})
	}

	log.Printf("補正対象: %d件\n", len(targets))

	if len(targets) == 0 {
		log.Println("補正対象がありません。")
		return
	}

	// 補正前データをJSONに保存（--apply 時のみ）
	if !dryRun {
		if err := saveBackup(backupPath, records); err != nil {
			log.Fatalf("バックアップ保存失敗: %v", err)
		}
		log.Printf("補正前データを %s に保存しました\n", backupPath)
	}

	// 補正実行
	const batchSize = 500
	successCount := 0
	errorCount := 0

	for i := 0; i < len(targets); i += batchSize {
		end := i + batchSize
		if end > len(targets) {
			end = len(targets)
		}
		batch := targets[i:end]

		if dryRun {
			for _, t := range batch {
				fmt.Printf("[DRY-RUN] StageNo=%d: %q -> %q\n",
					t.stage.StageNo, records[i].OriginalStage, t.stage.Stage)
			}
			successCount += len(batch)
			continue
		}

		keys := make([]*datastore.Key, len(batch))
		stages := make([]*KyouenPuzzle, len(batch))
		for j, t := range batch {
			keys[j] = t.key
			s := t.stage
			stages[j] = &s
		}

		if _, err := client.PutMulti(ctx, keys, stages); err != nil {
			log.Printf("バッチ %d-%d の更新失敗: %v\n", i, end-1, err)
			errorCount += len(batch)
			continue
		}
		successCount += len(batch)
		log.Printf("バッチ %d-%d 完了 (%d/%d)\n", i, end-1, successCount, len(targets))
	}

	fmt.Printf("\n=== 補正結果 ===\n")
	fmt.Printf("成功: %d件\n", successCount)
	fmt.Printf("失敗: %d件\n", errorCount)
	if dryRun {
		fmt.Println("\n[DRY-RUN] 上記は確認のみです。実行するには --apply を指定してください。")
	}
}

func saveBackup(path string, records []PaddedStageRecord) error {
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("ファイル作成失敗: %w", err)
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(records)
}
