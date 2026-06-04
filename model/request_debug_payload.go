package model

import (
	"github.com/QuantumNous/new-api/common"
)

// RequestDebugPayload stores sampled raw request/response bodies for L2 deep diagnostics.
// Populated asynchronously when L2 debug sampling is enabled.
type RequestDebugPayload struct {
	Id             int    `json:"id" gorm:"primaryKey"`
	RequestId      string `json:"request_id" gorm:"type:varchar(64);uniqueIndex;default:''"`
	LogId          int    `json:"log_id" gorm:"index;default:0"`
	UserId         int    `json:"user_id" gorm:"index"`
	ModelName      string `json:"model_name" gorm:"size:128;index;default:''"`
	ChannelId      int    `json:"channel_id" gorm:"index;default:0"`
	TokenId        int    `json:"token_id" gorm:"index;default:0"`
	Group          string `json:"group" gorm:"index;default:''"`
	CreatedAt      int64  `json:"created_at" gorm:"index"`
	RequestBody    string `json:"request_body" gorm:"type:text"`
	IsStream       bool   `json:"is_stream"`
	RelayFormat    string `json:"relay_format" gorm:"size:32;default:''"`
	Parsed         bool   `json:"parsed" gorm:"index;default:false"`
	SamplingReason string `json:"sampling_reason" gorm:"size:128;default:''"`
}

// RequestContextPart represents a parsed segment of a debug request body,
// enabling L2 deep diagnostics (context structure, prefix reuse, cache friendliness).
type RequestContextPart struct {
	Id              int    `json:"id" gorm:"primaryKey"`
	RequestId       string `json:"request_id" gorm:"type:varchar(64);index;default:''"`
	PartType        string `json:"part_type" gorm:"size:32;index;default:''"` // system, history, tool, file, memory, user
	PartName        string `json:"part_name" gorm:"size:128;default:''"`
	ContentHash     string `json:"content_hash" gorm:"size:64;index;default:''"` // SHA-256 prefix
	TokenCount      int    `json:"token_count" gorm:"default:0"`
	Position        int    `json:"position" gorm:"default:0"`
	IsPrefix        bool   `json:"is_prefix" gorm:"default:false"`
	IsRepeated      bool   `json:"is_repeated" gorm:"default:false"`
	IsStable        bool   `json:"is_stable" gorm:"default:false"`
	IsCacheFriendly bool   `json:"is_cache_friendly" gorm:"default:false"`
	CreatedAt       int64  `json:"created_at" gorm:"index"`
}

// CreateRequestDebugPayload creates a debug payload record.
// Ignores duplicate request_id (unique constraint).
func CreateRequestDebugPayload(payload *RequestDebugPayload) error {
	if payload.CreatedAt == 0 {
		payload.CreatedAt = common.GetTimestamp()
	}
	// Use Save to handle unique constraint gracefully (upsert on duplicate)
	result := LOG_DB.Where("request_id = ?", payload.RequestId).
		Assign(map[string]interface{}{
			"request_body":    payload.RequestBody,
			"parsed":          payload.Parsed,
			"sampling_reason": payload.SamplingReason,
		}).
		FirstOrCreate(payload)
	return result.Error
}

// CreateRequestContextParts bulk-inserts parsed context parts for a given request_id.
func CreateRequestContextParts(requestId string, parts []RequestContextPart) error {
	if len(parts) == 0 {
		return nil
	}
	ts := common.GetTimestamp()
	for i := range parts {
		parts[i].RequestId = requestId
		if parts[i].CreatedAt == 0 {
			parts[i].CreatedAt = ts
		}
	}
	return LOG_DB.Create(&parts).Error
}

// GetRequestContextPartsByRequestId returns all context parts for a given request.
func GetRequestContextPartsByRequestId(requestId string) ([]RequestContextPart, error) {
	var parts []RequestContextPart
	err := LOG_DB.Where("request_id = ?", requestId).
		Order("position ASC").
		Find(&parts).Error
	return parts, err
}

// MarkDebugPayloadParsed marks a debug payload as having been parsed into context parts.
func MarkDebugPayloadParsed(requestId string) error {
	return LOG_DB.Model(&RequestDebugPayload{}).
		Where("request_id = ?", requestId).
		Update("parsed", true).Error
}

// DeleteOldDebugPayloads removes debug payloads older than the given timestamp,
// and cascades deletion of associated context parts.
func DeleteOldDebugPayloads(targetTimestamp int64, limit int) (int64, error) {
	var total int64
	for {
		// First, get request_ids to delete
		var ids []string
		result := LOG_DB.Model(&RequestDebugPayload{}).
			Where("created_at < ?", targetTimestamp).
			Limit(limit).
			Pluck("request_id", &ids)
		if result.Error != nil {
			return total, result.Error
		}
		if len(ids) == 0 {
			break
		}

		// Delete context parts
		LOG_DB.Where("request_id IN ?", ids).Delete(&RequestContextPart{})

		// Delete payloads
		delResult := LOG_DB.Where("request_id IN ?", ids).Delete(&RequestDebugPayload{})
		if delResult.Error != nil {
			return total, delResult.Error
		}

		total += delResult.RowsAffected
		if delResult.RowsAffected < int64(limit) {
			break
		}
	}
	return total, nil
}
