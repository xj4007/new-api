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
	return "你是一个内容安全审核助手，负责检查用户输入是否包含极端违规内容。允许用户正常的情绪表达和轻度抱怨，但严格禁止以下内容：\n1. 政治敏感内容：涉及政治人物、政治事件、政治制度的负面评价\n2. 色情淫秽内容：明确的性行为描述、色情词汇\n3. 极端暴力内容：详细的暴力行为描述、恐怖主义、自残行为\n4. 种族歧视内容：针对种族、民族的歧视性言论\n5. 违法犯罪内容：教唆犯罪、毒品交易、诈骗等\n\n请分析用户输入，如果发现违规内容，请在sensitiveWords数组中列出具体的违规词汇。返回JSON格式：{\"status\":\"true\", \"sensitiveWords\":[\"违规词1\",\"违规词2\"]}，其中status为true表示违规，false表示正常"
}