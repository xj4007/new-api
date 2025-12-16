package service

import (
	"regexp"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting"
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

// ========== 智能内容提取函数 ==========

// indexedMessage 带索引的消息，用于排序
type indexedMessage struct {
	Index   int
	Role    string
	Content string
}

// ExtractRecentClaudeMessages 提取Claude请求中最近的user和assistant消息
// maxUserMsgs: 最大user消息数（默认3）
// maxAssistantMsgs: 最大assistant消息数（默认3）
// 返回按原始顺序拼接的内容
func ExtractRecentClaudeMessages(claudeMessages []dto.ClaudeMessage, maxChars int, maxUserMsgs int, maxAssistantMsgs int) string {
	if maxChars <= 0 {
		maxChars = setting.ModerationContentMaxChars
	}
	if maxUserMsgs <= 0 {
		maxUserMsgs = 3
	}
	if maxAssistantMsgs <= 0 {
		maxAssistantMsgs = 3
	}

	var collected []indexedMessage
	userCount := 0
	assistantCount := 0

	// 从后向前遍历，收集最近的user和assistant消息
	for i := len(claudeMessages) - 1; i >= 0; i-- {
		message := claudeMessages[i]
		role := strings.ToLower(message.Role)

		if role == "user" && userCount < maxUserMsgs {
			content := extractClaudeMessageContentWithLimit(message, maxChars)
			if content != "" {
				collected = append(collected, indexedMessage{Index: i, Role: "User", Content: content})
				userCount++
			}
		} else if role == "assistant" && assistantCount < maxAssistantMsgs {
			content := extractClaudeMessageContentWithLimit(message, maxChars)
			if content != "" {
				collected = append(collected, indexedMessage{Index: i, Role: "Assistant", Content: content})
				assistantCount++
			}
		}

		// 如果两种消息都收集够了，提前退出
		if userCount >= maxUserMsgs && assistantCount >= maxAssistantMsgs {
			break
		}
	}

	if len(collected) == 0 {
		return ""
	}

	// 按原始索引排序（升序，恢复对话顺序）
	for i := 0; i < len(collected)-1; i++ {
		for j := i + 1; j < len(collected); j++ {
			if collected[i].Index > collected[j].Index {
				collected[i], collected[j] = collected[j], collected[i]
			}
		}
	}

	// 拼接内容
	var parts []string
	for _, msg := range collected {
		parts = append(parts, "["+msg.Role+"] "+msg.Content)
	}

	combined := strings.Join(parts, "\n")
	combined = cleanUserContent(combined)

	return strings.TrimSpace(combined)
}

// ExtractRecentOpenAIMessages 提取OpenAI请求中最近的user和assistant消息
func ExtractRecentOpenAIMessages(messages []dto.Message, maxChars int, maxUserMsgs int, maxAssistantMsgs int) string {
	if maxChars <= 0 {
		maxChars = setting.ModerationContentMaxChars
	}
	if maxUserMsgs <= 0 {
		maxUserMsgs = 3
	}
	if maxAssistantMsgs <= 0 {
		maxAssistantMsgs = 3
	}

	var collected []indexedMessage
	userCount := 0
	assistantCount := 0

	// 从后向前遍历，收集最近的user和assistant消息
	for i := len(messages) - 1; i >= 0; i-- {
		message := messages[i]
		role := strings.ToLower(message.Role)

		if role == "user" && userCount < maxUserMsgs {
			content := extractOpenAIMessageContentWithLimit(message, maxChars)
			if content != "" {
				collected = append(collected, indexedMessage{Index: i, Role: "User", Content: content})
				userCount++
			}
		} else if role == "assistant" && assistantCount < maxAssistantMsgs {
			content := extractOpenAIMessageContentWithLimit(message, maxChars)
			if content != "" {
				collected = append(collected, indexedMessage{Index: i, Role: "Assistant", Content: content})
				assistantCount++
			}
		}

		// 如果两种消息都收集够了，提前退出
		if userCount >= maxUserMsgs && assistantCount >= maxAssistantMsgs {
			break
		}
	}

	if len(collected) == 0 {
		return ""
	}

	// 按原始索引排序（升序，恢复对话顺序）
	for i := 0; i < len(collected)-1; i++ {
		for j := i + 1; j < len(collected); j++ {
			if collected[i].Index > collected[j].Index {
				collected[i], collected[j] = collected[j], collected[i]
			}
		}
	}

	// 拼接内容
	var parts []string
	for _, msg := range collected {
		parts = append(parts, "["+msg.Role+"] "+msg.Content)
	}

	combined := strings.Join(parts, "\n")
	combined = cleanUserContent(combined)

	return strings.TrimSpace(combined)
}

// ExtractClaudeContentSmart 智能提取Claude请求中的内容
// 提取最近3条user消息和最近3条assistant消息，按原始顺序拼接
// 不再提取系统提示词（system参数保留用于兼容但不使用）
func ExtractClaudeContentSmart(claudeMessages []dto.ClaudeMessage, system any, maxChars int) string {
	// 直接调用新的提取函数，不提取系统提示词
	return ExtractRecentClaudeMessages(claudeMessages, maxChars, 3, 3)
}

// extractClaudeSystemContent 提取Claude请求的系统提示词内容
func extractClaudeSystemContent(system any, maxChars int) string {
	if system == nil {
		return ""
	}

	var content string

	switch s := system.(type) {
	case string:
		content = s
	default:
		// 处理复合系统提示词 (ClaudeMediaMessage数组)
		mediaContent, _ := common.Any2Type[[]dto.ClaudeMediaMessage](system)
		var textParts []string
		for _, m := range mediaContent {
			if m.Type == "text" {
				text := m.GetText()
				if text != "" {
					// 每个系统提示词块也限制字符数
					if len(text) > maxChars {
						text = text[:maxChars] + "..."
					}
					textParts = append(textParts, text)
				}
			}
		}
		content = strings.Join(textParts, " ")
	}

	// 限制总字符数
	content = strings.TrimSpace(content)
	if len(content) > maxChars {
		content = content[:maxChars] + "..."
	}

	return content
}

// extractClaudeMessageContentWithLimit 从Claude消息中提取内容并限制字符数
func extractClaudeMessageContentWithLimit(message dto.ClaudeMessage, maxChars int) string {
	var content string

	if message.IsStringContent() {
		content = message.GetStringContent()
	} else {
		parsedContent, err := message.ParseContent()
		if err == nil {
			var textParts []string
			for _, c := range parsedContent {
				if c.Type == "text" && c.Text != nil {
					textParts = append(textParts, *c.Text)
				}
			}
			content = strings.Join(textParts, " ")
		}
	}

	// 先清理系统标签，再限制字符数
	content = cleanUserContent(content)
	content = strings.TrimSpace(content)
	if len(content) > maxChars {
		content = content[:maxChars] + "..."
	}

	return content
}

// ExtractOpenAIContentSmart 智能提取OpenAI请求中的内容
// 提取最近3条user消息和最近3条assistant消息，按原始顺序拼接
// 不再提取系统提示词
func ExtractOpenAIContentSmart(messages []dto.Message, maxChars int) string {
	// 直接调用新的提取函数，不提取系统提示词
	return ExtractRecentOpenAIMessages(messages, maxChars, 3, 3)
}

// extractOpenAIMessageContentWithLimit 从OpenAI消息中提取内容并限制字符数
func extractOpenAIMessageContentWithLimit(message dto.Message, maxChars int) string {
	var content string

	if message.IsStringContent() {
		content = message.StringContent()
	} else {
		parsedContent := message.ParseContent()
		var textParts []string
		for _, c := range parsedContent {
			if c.Type == "text" && c.Text != "" {
				textParts = append(textParts, c.Text)
			}
		}
		content = strings.Join(textParts, " ")
	}

	// 先清理系统标签，再限制字符数
	content = cleanUserContent(content)
	content = strings.TrimSpace(content)
	if len(content) > maxChars {
		content = content[:maxChars] + "..."
	}

	return content
}
