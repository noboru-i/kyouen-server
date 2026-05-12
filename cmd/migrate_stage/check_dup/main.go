package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/datastore"
	"kyouen-server/pkg/models"
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

type DuplicateResult struct {
	MigratedStageNo   int64  `json:"migratedStageNo"`
	MigratedStage     string `json:"migratedStage"`
	DuplicateStageNo  int64  `json:"duplicateStageNo"`
	DuplicateStage    string `json:"duplicateStage"`
	MatchedVariant    string `json:"matchedVariant"`
}

func allVariants(stage models.KyouenStage) []struct {
	label string
	str   string
} {
	variants := make([]struct {
		label string
		str   string
	}, 0, 8)

	s := stage
	labels := []string{"rot0", "rot90", "rot180", "rot270"}
	for i := 0; i < 4; i++ {
		mirrored := models.NewMirroredKyouenStage(s)
		variants = append(variants, struct {
			label string
			str   string
		}{labels[i] + "_mirror", mirrored.ToString()})

		variants = append(variants, struct {
			label string
			str   string
		}{labels[i], s.ToString()})

		s = *models.NewRotatedKyouenStage(s)
	}
	return variants
}

func main() {
	backupPath := "cmd/migrate_stage/data/padded_stages_before.json"
	if len(os.Args) >= 2 {
		backupPath = os.Args[1]
	}

	// マイグレーション対象のJSONを読み込む
	f, err := os.Open(backupPath)
	if err != nil {
		log.Fatalf("バックアップファイルのオープン失敗: %v", err)
	}
	defer f.Close()

	var migratedRecords []PaddedStageRecord
	if err := json.NewDecoder(f).Decode(&migratedRecords); err != nil {
		log.Fatalf("JSONデコード失敗: %v", err)
	}
	log.Printf("マイグレーション対象: %d件\n", len(migratedRecords))

	// マイグレーション対象のStageNoセットを作成
	migratedStageNos := make(map[int64]bool, len(migratedRecords))
	for _, r := range migratedRecords {
		migratedStageNos[r.StageNo] = true
	}

	// Datastoreから全ステージを取得
	ctx := context.Background()
	client, err := datastore.NewClient(ctx, productionProjectID)
	if err != nil {
		log.Fatalf("Datastoreクライアント作成失敗: %v", err)
	}
	defer client.Close()

	log.Println("本番Datastoreからステージデータを取得中...")

	var allStages []KyouenPuzzle
	query := datastore.NewQuery("KyouenPuzzle").Order("stageNo")
	_, err = client.GetAll(ctx, query, &allStages)
	if err != nil {
		log.Fatalf("ステージ取得失敗: %v", err)
	}
	log.Printf("取得完了: %d ステージ\n", len(allStages))

	// ステージ文字列 -> StageNo の逆引きマップを構築（マイグレーション対象を除く）
	stageMap := make(map[string]int64, len(allStages))
	for _, s := range allStages {
		if !migratedStageNos[s.StageNo] {
			stageMap[s.Stage] = s.StageNo
		}
	}
	log.Printf("比較対象ステージ数（マイグレーション対象除く）: %d件\n", len(stageMap))
	fmt.Println()

	// 各マイグレーション済みステージについて重複チェック
	var duplicates []DuplicateResult

	for _, record := range migratedRecords {
		stage := models.NewKyouenStage(int(record.Size), record.TrimmedStage)
		variants := allVariants(*stage)

		for _, v := range variants {
			if dupNo, ok := stageMap[v.str]; ok {
				duplicates = append(duplicates, DuplicateResult{
					MigratedStageNo:  record.StageNo,
					MigratedStage:    record.TrimmedStage,
					DuplicateStageNo: dupNo,
					DuplicateStage:   v.str,
					MatchedVariant:   v.label,
				})
				break // 1つ見つかればそのステージの検索は終了
			}
		}
	}

	// 結果表示
	fmt.Printf("=== 重複チェック結果 ===\n")
	fmt.Printf("チェック対象: %d件\n", len(migratedRecords))
	fmt.Printf("重複あり    : %d件\n", len(duplicates))
	fmt.Printf("重複なし    : %d件\n", len(migratedRecords)-len(duplicates))
	fmt.Println()

	if len(duplicates) > 0 {
		fmt.Println("=== 重複ステージ一覧 ===")
		for _, d := range duplicates {
			fmt.Printf("[マイグレーション StageNo:%d] -> 重複 StageNo:%d (バリアント:%s)\n",
				d.MigratedStageNo, d.DuplicateStageNo, d.MatchedVariant)
			fmt.Printf("  マイグレーション: %s\n", d.MigratedStage)
			fmt.Printf("  重複先          : %s\n", d.DuplicateStage)
			fmt.Println()
		}

		// JSONでも出力
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		fmt.Println("\n=== JSON出力 ===")
		_ = enc.Encode(duplicates)
	} else {
		fmt.Println("重複するステージは見つかりませんでした。")
	}
}
