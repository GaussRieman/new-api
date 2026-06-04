package controller

import (
	"errors"
	"strconv"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// GetTokenScopeL1Summary returns the overall L1 metrics for admin.
func GetTokenScopeL1Summary(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	username := c.Query("username")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")

	metrics, err := model.GetTokenScopeL1Summary(startTimestamp, endTimestamp, modelName, username, channel, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeL1ByModel returns L1 metrics grouped by model for admin.
func GetTokenScopeL1ByModel(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	username := c.Query("username")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")

	metrics, err := model.GetTokenScopeL1ByModel(startTimestamp, endTimestamp, modelName, username, channel, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeL1TimeSeries returns L1 time series data for admin.
func GetTokenScopeL1TimeSeries(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	username := c.Query("username")
	channel, _ := strconv.Atoi(c.Query("channel"))
	group := c.Query("group")
	bucket := c.Query("bucket")
	if bucket == "" {
		bucket = "day"
	}

	metrics, err := model.GetTokenScopeL1TimeSeries(startTimestamp, endTimestamp, modelName, username, channel, group, bucket)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeSelfL1Summary returns the overall L1 metrics for the current user.
func GetTokenScopeSelfL1Summary(c *gin.Context) {
	userId := c.GetInt("id")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	group := c.Query("group")

	metrics, err := model.GetUserTokenScopeL1Summary(userId, startTimestamp, endTimestamp, modelName, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeSelfL1ByModel returns L1 metrics grouped by model for the current user.
func GetTokenScopeSelfL1ByModel(c *gin.Context) {
	userId := c.GetInt("id")
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	group := c.Query("group")

	metrics, err := model.GetUserTokenScopeL1ByModel(userId, startTimestamp, endTimestamp, modelName, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeL2Summary returns L2 diagnostic metrics for admin.
func GetTokenScopeL2Summary(c *gin.Context) {
	startTimestamp, _ := strconv.ParseInt(c.Query("start_timestamp"), 10, 64)
	endTimestamp, _ := strconv.ParseInt(c.Query("end_timestamp"), 10, 64)
	modelName := c.Query("model_name")
	group := c.Query("group")

	metrics, err := model.GetTokenScopeL2Summary(startTimestamp, endTimestamp, modelName, group)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, metrics)
}

// GetTokenScopeFilterOptions returns distinct model_name and group values for admin.
func GetTokenScopeFilterOptions(c *gin.Context) {
	options, err := model.GetTokenScopeFilterOptions(0)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, options)
}

// GetTokenScopeSelfFilterOptions returns distinct model_name and group values for the current user.
func GetTokenScopeSelfFilterOptions(c *gin.Context) {
	userId := c.GetInt("id")
	options, err := model.GetTokenScopeFilterOptions(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, options)
}

// GetTokenScopeL2RequestDetail returns the context parts for a specific request.
func GetTokenScopeL2RequestDetail(c *gin.Context) {
	requestId := c.Param("request_id")
	if requestId == "" {
		common.ApiError(c, errors.New("request_id is required"))
		return
	}

	payload, err := model.GetRequestDebugPayloadByRequestId(requestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	parts, err := model.GetRequestContextPartsByRequestId(requestId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"payload": payload,
		"parts":   parts,
	})
}

// GetTokenScopeL2RecentRequests returns recent debug requests for admin.
func GetTokenScopeL2RecentRequests(c *gin.Context) {
	modelName := c.Query("model_name")
	group := c.Query("group")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 20
	}

	requests, err := model.GetRecentDebugRequests(0, modelName, group, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, requests)
}

// GetTokenScopeSelfL2RecentRequests returns recent debug requests for the current user.
func GetTokenScopeSelfL2RecentRequests(c *gin.Context) {
	userId := c.GetInt("id")
	modelName := c.Query("model_name")
	group := c.Query("group")
	limit, _ := strconv.Atoi(c.Query("limit"))
	if limit <= 0 {
		limit = 20
	}

	requests, err := model.GetUserRecentDebugRequests(userId, modelName, group, limit)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, requests)
}
