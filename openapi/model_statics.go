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

type Statics struct {

	Count int64 `json:"count"`

	// date in UTC
	LastUpdatedAt time.Time `json:"last_updated_at"`
}
