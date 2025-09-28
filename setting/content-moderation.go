package setting

import (
	"one-api/common"
	"time"
)

// 内容审核配置
var (
	// ContentModerationEnabled 是否启用AI内容审核
	ContentModerationEnabled = common.GetEnvOrDefaultBool("CONTENT_MODERATION_ENABLED", false)

	// ModerationApiBaseUrl 审核API基础地址
	ModerationApiBaseUrl = common.GetEnvOrDefaultString("MODERATION_API_BASE_URL", "https://api.siliconflow.cn")

	// ModerationApiKey 审核API密钥
	ModerationApiKey = common.GetEnvOrDefaultString("MODERATION_API_KEY", "")

	// ModerationModel 审核模型名称
	ModerationModel = common.GetEnvOrDefaultString("MODERATION_MODEL", "Qwen/Qwen3-Next-80B-A3B-Instruct")

	// ModerationMaxTokens 审核响应最大Token数
	ModerationMaxTokens = common.GetEnvOrDefault("MODERATION_MAX_TOKENS", 100)

	// ModerationTimeout API调用超时时间（秒）
	ModerationTimeout = time.Duration(common.GetEnvOrDefault("MODERATION_TIMEOUT", 10)) * time.Second
)

// ShouldCheckContentModeration 是否应该进行内容审核
func ShouldCheckContentModeration() bool {
	return ContentModerationEnabled && ModerationApiKey != ""
}

// GetModerationSystemPrompt 获取审核系统提示词
func GetModerationSystemPrompt() string {
	return "You are a content safety moderator. For technical inputs (code/project discussions), always return {\"status\":\"false\", \"sensitiveWords\":[]}. For other inputs, check for: 1) Political sensitivity (negative comments on politicians/events/systems) 2) Pornography (explicit sexual content) 3) Extreme violence (detailed violence/terrorism/self-harm) 4) Racial discrimination 5) Illegal activities (crime/drugs/fraud). Allow normal emotions and mild complaints. Return JSON with status (true=violation) and sensitiveWords array."
}