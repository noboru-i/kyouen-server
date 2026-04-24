package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"kyouen-server/internal/datastore"

	gcdatastore "cloud.google.com/go/datastore"
)

func main() {
	oldUID := flag.String("old-uid", "", "補正前の Firebase UID（必須）")
	newUID := flag.String("new-uid", "", "補正後の Firebase UID（必須）")
	dryRun := flag.Bool("dry-run", false, "true にすると Datastore への書き込みを行わず検証結果のみ表示")
	flag.Parse()

	if *oldUID == "" || *newUID == "" {
		fmt.Fprintln(os.Stderr, "使用方法: migrate_user -old-uid=<旧UID> -new-uid=<新UID> [-dry-run]")
		flag.PrintDefaults()
		os.Exit(1)
	}

	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	if projectID == "" {
		projectID = "my-android-server"
		log.Printf("GOOGLE_CLOUD_PROJECT が未設定のためデフォルトを使用: %s", projectID)
	}

	svc, err := datastore.NewDatastoreService(projectID)
	if err != nil {
		log.Fatalf("Datastore 接続に失敗: %v", err)
	}
	defer svc.Close()

	fmt.Printf("接続先プロジェクト: %s\n", projectID)
	fmt.Printf("旧 UID: %s\n", *oldUID)
	fmt.Printf("新 UID: %s\n", *newUID)
	if *dryRun {
		fmt.Println("モード: dry-run（書き込みは行いません）")
	} else {
		fmt.Println("モード: 実行（Datastore に書き込みます）")
	}
	fmt.Println("---")

	// 旧 User を取得
	oldUser, _, err := svc.GetUserByID(*oldUID)
	if err != nil {
		log.Fatalf("旧ユーザーが見つかりません（%s）: %v", *oldUID, err)
	}
	fmt.Printf("旧ユーザー: screenName=%s, twitterUid=%s, clearStageCount=%d\n",
		oldUser.ScreenName, oldUser.TwitterUID, oldUser.ClearStageCount)

	// 新 User を取得
	newUser, _, err := svc.GetUserByID(*newUID)
	if err != nil {
		log.Fatalf("新ユーザーが見つかりません（%s）: %v\n補正前に新 UID でログインしてください。", *newUID, err)
	}
	fmt.Printf("新ユーザー: screenName=%s, twitterUid=%s, clearStageCount=%d\n",
		newUser.ScreenName, newUser.TwitterUID, newUser.ClearStageCount)

	// twitterUid の一致チェック（誤補正防止）
	if oldUser.TwitterUID != newUser.TwitterUID {
		log.Fatalf(
			"旧ユーザーと新ユーザーの twitterUid が一致しません（旧: %s / 新: %s）。同一人物であることを確認してください。",
			oldUser.TwitterUID, newUser.TwitterUID,
		)
	}
	fmt.Printf("twitterUid 一致確認: OK (%s)\n", oldUser.TwitterUID)

	// StageUser 件数確認
	oldKey := gcdatastore.NameKey("User", "KEY"+*oldUID, nil)
	newKey := gcdatastore.NameKey("User", "KEY"+*newUID, nil)

	oldStageUserCount, err := svc.CountStageUsersByUserKey(oldKey)
	if err != nil {
		log.Fatalf("旧ユーザーの StageUser 件数取得に失敗: %v", err)
	}
	fmt.Printf("旧ユーザー StageUser 件数: %d 件\n", oldStageUserCount)

	newStageUserCount, err := svc.CountStageUsersByUserKey(newKey)
	if err != nil {
		log.Fatalf("新ユーザーの StageUser 件数取得に失敗: %v", err)
	}
	fmt.Printf("新ユーザー StageUser 件数: %d 件\n", newStageUserCount)

	if newStageUserCount > 0 {
		log.Fatalf(
			"新ユーザー側に既に StageUser が %d 件存在します。データ損失を防ぐため中断しました。",
			newStageUserCount,
		)
	}

	fmt.Println("---")
	fmt.Printf("補正内容:\n")
	fmt.Printf("  旧 User エンティティ (KEY%s) を削除\n", *oldUID)
	fmt.Printf("  新 User エンティティ (KEY%s) の clearStageCount を %d に更新\n", *newUID, oldUser.ClearStageCount)
	fmt.Printf("  StageUser %d 件の UserKey を新 UID に差し替え\n", oldStageUserCount)
	fmt.Printf("  UserMigration レコードを 1 件追加\n")

	if *dryRun {
		fmt.Println("---")
		fmt.Println("dry-run 完了。上記の変更は行われていません。")
		return
	}

	fmt.Println("---")
	fmt.Println("補正を実行します...")

	migratedUser, err := svc.MigrateFirebaseUID(*oldUID, *newUID)
	if err != nil {
		log.Fatalf("補正に失敗しました: %v", err)
	}

	fmt.Printf("補正完了: screenName=%s, clearStageCount=%d\n",
		migratedUser.ScreenName, migratedUser.ClearStageCount)

	// 事後確認
	afterOldCount, _ := svc.CountStageUsersByUserKey(oldKey)
	afterNewCount, _ := svc.CountStageUsersByUserKey(newKey)
	fmt.Println("---")
	fmt.Println("事後確認:")
	fmt.Printf("  旧ユーザー StageUser 件数: %d 件（0 になっているはずです）\n", afterOldCount)
	fmt.Printf("  新ユーザー StageUser 件数: %d 件（補正前の旧ユーザー件数と一致するはずです）\n", afterNewCount)

	if afterOldCount != 0 {
		fmt.Println("  警告: 旧ユーザー側に StageUser が残っています。再実行が必要な可能性があります。")
	}
	if afterNewCount != oldStageUserCount {
		fmt.Printf("  警告: 新ユーザー StageUser 件数 (%d) が期待値 (%d) と異なります。\n",
			afterNewCount, oldStageUserCount)
	}
}
