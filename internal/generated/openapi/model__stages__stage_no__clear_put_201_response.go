// Code generated by OpenAPI Generator (https://openapi-generator.tech); DO NOT EDIT.

/*
 * 共円パズルゲーム API
 *
 * 共円パズルゲーム用REST APIサーバーです。  共円は、グリッド上に石を配置して、ちょうど4つの石で円や直線を形成する 知的パズルゲームです。このAPIは、ステージ管理、ユーザー認証、 ゲーム進行の追跡機能を提供します。  **アーキテクチャ:** - プラットフォーム: Cloud Run + DatastoreモードFirestore - フレームワーク: Gin (Go) - 認証: Twitter OAuth + Firebase  **ゲームルール:** - グリッド上に石を配置 - ちょうど4つの石で共円（円または直線）を形成 - パズル設定を解いてステージをクリア 
 *
 * API version: 2.0.0
 */

package openapi


import (
	"time"
)



type StagesStageNoClearPut201Response struct {

	StageNo int64 `json:"stage_no,omitempty"`

	// ステージがクリアされたタイムスタンプ (UTC)
	ClearDate time.Time `json:"clear_date,omitempty"`
}

// AssertStagesStageNoClearPut201ResponseRequired checks if the required fields are not zero-ed
func AssertStagesStageNoClearPut201ResponseRequired(obj StagesStageNoClearPut201Response) error {
	return nil
}

// AssertStagesStageNoClearPut201ResponseConstraints checks if the values respects the defined constraints
func AssertStagesStageNoClearPut201ResponseConstraints(obj StagesStageNoClearPut201Response) error {
	return nil
}
