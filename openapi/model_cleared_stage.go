/*
 * Kyouen API
 *
 * Kyouen server's API.
 *
 * API version: 0.0.1
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

import (
	"time"
)

type ClearedStage struct {

	StageNo int64 `json:"stageNo"`

	// date in UTC
	ClearDate time.Time `json:"clearDate"`
}
