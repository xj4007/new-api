package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"one-api/common"
	"one-api/dto"
	"one-api/logger"
	"one-api/service"
	"one-api/setting"
	"strings"

	"github.com/gin-gonic/gin"
)

// ContentModerationMiddleware 内容审核中间件
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

		// 读取请求体
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.Next()
			return
		}

		// 恢复请求体，供后续处理使用
		c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

		// 提取用户内容并进行审核
		userContent, err := extractUserContentFromRequest(c, requestBody, path)
		if err != nil {
			c.Next()
			return
		}

		// 跳过不需要审核的内容
		if service.ShouldSkipModeration(userContent) {
			c.Next()
			return
		}

		// 执行内容审核（带降级策略）
		isViolation, sensitiveWords := service.CheckContentModerationWithFallback(c, userContent)
		if isViolation {
			handleModerationViolation(c, path, sensitiveWords)
			return
		}

		// 内容通过审核，继续处理
		c.Next()
	})
}

// shouldApplyModeration 判断是否应该对该路径应用内容审核
func shouldApplyModeration(path string) bool {
	// 对Claude和OpenAI聊天接口进行审核
	moderationPaths := []string{
		"/v1/messages",        // Claude格式
		"/v1/chat/completions", // OpenAI格式
	}

	for _, moderationPath := range moderationPaths {
		if strings.HasSuffix(path, moderationPath) {
			return true
		}
	}

	return false
}

// extractUserContentFromRequest 从请求中提取用户内容
func extractUserContentFromRequest(c *gin.Context, requestBody []byte, path string) (string, error) {
	if strings.HasSuffix(path, "/v1/messages") {
		// Claude格式
		return extractUserContentFromClaudeRequest(requestBody)
	} else if strings.HasSuffix(path, "/v1/chat/completions") {
		// OpenAI格式
		return extractUserContentFromOpenAIRequest(requestBody)
	}

	return "", fmt.Errorf("unsupported request format")
}

// extractUserContentFromOpenAIRequest 从OpenAI格式请求中提取用户内容
func extractUserContentFromOpenAIRequest(requestBody []byte) (string, error) {
	var openaiRequest dto.GeneralOpenAIRequest
	if err := json.Unmarshal(requestBody, &openaiRequest); err != nil {
		return "", fmt.Errorf("failed to parse OpenAI request: %v", err)
	}

	userContent := service.ExtractUserContent(openaiRequest.Messages)
	return userContent, nil
}

// extractUserContentFromClaudeRequest 从Claude格式请求中提取用户内容
func extractUserContentFromClaudeRequest(requestBody []byte) (string, error) {
	var claudeRequest dto.ClaudeRequest
	if err := json.Unmarshal(requestBody, &claudeRequest); err != nil {
		return "", fmt.Errorf("failed to parse Claude request: %v", err)
	}

	userContent := service.ExtractUserContentFromClaudeRequest(claudeRequest.Messages)
	return userContent, nil
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