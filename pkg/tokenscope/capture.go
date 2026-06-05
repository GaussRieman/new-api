package tokenscope

import (
	"encoding/json"
	"fmt"
	"sync/atomic"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/pkg/tokenscope/parser"
	"github.com/QuantumNous/new-api/setting/tokenscope_setting"
	relaycommon "github.com/QuantumNous/new-api/relay/common"

	"github.com/bytedance/gopkg/util/gopool"
	"github.com/gin-gonic/gin"
)

const (
	sampleEvery  = 10  // guaranteed sample every N requests
	maxSamples   = 100 // max debug records to keep
)

var sampleCounter = uint64(0) // atomic counter for round-robin sampling

// shouldSampleForL2 decides whether to sample this request using a hybrid strategy:
// 1. Round-robin: every `sampleEvery` requests is guaranteed to be sampled
// 2. Random: deterministic hash-based sampling at the configured rate
func shouldSampleForL2(requestId string, rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	// Round-robin: guarantee 1 sample every N requests
	counter := atomic.AddUint64(&sampleCounter, 1)
	if counter%uint64(sampleEvery) == 0 {
		return true
	}
	// Random: hash-based sampling
	return shouldSample(requestId, rate)
}

// MaybeCaptureRequestBody stores a pre-captured request body for L2 diagnostics.
// This is called from the goroutine in text_quota.go with body bytes that were
// captured while the gin context was still alive.
func MaybeCaptureRequestBody(info *relaycommon.RelayInfo, body []byte) {
	setting := tokenscope_setting.GetTokenScopeSetting()
	if !setting.Enabled || setting.SampleRate <= 0 {
		return
	}
	if setting.CaptureModels != "" {
		if !matchModel(info.OriginModelName, setting.CaptureModels) {
			return
		}
	}
	if !shouldSampleForL2(info.RequestId, setting.SampleRate) {
		return
	}

	// Truncate if needed
	if len(body) > setting.MaxPayloadSize {
		body = body[:setting.MaxPayloadSize]
	}
	if !json.Valid(body) {
		body = truncateToJSONBoundary(body, setting.MaxPayloadSize)
	}

	common.SysLog(fmt.Sprintf("L2 sampled: requestId=%s model=%s bodyLen=%d", info.RequestId, info.OriginModelName, len(body)))

	storeRequestBody(info, body, setting.MaxPayloadSize)
	cleanupOldSamples(maxSamples)
}

// MaybeCaptureRequest decides whether to sample this request for L2 diagnostics,
// and if so, captures the raw request body asynchronously.
// This function is called after RelayInfo is populated but before the response is sent.
func MaybeCaptureRequest(c *gin.Context, info *relaycommon.RelayInfo) {
	setting := tokenscope_setting.GetTokenScopeSetting()
	if !setting.Enabled {
		return
	}
	if setting.SampleRate <= 0 {
		return
	}
	// Apply model filter
	if setting.CaptureModels != "" {
		if !matchModel(info.OriginModelName, setting.CaptureModels) {
			return
		}
	}
	// Hybrid sampling: round-robin + random
	if !shouldSampleForL2(info.RequestId, setting.SampleRate) {
		return
	}

	// CRITICAL: Read body storage BEFORE launching goroutine.
	// Gin recycles contexts via sync.Pool after response is sent, which clears
	// all context values. The goroutine would see an empty context if we defer
	// the read. Capturing here ensures the data survives past context recycling.
	storage, err := common.GetBodyStorage(c)
	if err != nil || storage == nil {
		return
	}
	data, err := storage.Bytes()
	if err != nil || len(data) == 0 {
		return
	}
	// Truncate if needed
	body := data
	if len(body) > setting.MaxPayloadSize {
		body = body[:setting.MaxPayloadSize]
	}
	// Validate and truncate to valid JSON boundary
	if !json.Valid(body) {
		body = truncateToJSONBoundary(data, setting.MaxPayloadSize)
	}

	// Capture asynchronously with captured data (no gin context needed)
	gopool.Go(func() {
		storeRequestBody(info, body, setting.MaxPayloadSize)
		// Enforce max samples: delete oldest if exceeding limit
		cleanupOldSamples(maxSamples)
	})
}

func cleanupOldSamples(limit int) {
	// Delete oldest records to keep only the newest `limit` records
	_, _ = model.DeleteOldestDebugPayloads(limit)
}

// shouldSample returns true based on deterministic hash of request_id.
func shouldSample(requestId string, rate float64) bool {
	if rate >= 1.0 {
		return true
	}
	// Use a simple hash-based approach
	var h uint32
	for _, b := range []byte(requestId) {
		h = h*31 + uint32(b)
	}
	return float64(h%10000)/10000.0 < rate
}

// matchModel checks if modelName matches the comma-separated capture_models filter.
func matchModel(modelName, captureModels string) bool {
	if captureModels == "" {
		return true
	}
	for _, m := range splitComma(captureModels) {
		if m == modelName {
			return true
		}
	}
	return false
}

func splitComma(s string) []string {
	var result []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			if start < i {
				result = append(result, s[start:i])
			}
			start = i + 1
		}
	}
	if start < len(s) {
		result = append(result, s[start:])
	}
	return result
}

// storeRequestBody persists the pre-captured request body to the database.
// The body bytes are captured BEFORE launching the goroutine to avoid gin
// context recycling issues (gin v1.9 uses sync.Pool which clears context
// values after response is sent).
func storeRequestBody(info *relaycommon.RelayInfo, body []byte, maxPayloadSize int) {
	if info == nil || len(body) == 0 {
		return
	}

	payload := &model.RequestDebugPayload{
		RequestId:   info.RequestId,
		LogId:       0, // Will be linked after log is created
		UserId:      info.UserId,
		ModelName:   info.OriginModelName,
		ChannelId:   info.ChannelId,
		TokenId:     info.TokenId,
		Group:       info.UsingGroup,
		RequestBody: string(body),
		IsStream:    info.IsStream,
		RelayFormat: string(info.GetFinalRequestRelayFormat()),
	}

	if err := model.CreateRequestDebugPayload(payload); err != nil {
		common.SysError("failed to create request debug payload: " + err.Error())
		return
	}

	// Parse context parts inline
	if err := parser.ParseAndStore(info.RequestId, body, info); err != nil {
		common.SysError("failed to parse context parts: " + err.Error())
	}
}

// truncateToJSONBoundary truncates data to a valid JSON boundary within maxBytes.
func truncateToJSONBoundary(data []byte, maxBytes int) []byte {
	if maxBytes <= 1 {
		return data[:maxBytes]
	}
	// Try progressively smaller sizes to find valid JSON
	for size := maxBytes; size > 0; size-- {
		if data[size-1] == '\n' || data[size-1] == '\r' || data[size-1] == ' ' {
			continue // skip trailing whitespace
		}
		truncated := data[:size]
		if json.Valid(truncated) {
			return truncated
		}
	}
	// If no valid JSON found, just return up to maxBytes
	if len(data) > maxBytes {
		return data[:maxBytes]
	}
	return data
}

// LinkDebugPayloadToLog updates a debug payload's log_id after the log is created.
func LinkDebugPayloadToLog(requestId string, logId int) {
	if requestId == "" || logId == 0 {
		return
	}
	gopool.Go(func() {
		_ = model.LOG_DB.Model(&model.RequestDebugPayload{}).
			Where("request_id = ?", requestId).
			Update("log_id", logId).Error
	})
}

// DetermineSamplingReason returns a reason for why a request should be sampled
// for L2 diagnostics. Returns empty string if no sampling reason applies.
func DetermineSamplingReason(info *relaycommon.RelayInfo, inputTokens, outputTokens int) string {
	if info == nil {
		return ""
	}

	// High context load: input/output ratio > 50
	if outputTokens > 0 && float64(inputTokens)/float64(outputTokens) > 50 {
		return "high_context_load"
	}

	// Large input: > 100K tokens
	if inputTokens > 100_000 {
		return "large_input"
	}

	return ""
}
