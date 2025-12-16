package setting

import (
	"strconv"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// 内容审核配置
var (
	// ContentModerationEnabled 是否启用AI内容审核
	ContentModerationEnabled = common.GetEnvOrDefaultBool("CONTENT_MODERATION_ENABLED", false)

	// ModerationApiBaseUrl 审核API基础地址
	ModerationApiBaseUrl = common.GetEnvOrDefaultString("MODERATION_API_BASE_URL", "https://api.siliconflow.cn")

	// ModerationApiKey 审核API密钥
	ModerationApiKey = common.GetEnvOrDefaultString("MODERATION_API_KEY", "")

	// ModerationModel 审核模型名称 (Phase 1 快速检测)
	ModerationModel = common.GetEnvOrDefaultString("MODERATION_MODEL", "Qwen/Qwen3-Next-80B-A3B-Instruct")

	// ModerationMaxTokens 审核响应最大Token数
	ModerationMaxTokens = common.GetEnvOrDefault("MODERATION_MAX_TOKENS", 100)

	// ModerationTimeout API调用超时时间（秒）
	ModerationTimeout = time.Duration(common.GetEnvOrDefault("MODERATION_TIMEOUT", 10)) * time.Second

	// ========== 二级审核配置 (Phase 2 Pro模型复核) ==========

	// ModerationProModel Pro模型名称，用于Phase 2复核
	ModerationProModel = common.GetEnvOrDefaultString("MODERATION_PRO_MODEL", "")

	// ModerationProApiKey Pro模型API密钥
	ModerationProApiKey = common.GetEnvOrDefaultString("MODERATION_PRO_API_KEY", "")

	// ModerationProApiBaseUrl Pro模型API基础地址
	ModerationProApiBaseUrl = common.GetEnvOrDefaultString("MODERATION_PRO_API_BASE_URL", "")

	// ========== 渠道过滤配置 ==========

	// ModerationChannelIds 需要审核的渠道ID列表 (逗号分隔，空则审核所有)
	ModerationChannelIdsStr = common.GetEnvOrDefaultString("MODERATION_CHANNEL_IDS", "")

	// ModerationChannelIdSet 渠道ID集合 (用于快速查询)
	ModerationChannelIdSet = parseChannelIds(ModerationChannelIdsStr)

	// ========== Session缓存配置 ==========

	// ModerationSessionCacheEnabled 是否启用Session级缓存
	ModerationSessionCacheEnabled = common.GetEnvOrDefaultBool("MODERATION_SESSION_CACHE_ENABLED", true)

	// ModerationSessionCacheTTL Session缓存过期时间（秒）
	ModerationSessionCacheTTL = common.GetEnvOrDefault("MODERATION_SESSION_CACHE_TTL", 1800)

	// ========== 内容提取配置 ==========

	// ModerationContentMaxChars 每个消息/系统提示词最大提取字符数
	ModerationContentMaxChars = common.GetEnvOrDefault("MODERATION_CONTENT_MAX_CHARS", 100)

	// ========== 重试和失败策略配置 ==========

	// ModerationMaxRetries 审核API最大重试次数（失败N次后放行）
	ModerationMaxRetries = common.GetEnvOrDefault("MODERATION_MAX_RETRIES", 3)

	// ModerationFailOpen 审核失败时是否放行 (true=放行, false=降级到敏感词检测)
	ModerationFailOpen = common.GetEnvOrDefaultBool("MODERATION_FAIL_OPEN", true)
)

// parseChannelIds 解析渠道ID字符串为集合
func parseChannelIds(idsStr string) map[int]bool {
	result := make(map[int]bool)
	if idsStr == "" {
		return result
	}

	ids := strings.Split(idsStr, ",")
	for _, idStr := range ids {
		idStr = strings.TrimSpace(idStr)
		if id, err := strconv.Atoi(idStr); err == nil {
			result[id] = true
		}
	}
	return result
}

// ShouldCheckContentModeration 是否应该进行内容审核
func ShouldCheckContentModeration() bool {
	return ContentModerationEnabled && ModerationApiKey != ""
}

// ShouldModerateChannel 检查是否需要对该渠道进行审核
func ShouldModerateChannel(channelId int) bool {
	// 如果未配置渠道列表，则审核所有渠道
	if len(ModerationChannelIdSet) == 0 {
		return true
	}
	// 检查渠道ID是否在列表中
	return ModerationChannelIdSet[channelId]
}

// HasProModel 是否配置了Pro模型（用于二级审核）
func HasProModel() bool {
	return ModerationProModel != "" && ModerationProApiKey != ""
}

// GetProApiBaseUrl 获取Pro模型API基础地址，如果未配置则使用默认地址
func GetProApiBaseUrl() string {
	if ModerationProApiBaseUrl != "" {
		return ModerationProApiBaseUrl
	}
	return ModerationApiBaseUrl
}

// GetSessionCacheTTL 获取Session缓存过期时间
func GetSessionCacheTTL() time.Duration {
	return time.Duration(ModerationSessionCacheTTL) * time.Second
}

// GetModerationSystemPrompt 获取审核系统提示词
func GetModerationSystemPrompt() string {
	return `你是一个内容安全审核员，负责检测用户输入中的不当内容。

## 审核规则

### 必须阻止的内容 (返回 {"status":true, "sensitiveWords":["具体违规词"]}):
1. **色情/性暗示内容**：包括但不限于
   - 直接的色情描述或请求
   - 性暗示词汇：涩涩、瑟瑟、色色、开车、黄色、18禁、NSFW、擦边、福利姬等
   - 任何关于色情内容的请求，即使是委婉的表达
2. **暴力内容**：煽动暴力、攻击特定个人或群体
3. **违法内容**：制作武器、毒品、诈骗等
4. **仇恨言论**：种族歧视、性别歧视等

### 允许通过的内容 (返回 {"status":false, "sensitiveWords":[]}):
- 代码片段和编程讨论
- 技术问题和架构讨论
- 错误日志和调试信息
- 正常的工作和学习内容
- 安全研究（教育目的）

## 重要提示
- 用户可能会使用谐音、变体、emoji来绕过检测，请保持警惕
- 只返回JSON格式，不要有其他文字

返回格式: {"status": boolean, "sensitiveWords": string[]}`
}
