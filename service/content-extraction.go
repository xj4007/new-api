package service

import (
	"one-api/dto"
	"regexp"
	"strings"
)

// ExtractUserContent 从请求中提取用户输入的文本内容
func ExtractUserContent(messages []dto.Message) string {
	if len(messages) == 0 {
		return ""
	}

	// 从后向前遍历消息，找最后一个用户消息
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		if strings.ToLower(message.Role) == "user" {
			return extractTextFromMessage(message)
		}
	}

	return ""
}

// extractTextFromMessage 从单个消息中提取文本内容
func extractTextFromMessage(message dto.Message) string {
	var textParts []string

	// 如果是简单字符串内容
	if message.IsStringContent() {
		content := message.StringContent()
		if content != "" {
			textParts = append(textParts, content)
		}
	} else {
		// 如果是复合内容数组，提取其中的文本部分
		parsedContent := message.ParseContent()
		for _, content := range parsedContent {
			if content.Type == "text" && content.Text != "" {
				textParts = append(textParts, content.Text)
			}
		}
	}

	if len(textParts) == 0 {
		return ""
	}

	// 合并所有文本部分
	fullText := strings.Join(textParts, "\n")

	// 清理文本：移除系统提醒标签和其他系统内容
	fullText = cleanUserContent(fullText)

	return strings.TrimSpace(fullText)
}

// cleanUserContent 清理用户内容，移除系统标签和其他不需要审核的内容
func cleanUserContent(content string) string {
	// 移除 <system-reminder> 标签及其内容
	systemReminderRegex := regexp.MustCompile(`<system-reminder>[\s\S]*?</system-reminder>`)
	content = systemReminderRegex.ReplaceAllString(content, "")

	// 移除其他可能的系统标签
	systemTagRegex := regexp.MustCompile(`<system[^>]*>[\s\S]*?</system[^>]*>`)
	content = systemTagRegex.ReplaceAllString(content, "")

	// 移除可能的内部标记
	internalMarkRegex := regexp.MustCompile(`\[INTERNAL[^\]]*\][\s\S]*?\[/INTERNAL[^\]]*\]`)
	content = internalMarkRegex.ReplaceAllString(content, "")

	// 移除多余的空白行
	content = regexp.MustCompile(`\n\s*\n\s*\n`).ReplaceAllString(content, "\n\n")

	return content
}

// ExtractUserContentFromClaudeRequest 从Claude请求格式提取用户内容
func ExtractUserContentFromClaudeRequest(claudeMessages []dto.ClaudeMessage) string {
	if len(claudeMessages) == 0 {
		return ""
	}

	// 从后向前遍历消息，找最后一个用户消息
	for i := len(claudeMessages) - 1; i >= 0; i-- {
		message := claudeMessages[i]
		if strings.ToLower(message.Role) == "user" {
			return extractTextFromClaudeMessage(message)
		}
	}

	return ""
}

// extractTextFromClaudeMessage 从Claude消息中提取文本内容
func extractTextFromClaudeMessage(message dto.ClaudeMessage) string {
	var textParts []string

	// 检查消息内容是否为字符串格式
	if message.IsStringContent() {
		content := message.GetStringContent()
		if content != "" {
			textParts = append(textParts, content)
		}
	} else {
		// 解析复合内容
		parsedContent, err := message.ParseContent()
		if err == nil {
			for _, content := range parsedContent {
				if content.Type == "text" && content.Text != nil {
					textParts = append(textParts, *content.Text)
				}
			}
		}
	}

	if len(textParts) == 0 {
		return ""
	}

	// 合并所有文本部分
	fullText := strings.Join(textParts, "\n")

	// 清理文本
	fullText = cleanUserContent(fullText)

	return strings.TrimSpace(fullText)
}

// ShouldSkipModeration 判断是否应该跳过内容审核
func ShouldSkipModeration(content string) bool {
	if content == "" {
		return true
	}

	// 跳过过短的内容（可能只是简单问候）
	if len(strings.TrimSpace(content)) < 3 {
		return true
	}

	// 跳过纯数字或符号
	if regexp.MustCompile(`^[\d\s\p{P}]*$`).MatchString(content) {
		return true
	}

	// 跳过常见的简单问候和测试内容
	lowerContent := strings.ToLower(strings.TrimSpace(content))
	skipPatterns := []string{
		"hello", "hi", "test", "测试", "你好", "hello world",
		"ping", "pong", "ok", "yes", "no", ".", "?", "！",
	}

	for _, pattern := range skipPatterns {
		if lowerContent == pattern {
			return true
		}
	}

	return false
}