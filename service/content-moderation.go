package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/setting"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
)

// ModerationRequest 审核请求结构
type ModerationRequest struct {
	Model             string                     `json:"model"`
	MaxTokens         int                        `json:"max_tokens,omitempty"`
	TopP              float64                    `json:"top_p,omitempty"`
	Messages          []ModerationMessage        `json:"messages"`
	ResponseFormat    *ModerationResponseFormat  `json:"response_format,omitempty"`
	EnableThinking    bool                       `json:"enable_thinking,omitempty"`
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
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []ModerationChoice `json:"choices"`
}

// ModerationChoice 审核选择结构
type ModerationChoice struct {
	Index   int              `json:"index"`
	Message ModerationMessage `json:"message"`
}

// ModerationResult 审核结果
type ModerationResult struct {
	IsViolation     bool     `json:"status"`
	SensitiveWords  []string `json:"sensitiveWords,omitempty"`
}

// CheckContentModeration 检查内容是否违规
func CheckContentModeration(c *gin.Context, userContent string) (*ModerationResult, error) {
	if !setting.ShouldCheckContentModeration() {
		return &ModerationResult{IsViolation: false}, nil
	}

	if userContent == "" {
		return &ModerationResult{IsViolation: false}, nil
	}

	// 构建审核请求
	moderationReq := ModerationRequest{
		Model:         setting.ModerationModel,
		MaxTokens:     setting.ModerationMaxTokens,
		TopP:          0.7,
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
	apiUrl := fmt.Sprintf("%s/v1/chat/completions", strings.TrimRight(setting.ModerationApiBaseUrl, "/"))

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

	req.Header.Set("Authorization", "Bearer "+setting.ModerationApiKey)
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
	logger.LogInfo(c, fmt.Sprintf("content moderation: API raw response: %s", assistantContent))

	// 尝试解析JSON格式的响应
	var result ModerationResult
	if err := json.Unmarshal([]byte(assistantContent), &result); err != nil {
		// JSON解析失败，尝试从文本中提取状态和敏感词
		logger.LogWarn(c, fmt.Sprintf("content moderation: failed to parse JSON response: %v, raw content: %s", err, assistantContent))

		// 检查是否违规
		if strings.Contains(strings.ToLower(assistantContent), "\"status\":\"true\"") ||
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
		logger.LogWarn(c, fmt.Sprintf("content moderation: violation detected - sensitive words: %v", result.SensitiveWords))
	} else {
		logger.LogInfo(c, "content moderation: content passed review")
	}

	return &result, nil
}

// CheckContentModerationWithFallback 带降级策略的内容审核
func CheckContentModerationWithFallback(c *gin.Context, userContent string) (bool, []string) {
	// 首先尝试AI审核
	if setting.ShouldCheckContentModeration() {
		result, err := CheckContentModeration(c, userContent)
		if err == nil {
			if result.IsViolation {
				// AI检测到违规
				if len(result.SensitiveWords) > 0 {
					return true, result.SensitiveWords
				} else {
					return true, []string{"检测到不当内容"}
				}
			}
			return false, []string{}
		} else {
			// AI审核失败，记录错误但不阻断请求
			logger.LogError(c, fmt.Sprintf("content moderation API failed, falling back to sensitive word check: %v", err))
		}
	}

	// 降级到敏感词检测
	return CheckSensitiveText(userContent)
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