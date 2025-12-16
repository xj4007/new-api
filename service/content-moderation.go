package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting"

	"github.com/gin-gonic/gin"
)

// ModerationRequest 审核请求结构
type ModerationRequest struct {
	Model          string                    `json:"model"`
	MaxTokens      int                       `json:"max_tokens,omitempty"`
	TopP           float64                   `json:"top_p,omitempty"`
	Messages       []ModerationMessage       `json:"messages"`
	ResponseFormat *ModerationResponseFormat `json:"response_format,omitempty"`
	EnableThinking bool                      `json:"enable_thinking,omitempty"`
}

// ModerationMessage 审核消息结构
type ModerationMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ModerationResponseFormat 审核响应格式
type ModerationResponseFormat struct {
	Type string `json:"type"`
}

// ModerationResponse 审核响应结构
type ModerationResponse struct {
	ID      string             `json:"id"`
	Object  string             `json:"object"`
	Created int64              `json:"created"`
	Model   string             `json:"model"`
	Choices []ModerationChoice `json:"choices"`
}

// ModerationChoice 审核选择结构
type ModerationChoice struct {
	Index   int               `json:"index"`
	Message ModerationMessage `json:"message"`
}

// ModerationResult 审核结果
type ModerationResult struct {
	IsViolation    bool     `json:"status"`
	SensitiveWords []string `json:"sensitiveWords,omitempty"`
}

// CheckContentModeration 使用默认配置检查内容是否违规
func CheckContentModeration(c *gin.Context, userContent string) (*ModerationResult, error) {
	return CheckContentModerationWithModel(c, userContent, setting.ModerationModel, setting.ModerationApiKey, setting.ModerationApiBaseUrl)
}

// CheckContentModerationWithRetry 带重试机制的内容审核
// 如果连续失败达到最大重试次数，根据配置决定是放行还是降级
func CheckContentModerationWithRetry(c *gin.Context, userContent, model, apiKey, apiBaseUrl string) (*ModerationResult, error) {
	maxRetries := setting.ModerationMaxRetries
	if maxRetries <= 0 {
		maxRetries = 3
	}

	var lastErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			logger.LogWarn(c, fmt.Sprintf("content moderation: retry %d/%d", i+1, maxRetries))
			// 重试前等待一小段时间
			time.Sleep(time.Duration(i*500) * time.Millisecond)
		}

		result, err := CheckContentModerationWithModel(c, userContent, model, apiKey, apiBaseUrl)
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	// 所有重试都失败了
	logger.LogError(c, fmt.Sprintf("content moderation: all %d retries failed, last error: %v", maxRetries, lastErr))
	return nil, lastErr
}

// CheckContentModerationWithModel 使用指定模型检查内容是否违规
func CheckContentModerationWithModel(c *gin.Context, userContent, model, apiKey, apiBaseUrl string) (*ModerationResult, error) {
	if userContent == "" {
		return &ModerationResult{IsViolation: false}, nil
	}

	if model == "" || apiKey == "" {
		return &ModerationResult{IsViolation: false}, fmt.Errorf("model or apiKey is empty")
	}

	if apiBaseUrl == "" {
		apiBaseUrl = setting.ModerationApiBaseUrl
	}

	// 构建审核请求
	moderationReq := ModerationRequest{
		Model:          model,
		MaxTokens:      setting.ModerationMaxTokens,
		TopP:           0.7,
		EnableThinking: false,
		Messages: []ModerationMessage{
			{
				Role:    "system",
				Content: setting.GetModerationSystemPrompt(),
			},
			{
				Role:    "user",
				Content: userContent,
			},
		},
		ResponseFormat: &ModerationResponseFormat{
			Type: "json_object",
		},
	}

	// 发送API请求
	apiUrl := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(apiBaseUrl, "/"))

	jsonData, err := json.Marshal(moderationReq)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("content moderation: failed to marshal request: %v", err))
		return nil, err
	}

	req, err := http.NewRequest("POST", apiUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		logger.LogError(c, fmt.Sprintf("content moderation: failed to create request: %v", err))
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	// 设置超时
	client := &http.Client{
		Timeout: setting.ModerationTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		logger.LogError(c, fmt.Sprintf("content moderation: API request failed: %v", err))
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		logger.LogError(c, fmt.Sprintf("content moderation: API returned status %d", resp.StatusCode))
		return nil, fmt.Errorf("API request failed with status %d", resp.StatusCode)
	}

	// 解析响应
	var moderationResp ModerationResponse
	if err := json.NewDecoder(resp.Body).Decode(&moderationResp); err != nil {
		logger.LogError(c, fmt.Sprintf("content moderation: failed to decode response: %v", err))
		return nil, err
	}

	if len(moderationResp.Choices) == 0 {
		logger.LogError(c, "content moderation: no choices in response")
		return nil, fmt.Errorf("no choices in response")
	}

	// 解析AI返回的JSON结果
	assistantContent := moderationResp.Choices[0].Message.Content
	assistantContent = strings.TrimSpace(assistantContent)

	// 记录原始API响应用于调试
	logger.LogInfo(c, fmt.Sprintf("content moderation [%s]: API raw response: %s", model, assistantContent))

	// 尝试解析JSON格式的响应
	var result ModerationResult
	if err := json.Unmarshal([]byte(assistantContent), &result); err != nil {
		// JSON解析失败，尝试从文本中提取状态和敏感词
		logger.LogWarn(c, fmt.Sprintf("content moderation: failed to parse JSON response: %v, raw content: %s", err, assistantContent))

		// 检查是否违规
		if strings.Contains(strings.ToLower(assistantContent), "\"status\":true") ||
			strings.Contains(strings.ToLower(assistantContent), "\"status\": true") ||
			strings.Contains(strings.ToLower(assistantContent), "\"status\":\"true\"") ||
			strings.Contains(strings.ToLower(assistantContent), "\"status\": \"true\"") {
			result.IsViolation = true

			// 尝试提取 sensitiveWords 数组
			sensitiveWords := extractSensitiveWordsFromText(assistantContent)
			if len(sensitiveWords) > 0 {
				result.SensitiveWords = sensitiveWords
			} else {
				// 如果没有找到sensitiveWords字段，提供默认提示
				result.SensitiveWords = []string{"内容包含不当信息，请检查并修改"}
			}
		} else {
			result.IsViolation = false
			result.SensitiveWords = []string{}
		}
	}

	// 记录审核结果
	if result.IsViolation {
		logger.LogWarn(c, fmt.Sprintf("content moderation [%s]: violation detected - sensitive words: %v", model, result.SensitiveWords))
	} else {
		logger.LogInfo(c, fmt.Sprintf("content moderation [%s]: content passed review", model))
	}

	return &result, nil
}

// CheckContentModerationTwoPhase 二级审核机制
// Phase 1: 使用默认模型快速检测（带重试）
// Phase 2: 可选，如果配置了Pro模型，用于确认违规（不再用于"清除误报"）
// 注意：Phase 1 检测到违规就直接阻止，不会被 Phase 2 "清除"
func CheckContentModerationTwoPhase(c *gin.Context, userContent string) (bool, []string) {
	if !setting.ShouldCheckContentModeration() {
		return false, []string{}
	}

	// 记录待审核内容（用于调试，限制长度）
	contentPreview := userContent
	if len(contentPreview) > 200 {
		contentPreview = contentPreview[:200] + "..."
	}
	logger.LogInfo(c, fmt.Sprintf("content moderation: checking content (preview): %s", contentPreview))

	// Phase 1: 默认模型快速检测（带重试）
	logger.LogInfo(c, "content moderation Phase 1: starting fast detection")
	result, err := CheckContentModerationWithRetry(
		c,
		userContent,
		setting.ModerationModel,
		setting.ModerationApiKey,
		setting.ModerationApiBaseUrl,
	)
	if err != nil {
		// Phase 1 所有重试都失败了
		logger.LogError(c, fmt.Sprintf("content moderation Phase 1: all retries failed: %v", err))

		// 根据配置决定是放行还是降级到敏感词检测
		if setting.ModerationFailOpen {
			logger.LogWarn(c, "content moderation: fail-open mode, allowing request to pass")
			return false, []string{}
		}

		// 降级到敏感词检测
		logger.LogWarn(c, "content moderation: fail-close mode, falling back to sensitive word check")
		return CheckSensitiveText(userContent)
	}

	if !result.IsViolation {
		// Phase 1 通过，直接放行
		logger.LogInfo(c, "content moderation Phase 1: passed")
		return false, []string{}
	}

	// Phase 1 检测到违规，记录详细信息
	logger.LogWarn(c, fmt.Sprintf("content moderation Phase 1: VIOLATION DETECTED! sensitive words: %v", result.SensitiveWords))
	logger.LogWarn(c, fmt.Sprintf("content moderation: blocked content (preview): %s", contentPreview))

	// 直接返回违规结果，不再让 Phase 2 "清除"
	// Phase 2 仅用于提供更准确的敏感词（可选，不影响阻止决定）
	sensitiveWords := result.SensitiveWords
	if len(sensitiveWords) == 0 {
		sensitiveWords = []string{"检测到不当内容"}
	}

	// 如果配置了Pro模型，可以用来获取更准确的敏感词（但不会改变阻止决定）
	if setting.HasProModel() {
		logger.LogInfo(c, "content moderation Phase 2: using Pro model for detailed analysis (will not override block decision)")
		proResult, err := CheckContentModerationWithRetry(
			c,
			userContent,
			setting.ModerationProModel,
			setting.ModerationProApiKey,
			setting.GetProApiBaseUrl(),
		)
		if err == nil && proResult.IsViolation && len(proResult.SensitiveWords) > 0 {
			// Pro模型也检测到违规，使用其更准确的敏感词
			logger.LogWarn(c, fmt.Sprintf("content moderation Phase 2: confirmed with sensitive words: %v", proResult.SensitiveWords))
			sensitiveWords = proResult.SensitiveWords
		} else if err == nil && !proResult.IsViolation {
			// Pro模型认为没问题，但我们仍然阻止（记录日志用于分析）
			logger.LogWarn(c, "content moderation Phase 2: Pro model disagreed, but blocking anyway (Phase 1 detected violation)")
		}
	}

	// 返回违规结果
	return true, sensitiveWords
}

// CheckContentModerationWithFallback 带降级策略的内容审核（向后兼容）
func CheckContentModerationWithFallback(c *gin.Context, userContent string) (bool, []string) {
	// 使用二级审核机制
	return CheckContentModerationTwoPhase(c, userContent)
}

// extractSensitiveWordsFromText 从文本中提取敏感词数组
func extractSensitiveWordsFromText(text string) []string {
	// 尝试提取 sensitiveWords 数组
	// 匹配模式：\"sensitiveWords\":[\"word1\",\"word2\"] 或 \"sensitiveWords\": [\"word1\", \"word2\"]
	re := regexp.MustCompile(`"sensitiveWords"\s*:\s*\[(.*?)\]`)
	matches := re.FindStringSubmatch(text)

	if len(matches) >= 2 {
		wordsStr := matches[1]
		if strings.TrimSpace(wordsStr) == "" {
			return []string{}
		}

		// 提取引号内的词汇
		wordRe := regexp.MustCompile(`"([^"]+)"`)
		wordMatches := wordRe.FindAllStringSubmatch(wordsStr, -1)

		var words []string
		for _, match := range wordMatches {
			if len(match) >= 2 {
				word := strings.TrimSpace(match[1])
				if word != "" {
					words = append(words, word)
				}
			}
		}
		return words
	}

	return []string{}
}
