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
	StageNo      int64     `json:"stageNo"`
	Size         int64     `json:"size"`
	OriginalStage string   `json:"originalStage"`
	TrimmedStage  string   `json:"trimmedStage"`
	Creator      string    `json:"creator"`
	RegistDate   time.Time `json:"registDate"`
}

type ValidationResult struct {
	StageNo           int64
	Size              int64
	Stage             string
	Creator           string
	RegistDate        time.Time
	PaddedZero        bool
	TrimmedStage      string
	TrimmedHasKyouen  bool
	TrimmedStoneCount int
	Errors            []string
}

func main() {
	outputPath := ""
	if len(os.Args) >= 2 {
		outputPath = os.Args[1]
	}

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
	fmt.Println()

	var (
		validCount    int
		paddedValid   []ValidationResult
		paddedInvalid []ValidationResult
		trulyInvalid  []ValidationResult
	)

	for _, stage := range allStages {
		result := validateStage(stage)
		if len(result.Errors) == 0 {
			validCount++
			continue
		}
		if result.PaddedZero {
			if result.TrimmedHasKyouen {
				paddedValid = append(paddedValid, result)
			} else {
				paddedInvalid = append(paddedInvalid, result)
			}
		} else {
			trulyInvalid = append(trulyInvalid, result)
		}
	}

	fmt.Printf("=== 検証結果サマリー ===\n")
	fmt.Printf("総ステージ数           : %d\n", len(allStages))
	fmt.Printf("正常ステージ数         : %d\n", validCount)
	fmt.Printf("ゼロパディング（有効）  : %d\n", len(paddedValid))
	fmt.Printf("ゼロパディング（無効）  : %d\n", len(paddedInvalid))
	fmt.Printf("その他の不正ステージ数 : %d\n", len(trulyInvalid))
	fmt.Println()

	if len(paddedValid) > 0 {
		fmt.Printf("=== ゼロパディングあり（先頭 size*size 文字は有効な共円を持つ） ===\n")
		fmt.Printf("総数: %d件\n", len(paddedValid))
		for i, r := range paddedValid {
			if i >= 10 {
				fmt.Printf("  ... 他%d件省略\n", len(paddedValid)-10)
				break
			}
			fmt.Printf("  [StageNo:%d] size=%d creator=%q trimmedStone=%d\n",
				r.StageNo, r.Size, r.Creator, r.TrimmedStoneCount)
			fmt.Printf("    trimmed: %s\n", r.TrimmedStage)
		}
		fmt.Println()
	}

	if len(paddedInvalid) > 0 {
		fmt.Printf("=== ゼロパディングあり（先頭部分も共円なし） ===\n")
		for _, r := range paddedInvalid {
			fmt.Printf("[StageNo: %d] (size=%d, creator=%q)\n", r.StageNo, r.Size, r.Creator)
			fmt.Printf("  TrimmedStage: %s (stones=%d)\n", r.TrimmedStage, r.TrimmedStoneCount)
			fmt.Printf("  問題: %s\n\n", strings.Join(r.Errors, ", "))
		}
	}

	if len(trulyInvalid) > 0 {
		fmt.Printf("=== その他の不正ステージ ===\n")
		for _, r := range trulyInvalid {
			fmt.Printf("[StageNo: %d] (size=%d, creator=%q)\n", r.StageNo, r.Size, r.Creator)
			fmt.Printf("  Stage: %s\n", r.Stage)
			fmt.Printf("  問題: %s\n\n", strings.Join(r.Errors, ", "))
		}
	}

	if len(paddedValid) == 0 && len(paddedInvalid) == 0 && len(trulyInvalid) == 0 {
		fmt.Println("不正なステージは見つかりませんでした。")
	}

	// 補正対象データをJSONに出力
	if outputPath != "" {
		records := make([]PaddedStageRecord, 0, len(paddedValid)+len(paddedInvalid))
		for _, r := range paddedValid {
			records = append(records, PaddedStageRecord{
				StageNo:       r.StageNo,
				Size:          r.Size,
				OriginalStage: r.Stage,
				TrimmedStage:  r.TrimmedStage,
				Creator:       r.Creator,
				RegistDate:    r.RegistDate,
			})
		}
		for _, r := range paddedInvalid {
			records = append(records, PaddedStageRecord{
				StageNo:       r.StageNo,
				Size:          r.Size,
				OriginalStage: r.Stage,
				TrimmedStage:  r.TrimmedStage,
				Creator:       r.Creator,
				RegistDate:    r.RegistDate,
			})
		}

		f, err := os.Create(outputPath)
		if err != nil {
			log.Fatalf("出力ファイル作成失敗: %v", err)
		}
		defer f.Close()

		enc := json.NewEncoder(f)
		enc.SetIndent("", "  ")
		if err := enc.Encode(records); err != nil {
			log.Fatalf("JSON書き込み失敗: %v", err)
		}
		log.Printf("補正前データを %s に保存しました (%d件)\n", outputPath, len(records))
	}
}

func validateStage(stage KyouenPuzzle) ValidationResult {
	result := ValidationResult{
		StageNo:    stage.StageNo,
		Size:       stage.Size,
		Stage:      stage.Stage,
		Creator:    stage.Creator,
		RegistDate: stage.RegistDate,
	}

	if stage.Size <= 0 {
		result.Errors = append(result.Errors, fmt.Sprintf("sizeが不正: %d", stage.Size))
		return result
	}

	expectedLen := int(stage.Size * stage.Size)
	actualLen := len(stage.Stage)

	if actualLen > expectedLen {
		padding := stage.Stage[expectedLen:]
		if strings.Count(padding, "0") == len(padding) {
			result.PaddedZero = true
			trimmed := stage.Stage[:expectedLen]
			result.TrimmedStage = trimmed

			if errs := validateStageContent(trimmed, stage.Size); len(errs) > 0 {
				result.Errors = append(result.Errors, errs...)
				result.TrimmedStoneCount = strings.Count(trimmed, "1")
			} else {
				result.TrimmedStoneCount = strings.Count(trimmed, "1")
				kyouenStage := models.NewKyouenStage(int(stage.Size), trimmed)
				result.TrimmedHasKyouen = kyouenStage.HasKyouen() != nil
				if !result.TrimmedHasKyouen {
					result.Errors = append(result.Errors, "先頭部分に有効な共円が存在しない")
				} else {
					result.Errors = append(result.Errors, fmt.Sprintf("ゼロパディングあり(len=%d)", actualLen))
				}
			}
			return result
		}

		result.Errors = append(result.Errors, fmt.Sprintf("stage文字列長が不正: 期待=%d 実際=%d（末尾に非ゼロあり）", expectedLen, actualLen))
		return result
	}

	if actualLen < expectedLen {
		result.Errors = append(result.Errors, fmt.Sprintf("stage文字列長が不正: 期待=%d 実際=%d", expectedLen, actualLen))
		return result
	}

	if errs := validateStageContent(stage.Stage, stage.Size); len(errs) > 0 {
		result.Errors = append(result.Errors, errs...)
	}

	return result
}

func validateStageContent(stageStr string, size int64) []string {
	var errs []string

	for i, c := range stageStr {
		if c != '0' && c != '1' && c != '2' {
			errs = append(errs, fmt.Sprintf("不正な文字 %q がインデックス %d に存在", c, i))
		}
	}
	if len(errs) > 0 {
		return errs
	}

	stoneCount := strings.Count(stageStr, "1")
	if stoneCount < 5 {
		errs = append(errs, fmt.Sprintf("石の数が不足: %d個 (最低5個必要)", stoneCount))
		return errs
	}

	kyouenStage := models.NewKyouenStage(int(size), stageStr)
	if kyouenStage.HasKyouen() == nil {
		errs = append(errs, "有効な共円が存在しない")
	}

	return errs
}
