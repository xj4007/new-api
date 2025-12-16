package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/constant"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
)

// ContentModerationMiddleware 内容审核中间件
// 支持以下功能：
// 1. 渠道ID过滤 - 只对配置的渠道进行审核
// 2. Session级缓存 - 连续通过3次审核后，30分钟内免审核
// 3. 智能内容提取 - 提取最近3条user和3条assistant消息
// 4. 二级审核机制 - Phase 1快速检测 + Phase 2 Pro模型复核
func ContentModerationMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// 检查是否启用内容审核
		if !setting.ShouldCheckContentModeration() {
			c.Next()
			return
		}

		// 只对相关路径进行审核
		path := c.Request.URL.Path
		if !shouldApplyModeration(path) {
			c.Next()
			return
		}

		// 检查渠道ID过滤 (需要在Distribute之后执行才有渠道信息)
		channelId := common.GetContextKeyInt(c, constant.ContextKeyChannelId)
		if channelId > 0 && !setting.ShouldModerateChannel(channelId) {
			logger.LogInfo(c, fmt.Sprintf("content moderation: skipped for channel %d (not in moderation list)", channelId))
			c.Next()
			return
		}

		// 读取请求体
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// 恢复请求体，供后续处理使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		// 检查Session缓存 (仅Claude请求)
		var sessionId string
		if strings.HasSuffix(path, "/v1/messages") && setting.ModerationSessionCacheEnabled {
			sessionId = service.ExtractSessionIdFromClaudeRequest(requestBody)
			if sessionId != "" {
				if service.CheckSessionCache(sessionId) {
					logger.LogInfo(c, fmt.Sprintf("content moderation: session %s passed (cached, threshold reached)", service.TruncateSessionId(sessionId)))
					c.Next()
					return
				}
				// 显示当前通过次数
				currentCount := service.GetSessionPassCount(sessionId)
				logger.LogInfo(c, fmt.Sprintf("content moderation: session %s check #%d (need %d to skip)", service.TruncateSessionId(sessionId), currentCount+1, service.SessionPassThreshold))
			}
		}

		// 智能提取用户内容
		userContent, err := extractUserContentFromRequestSmart(requestBody, path)
		if err != nil {
			logger.LogWarn(c, fmt.Sprintf("content moderation: failed to extract content: %v", err))
			c.Next()
			return
		}

		// 跳过不需要审核的内容
		if service.ShouldSkipModeration(userContent) {
			// 内容被跳过时也增加计数
			if sessionId != "" && setting.ModerationSessionCacheEnabled {
				newCount, reachedThreshold, _ := service.IncrementSessionPassCount(sessionId)
				if reachedThreshold {
					logger.LogInfo(c, fmt.Sprintf("content moderation: session %s reached threshold (%d/%d), now exempt for %ds", service.TruncateSessionId(sessionId), newCount, service.SessionPassThreshold, setting.ModerationSessionCacheTTL))
				}
			}
			c.Next()
			return
		}

		// 执行二级审核
		isViolation, sensitiveWords := service.CheckContentModerationTwoPhase(c, userContent)
		if isViolation {
			handleModerationViolation(c, path, sensitiveWords)
			return
		}

		// 审核通过，增加Session通过计数
		if sessionId != "" && setting.ModerationSessionCacheEnabled {
			newCount, reachedThreshold, err := service.IncrementSessionPassCount(sessionId)
			if err != nil {
				logger.LogWarn(c, fmt.Sprintf("content moderation: failed to increment session count: %v", err))
			} else if reachedThreshold {
				logger.LogInfo(c, fmt.Sprintf("content moderation: session %s reached threshold (%d/%d), now exempt for %ds", service.TruncateSessionId(sessionId), newCount, service.SessionPassThreshold, setting.ModerationSessionCacheTTL))
			} else {
				logger.LogInfo(c, fmt.Sprintf("content moderation: session %s passed (%d/%d)", service.TruncateSessionId(sessionId), newCount, service.SessionPassThreshold))
			}
		}

		// 内容通过审核，继续处理
		c.Next()
	})
}

// shouldApplyModeration 判断是否应该对该路径应用内容审核
func shouldApplyModeration(path string) bool {
	// 对Claude和OpenAI聊天接口进行审核
	moderationPaths := []string{
		"/v1/messages",         // Claude格式
		"/v1/chat/completions", // OpenAI格式
	}

	for _, moderationPath := range moderationPaths {
		if strings.HasSuffix(path, moderationPath) {
			return true
		}
	}

	return false
}

// extractUserContentFromRequestSmart 智能提取用户内容
// 每个消息/系统提示词抽取前N字符
func extractUserContentFromRequestSmart(requestBody []byte, path string) (string, error) {
	maxChars := setting.ModerationContentMaxChars

	if strings.HasSuffix(path, "/v1/messages") {
		// Claude格式
		var claudeRequest dto.ClaudeRequest
		if err := json.Unmarshal(requestBody, &claudeRequest); err != nil {
			return "", fmt.Errorf("failed to parse Claude request: %v", err)
		}
		return service.ExtractClaudeContentSmart(claudeRequest.Messages, claudeRequest.System, maxChars), nil
	} else if strings.HasSuffix(path, "/v1/chat/completions") {
		// OpenAI格式
		var openaiRequest dto.GeneralOpenAIRequest
		if err := json.Unmarshal(requestBody, &openaiRequest); err != nil {
			return "", fmt.Errorf("failed to parse OpenAI request: %v", err)
		}
		return service.ExtractOpenAIContentSmart(openaiRequest.Messages, maxChars), nil
	}

	return "", fmt.Errorf("unsupported request format")
}

// handleModerationViolation 处理内容审核违规
func handleModerationViolation(c *gin.Context, path string, sensitiveWords []string) {
	requestId := c.GetString(common.RequestIdKey)

	// 构建违规信息
	violationMessage := "检测到不当内容，请修改后重试。"
	if len(sensitiveWords) > 0 {
		violationMessage = fmt.Sprintf("检测到违规词汇：[%s]，请修改后重试。", strings.Join(sensitiveWords, "、"))
	}

	// 添加友好提示
	violationMessage += "在Claude Code中按ESC+ESC可返回上次输入进行修改。"

	// 添加请求ID
	if requestId != "" {
		violationMessage = common.MessageWithRequestId(violationMessage, requestId)
	}

	// 记录违规日志
	logger.LogWarn(c, fmt.Sprintf("content moderation violation: %s", violationMessage))

	// 根据不同的API格式返回相应的错误格式
	if strings.HasSuffix(path, "/v1/messages") {
		// Claude格式错误响应
		c.JSON(http.StatusBadRequest, gin.H{
			"type": "error",
			"error": gin.H{
				"type":    "invalid_request_error",
				"message": violationMessage,
			},
		})
	} else {
		// OpenAI格式错误响应
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"message": violationMessage,
				"type":    "invalid_request_error",
				"param":   nil,
				"code":    "content_policy_violation",
			},
		})
	}

	c.Abort()
}

// ContentModerationStats 内容审核统计（可选功能）
type ContentModerationStats struct {
	TotalRequests     int64
	ViolationCount    int64
	FallbackCount     int64
	LastViolationTime int64
}

// UpdateModerationStats 更新审核统计（可选功能，用于监控）
func UpdateModerationStats(isViolation, isFallback bool) {
	// 这里可以实现统计逻辑，例如写入Redis或内存缓存
	// 暂时留空，可以根据需要实现
}
