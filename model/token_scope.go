package model

import (
	"fmt"
	"strconv"

	"github.com/QuantumNous/new-api/common"
)

// Dimension identifies the grouping axis for L1 queries.
type Dimension string

const (
	DimensionModel   Dimension = "model"
	DimensionUser    Dimension = "user"
	DimensionKey     Dimension = "key"
	DimensionChannel Dimension = "channel"
)

// ValidateDimension checks if the dimension string is valid.
func ValidateDimension(d string) (Dimension, error) {
	switch Dimension(d) {
	case DimensionModel, DimensionUser, DimensionKey, DimensionChannel:
		return Dimension(d), nil
	default:
		return "", fmt.Errorf("invalid dimension: %s", d)
	}
}

// TokenScopeL1Agg holds the raw SQL aggregation results for L1 metrics.
type TokenScopeL1Agg struct {
	Name              string `json:"name"`
	SubIdCol          int64  `json:"sub_id_col" gorm:"column:sub_id_col"`
	Bucket            string `json:"bucket"` // Used for timeseries queries
	RequestCount      int64  `json:"request_count"`
	TotalQuota        int64  `json:"total_quota"`
	TotalPromptTokens int64  `json:"total_prompt_tokens"`
	TotalOutputTokens int64  `json:"total_output_tokens"`
	TotalCacheRead    int64  `json:"total_cache_read"`
	TotalCacheWrite   int64  `json:"total_cache_write"`
	TotalInputTokens  int64  `json:"total_input_tokens"`
}

// TokenScopeL1Metrics holds the computed L1 metrics with derived ratios.
type TokenScopeL1Metrics struct {
	Name              string  `json:"name"`
	Dimension         string  `json:"dimension"`
	SubId             string  `json:"sub_id,omitempty"`
	SubName           string  `json:"sub_name,omitempty"`
	RequestCount      int64   `json:"request_count"`
	OutputCost        float64 `json:"output_cost"`
	ContextLoad       float64 `json:"context_load"`
	CacheReuseRate    float64 `json:"cache_reuse_rate"`
	TotalQuota        int64   `json:"total_quota"`
	TotalPromptTokens int64   `json:"total_prompt_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalCacheRead    int64   `json:"total_cache_read"`
	TotalCacheWrite   int64   `json:"total_cache_write"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
}

// TokenScopeL1TimePoint holds L1 metrics for a single time bucket.
type TokenScopeL1TimePoint struct {
	Bucket            string  `json:"bucket"`
	RequestCount      int64   `json:"request_count"`
	OutputCost        float64 `json:"output_cost"`
	ContextLoad       float64 `json:"context_load"`
	CacheReuseRate    float64 `json:"cache_reuse_rate"`
	TotalQuota        int64   `json:"total_quota"`
	TotalPromptTokens int64   `json:"total_prompt_tokens"`
	TotalOutputTokens int64   `json:"total_output_tokens"`
	TotalCacheRead    int64   `json:"total_cache_read"`
	TotalCacheWrite   int64   `json:"total_cache_write"`
	TotalInputTokens  int64   `json:"total_input_tokens"`
}

// aggToMetrics converts a raw aggregation result into computed L1 metrics.
func aggToMetrics(agg TokenScopeL1Agg) TokenScopeL1Metrics {
	m := TokenScopeL1Metrics{
		Name:              agg.Name,
		RequestCount:      agg.RequestCount,
		TotalQuota:        agg.TotalQuota,
		TotalPromptTokens: agg.TotalPromptTokens,
		TotalOutputTokens: agg.TotalOutputTokens,
		TotalCacheRead:    agg.TotalCacheRead,
		TotalCacheWrite:   agg.TotalCacheWrite,
		TotalInputTokens:  agg.TotalInputTokens,
	}
	if agg.TotalOutputTokens > 0 {
		m.OutputCost = float64(agg.TotalQuota) / float64(agg.TotalOutputTokens)
		m.ContextLoad = float64(agg.TotalPromptTokens) / float64(agg.TotalOutputTokens)
	}
	if agg.TotalPromptTokens+agg.TotalCacheRead > 0 {
		m.CacheReuseRate = float64(agg.TotalCacheRead) / float64(agg.TotalPromptTokens+agg.TotalCacheRead)
	}
	return m
}

// dimensionConfig holds SQL expressions for a given grouping dimension.
type dimensionConfig struct {
	nameExpr  string // SELECT ... as name
	groupCol  string // GROUP BY column
	excludeWhere string // e.g. "username != ''"
	subIdExpr string // e.g. "MIN(user_id) as sub_id_col" — empty if not needed
}

// getDimensionConfig returns the SQL configuration for a given dimension.
func getDimensionConfig(dimension Dimension) dimensionConfig {
	switch dimension {
	case DimensionModel:
		return dimensionConfig{
			nameExpr:  "model_name as name",
			groupCol:  "model_name",
			excludeWhere: "model_name != ''",
		}
	case DimensionUser:
		return dimensionConfig{
			nameExpr:  "username as name",
			groupCol:  "username",
			excludeWhere: "username != ''",
			subIdExpr: "MIN(user_id) as sub_id_col",
		}
	case DimensionKey:
		return dimensionConfig{
			nameExpr:  "token_name as name",
			groupCol:  "token_name",
			excludeWhere: "token_name != ''",
			subIdExpr: "MIN(token_id) as sub_id_col",
		}
	case DimensionChannel:
		nameExpr := "CAST(channel_id AS CHAR) as name"
		if common.UsingPostgreSQL {
			nameExpr = "channel_id::text as name"
		}
		return dimensionConfig{
			nameExpr:  nameExpr,
			groupCol:  "channel_id",
			excludeWhere: "channel_id != 0",
		}
	default:
		return dimensionConfig{
			nameExpr:  "model_name as name",
			groupCol:  "model_name",
			excludeWhere: "model_name != ''",
		}
	}
}

// GetTokenScopeL1ByDimension returns L1 metrics grouped by the specified dimension (admin).
func GetTokenScopeL1ByDimension(dimension Dimension, startTimestamp, endTimestamp int64, modelName, username string, channel int, group string) ([]TokenScopeL1Metrics, error) {
	return getTokenScopeL1ByDimension(dimension, startTimestamp, endTimestamp, modelName, username, channel, group, 0)
}

// GetUserTokenScopeL1ByDimension returns L1 metrics grouped by the specified dimension for a specific user.
func GetUserTokenScopeL1ByDimension(dimension Dimension, userId int, startTimestamp, endTimestamp int64, modelName string, group string) ([]TokenScopeL1Metrics, error) {
	return getTokenScopeL1ByDimension(dimension, startTimestamp, endTimestamp, modelName, "", 0, group, userId)
}

func getTokenScopeL1ByDimension(dimension Dimension, startTimestamp, endTimestamp int64, modelName, username string, channel int, group string, userId int) ([]TokenScopeL1Metrics, error) {
	cfg := getDimensionConfig(dimension)

	selectCols := cfg.nameExpr + ", count(*) as request_count, " +
		"COALESCE(SUM(quota),0) as total_quota, " +
		"COALESCE(SUM(prompt_tokens),0) as total_prompt_tokens, " +
		"COALESCE(SUM(completion_tokens),0) as total_output_tokens, " +
		"COALESCE(SUM(cache_read_tokens),0) as total_cache_read, " +
		"COALESCE(SUM(cache_write_tokens),0) as total_cache_write, " +
		"COALESCE(SUM(input_tokens_total),0) as total_input_tokens"

	if cfg.subIdExpr != "" {
		selectCols += ", " + cfg.subIdExpr
	}

	tx := LOG_DB.Table("logs").Select(selectCols)
	tx = tx.Where("type = ?", LogTypeConsume)

	if cfg.excludeWhere != "" {
		tx = tx.Where(cfg.excludeWhere)
	}
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where(logGroupCol+" = ?", group)
	}

	var aggs []TokenScopeL1Agg
	if err := tx.Group(cfg.groupCol).Scan(&aggs).Error; err != nil {
		return nil, fmt.Errorf("failed to query L1 metrics by %s: %w", dimension, err)
	}

	results := make([]TokenScopeL1Metrics, len(aggs))

	// For channel dimension, batch-resolve channel names
	var channelIds []int
	if dimension == DimensionChannel {
		for i, agg := range aggs {
			id, _ := strconv.Atoi(agg.Name)
			if id > 0 {
				channelIds = append(channelIds, id)
			}
			// Pre-convert: store channel_id as sub_id
			results[i] = aggToMetrics(agg)
			results[i].Dimension = string(dimension)
			results[i].SubId = agg.Name
		}
		if len(channelIds) > 0 {
			channels, err := GetChannelsByIds(channelIds)
			if err == nil {
				channelMap := make(map[int]string, len(channels))
				for _, ch := range channels {
					channelMap[ch.Id] = ch.Name
				}
				for i := range results {
					id, _ := strconv.Atoi(results[i].SubId)
					if name, ok := channelMap[id]; ok {
						results[i].SubName = name
						results[i].Name = name
					}
				}
			}
		}
	} else {
		for i, agg := range aggs {
			results[i] = aggToMetrics(agg)
			results[i].Dimension = string(dimension)
			if agg.SubIdCol > 0 {
				results[i].SubId = strconv.FormatInt(agg.SubIdCol, 10)
			}
		}
	}

	return results, nil
}

// GetTokenScopeL1ByModel returns L1 metrics grouped by model_name (backward-compatible wrapper).
func GetTokenScopeL1ByModel(startTimestamp, endTimestamp int64, modelName, username string, channel int, group string) ([]TokenScopeL1Metrics, error) {
	return getTokenScopeL1ByDimension(DimensionModel, startTimestamp, endTimestamp, modelName, username, channel, group, 0)
}

// GetUserTokenScopeL1ByModel returns L1 metrics grouped by model_name for a specific user (backward-compatible wrapper).
func GetUserTokenScopeL1ByModel(userId int, startTimestamp, endTimestamp int64, modelName string, group string) ([]TokenScopeL1Metrics, error) {
	return getTokenScopeL1ByDimension(DimensionModel, startTimestamp, endTimestamp, modelName, "", 0, group, userId)
}

// GetTokenScopeL1Summary returns the overall L1 metrics (all models combined).
func GetTokenScopeL1Summary(startTimestamp, endTimestamp int64, modelName, username string, channel int, group string) (*TokenScopeL1Metrics, error) {
	return getTokenScopeL1Summary(startTimestamp, endTimestamp, modelName, username, channel, group, 0)
}

// GetUserTokenScopeL1Summary returns the overall L1 metrics for a specific user.
func GetUserTokenScopeL1Summary(userId int, startTimestamp, endTimestamp int64, modelName string, group string) (*TokenScopeL1Metrics, error) {
	return getTokenScopeL1Summary(startTimestamp, endTimestamp, modelName, "", 0, group, userId)
}

func getTokenScopeL1Summary(startTimestamp, endTimestamp int64, modelName, username string, channel int, group string, userId int) (*TokenScopeL1Metrics, error) {
	selectCols := "'(all)' as name, count(*) as request_count, " +
		"COALESCE(SUM(quota),0) as total_quota, " +
		"COALESCE(SUM(prompt_tokens),0) as total_prompt_tokens, " +
		"COALESCE(SUM(completion_tokens),0) as total_output_tokens, " +
		"COALESCE(SUM(cache_read_tokens),0) as total_cache_read, " +
		"COALESCE(SUM(cache_write_tokens),0) as total_cache_write, " +
		"COALESCE(SUM(input_tokens_total),0) as total_input_tokens"

	tx := LOG_DB.Table("logs").Select(selectCols)
	tx = tx.Where("type = ?", LogTypeConsume)

	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where(logGroupCol+" = ?", group)
	}

	var agg TokenScopeL1Agg
	if err := tx.Scan(&agg).Error; err != nil {
		return nil, fmt.Errorf("failed to query L1 summary: %w", err)
	}

	m := aggToMetrics(agg)
	return &m, nil
}

// timeBucketExpr returns the SQL expression for time bucketing, compatible with all three DBs.
// bucket must be "hour" or "day".
func timeBucketExpr(bucket string) string {
	switch {
	case common.UsingPostgreSQL:
		if bucket == "hour" {
			return "date_trunc('hour', to_timestamp(created_at)) as bucket"
		}
		return "date_trunc('day', to_timestamp(created_at)) as bucket"
	case common.UsingSQLite:
		if bucket == "hour" {
			return "strftime('%Y-%m-%d %H:00', datetime(created_at, 'unixepoch')) as bucket"
		}
		return "strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) as bucket"
	default: // MySQL
		if bucket == "hour" {
			return "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00') as bucket"
		}
		return "DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') as bucket"
	}
}

// GetTokenScopeL1TimeSeries returns L1 metrics bucketed by time.
func GetTokenScopeL1TimeSeries(startTimestamp, endTimestamp int64, modelName, username string, channel int, group string, bucket string) ([]TokenScopeL1TimePoint, error) {
	if bucket == "" {
		bucket = "day"
	}

	bucketExpr := timeBucketExpr(bucket)
	selectCols := bucketExpr + ", count(*) as request_count, " +
		"COALESCE(SUM(quota),0) as total_quota, " +
		"COALESCE(SUM(prompt_tokens),0) as total_prompt_tokens, " +
		"COALESCE(SUM(completion_tokens),0) as total_output_tokens, " +
		"COALESCE(SUM(cache_read_tokens),0) as total_cache_read, " +
		"COALESCE(SUM(cache_write_tokens),0) as total_cache_write, " +
		"COALESCE(SUM(input_tokens_total),0) as total_input_tokens"

	tx := LOG_DB.Table("logs").Select(selectCols)
	tx = tx.Where("type = ?", LogTypeConsume)

	if startTimestamp != 0 {
		tx = tx.Where("created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if username != "" {
		tx = tx.Where("username = ?", username)
	}
	if channel != 0 {
		tx = tx.Where("channel_id = ?", channel)
	}
	if group != "" {
		tx = tx.Where(logGroupCol+" = ?", group)
	}

	var aggs []TokenScopeL1Agg
	if err := tx.Group("bucket").Order("bucket ASC").Scan(&aggs).Error; err != nil {
		return nil, fmt.Errorf("failed to query L1 timeseries: %w", err)
	}

	results := make([]TokenScopeL1TimePoint, len(aggs))
	for i, agg := range aggs {
		m := aggToMetrics(agg)
		results[i] = TokenScopeL1TimePoint{
			Bucket:            agg.Bucket,
			RequestCount:      m.RequestCount,
			OutputCost:        m.OutputCost,
			ContextLoad:       m.ContextLoad,
			CacheReuseRate:    m.CacheReuseRate,
			TotalQuota:        m.TotalQuota,
			TotalPromptTokens: m.TotalPromptTokens,
			TotalOutputTokens: m.TotalOutputTokens,
			TotalCacheRead:    m.TotalCacheRead,
			TotalCacheWrite:   m.TotalCacheWrite,
			TotalInputTokens:  m.TotalInputTokens,
		}
	}
	return results, nil
}

// ---------------------------------------------------------------------------
// L2 Deep Diagnostics Queries
// ---------------------------------------------------------------------------

// TokenScopeL2Summary holds aggregated L2 diagnostic metrics.
type TokenScopeL2Summary struct {
	SampleCount         int64   `json:"sample_count"`
	AvgSystemTokens     float64 `json:"avg_system_tokens"`
	AvgHistoryTokens    float64 `json:"avg_history_tokens"`
	AvgToolTokens       float64 `json:"avg_tool_tokens"`
	AvgFileTokens       float64 `json:"avg_file_tokens"`
	RepeatedPrefixRate  float64 `json:"repeated_prefix_rate"`
	CacheFriendliness   float64 `json:"cache_friendliness"`
	CacheFulfillmentRate float64 `json:"cache_fulfillment_rate"`
}

// GetTokenScopeL2Summary returns L2 diagnostic metrics grouped by model.
func GetTokenScopeL2Summary(startTimestamp, endTimestamp int64, modelName string, group string) ([]TokenScopeL2Summary, error) {
	selectCols := "count(DISTINCT rdp.request_id) as sample_count, " +
		"COALESCE(AVG(CASE WHEN rc.part_type = 'system' THEN rc.token_count ELSE NULL END), 0) as avg_system_tokens, " +
		"COALESCE(AVG(CASE WHEN rc.part_type = 'history' THEN rc.token_count ELSE NULL END), 0) as avg_history_tokens, " +
		"COALESCE(AVG(CASE WHEN rc.part_type = 'tool' THEN rc.token_count ELSE NULL END), 0) as avg_tool_tokens, " +
		"COALESCE(AVG(CASE WHEN rc.part_type = 'file' THEN rc.token_count ELSE NULL END), 0) as avg_file_tokens, " +
		"COALESCE(AVG(CASE WHEN rc.is_repeated = true AND rc.is_prefix = true THEN 1.0 ELSE 0.0 END), 0) as repeated_prefix_rate, " +
		"COALESCE(AVG(CASE WHEN rc.is_prefix = true AND rc.is_stable = true THEN 1.0 ELSE 0.0 END), 0) as cache_friendliness, " +
		"COALESCE(AVG(CASE WHEN rc.is_cache_friendly = true THEN 1.0 ELSE 0.0 END), 0) as cache_fulfillment_rate"

	tx := LOG_DB.Table("request_debug_payload rdp").
		Joins("INNER JOIN request_context_parts rc ON rc.request_id = rdp.request_id").
		Select(selectCols).
		Group("rdp.model_name")

	if startTimestamp != 0 {
		tx = tx.Where("rdp.created_at >= ?", startTimestamp)
	}
	if endTimestamp != 0 {
		tx = tx.Where("rdp.created_at <= ?", endTimestamp)
	}
	if modelName != "" {
		tx = tx.Where("rdp.model_name = ?", modelName)
	}
	if group != "" {
		tx = tx.Where("rdp."+logGroupCol+" = ?", group)
	}

	var summaries []TokenScopeL2Summary
	if err := tx.Scan(&summaries).Error; err != nil {
		return nil, fmt.Errorf("failed to query L2 summary: %w", err)
	}
	return summaries, nil
}

// TokenScopeFilterOptions holds the available filter values for the tokenscope UI.
type TokenScopeFilterOptions struct {
	ModelNames []string `json:"model_names"`
	Groups     []string `json:"groups"`
}

// GetTokenScopeFilterOptions returns distinct model_name and group values from logs.
func GetTokenScopeFilterOptions(userId int) (*TokenScopeFilterOptions, error) {
	opts := &TokenScopeFilterOptions{}

	tx := LOG_DB.Table("logs").Where("type = ?", LogTypeConsume)
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}

	if err := tx.Distinct("model_name").Where("model_name != ''").Pluck("model_name", &opts.ModelNames).Error; err != nil {
		return nil, fmt.Errorf("failed to query distinct model names: %w", err)
	}

	tx2 := LOG_DB.Table("logs").Where("type = ?", LogTypeConsume)
	if userId > 0 {
		tx2 = tx2.Where("user_id = ?", userId)
	}

	if err := tx2.Distinct(logGroupCol).Where(logGroupCol+" != ''").Pluck(logGroupCol, &opts.Groups).Error; err != nil {
		return nil, fmt.Errorf("failed to query distinct groups: %w", err)
	}

	return opts, nil
}

// GetRequestContextPartsByRequestId returns context parts for a specific request.
func GetRequestDebugPayloadByRequestId(requestId string) (*RequestDebugPayload, error) {
	var payload RequestDebugPayload
	err := LOG_DB.Where("request_id = ?", requestId).First(&payload).Error
	if err != nil {
		return nil, err
	}
	return &payload, nil
}

// GetRecentDebugRequests returns recent debug requests for a user/model.
func GetRecentDebugRequests(userId int, modelName string, group string, limit int) ([]RequestDebugPayload, error) {
	var payloads []RequestDebugPayload
	tx := LOG_DB.Where("parsed = ?", true)
	if userId > 0 {
		tx = tx.Where("user_id = ?", userId)
	}
	if modelName != "" {
		tx = tx.Where("model_name = ?", modelName)
	}
	if group != "" {
		tx = tx.Where(logGroupCol+" = ?", group)
	}
	err := tx.Order("created_at DESC").Limit(limit).Find(&payloads).Error
	return payloads, err
}

// GetUserRecentDebugRequests returns recent debug requests for the current user.
func GetUserRecentDebugRequests(userId int, modelName string, group string, limit int) ([]RequestDebugPayload, error) {
	return GetRecentDebugRequests(userId, modelName, group, limit)
}
