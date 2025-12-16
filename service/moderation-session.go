package service

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/dto"
	"github.com/QuantumNous/new-api/setting"
)

// ModerationSessionPrefix Redis key 前缀
const ModerationSessionPrefix = "moderation_session:"

// SessionPassThreshold 达到免审核所需的连续通过次数
const SessionPassThreshold = 3

// CheckSessionCache 检查session是否已达到免审核状态
// 只有连续通过次数 >= SessionPassThreshold 时才返回true
func CheckSessionCache(sessionId string) bool {
	if !common.RedisEnabled || sessionId == "" || !setting.ModerationSessionCacheEnabled {
		return false
	}

	key := ModerationSessionPrefix + sessionId
	value, err := common.RedisGet(key)
	if err != nil {
		return false // key不存在
	}

	// 解析通过次数
	passCount, err := strconv.Atoi(value)
	if err != nil {
		// 兼容旧格式（"passed"），视为已达标
		if value == "passed" {
			return true
		}
		return false
	}

	// 只有达到阈值才免审核
	return passCount >= SessionPassThreshold
}

// GetSessionPassCount 获取session当前的通过次数
func GetSessionPassCount(sessionId string) int {
	if !common.RedisEnabled || sessionId == "" {
		return 0
	}

	key := ModerationSessionPrefix + sessionId
	value, err := common.RedisGet(key)
	if err != nil {
		return 0
	}

	passCount, err := strconv.Atoi(value)
	if err != nil {
		return 0
	}

	return passCount
}

// IncrementSessionPassCount 增加session的通过次数
// 返回增加后的次数和是否达到免审核阈值
func IncrementSessionPassCount(sessionId string) (int, bool, error) {
	if !common.RedisEnabled || sessionId == "" || !setting.ModerationSessionCacheEnabled {
		return 0, false, nil
	}

	key := ModerationSessionPrefix + sessionId
	ttl := setting.GetSessionCacheTTL()

	// 获取当前次数
	currentCount := 0
	value, err := common.RedisGet(key)
	if err == nil {
		currentCount, _ = strconv.Atoi(value)
	}

	// 增加次数
	newCount := currentCount + 1

	// 存储新次数
	err = common.RedisSet(key, strconv.Itoa(newCount), ttl)
	if err != nil {
		return currentCount, false, err
	}

	// 返回新次数和是否达到阈值
	return newCount, newCount >= SessionPassThreshold, nil
}

// SetSessionCache 设置session审核通过缓存（兼容旧接口）
// 直接设置为已达标状态
func SetSessionCache(sessionId string) error {
	if !common.RedisEnabled || sessionId == "" || !setting.ModerationSessionCacheEnabled {
		return nil
	}

	key := ModerationSessionPrefix + sessionId
	ttl := setting.GetSessionCacheTTL()
	// 直接设置为阈值，表示已达标
	return common.RedisSet(key, strconv.Itoa(SessionPassThreshold), ttl)
}

// SetSessionCacheWithTTL 设置session审核通过缓存（自定义TTL）
func SetSessionCacheWithTTL(sessionId string, ttl time.Duration) error {
	if !common.RedisEnabled || sessionId == "" || !setting.ModerationSessionCacheEnabled {
		return nil
	}

	key := ModerationSessionPrefix + sessionId
	return common.RedisSet(key, "passed", ttl)
}

// DeleteSessionCache 删除session审核缓存
// 用于强制下次重新审核
func DeleteSessionCache(sessionId string) error {
	if !common.RedisEnabled || sessionId == "" {
		return nil
	}

	key := ModerationSessionPrefix + sessionId
	return common.RedisDel(key)
}

// ExtractSessionIdFromClaudeRequest 从Claude请求体中提取session ID
// Claude Code 发送的请求中 metadata.user_id 包含 session 信息
func ExtractSessionIdFromClaudeRequest(requestBody []byte) string {
	if len(requestBody) == 0 {
		return ""
	}

	var claudeRequest dto.ClaudeRequest
	if err := json.Unmarshal(requestBody, &claudeRequest); err != nil {
		return ""
	}

	return ExtractSessionIdFromClaudeRequestStruct(&claudeRequest)
}

// ExtractSessionIdFromClaudeRequestStruct 从ClaudeRequest结构体中提取session ID
func ExtractSessionIdFromClaudeRequestStruct(claudeRequest *dto.ClaudeRequest) string {
	if claudeRequest == nil || claudeRequest.Metadata == nil {
		return ""
	}

	var metadata dto.ClaudeMetadata
	if err := json.Unmarshal(claudeRequest.Metadata, &metadata); err != nil {
		return ""
	}

	return metadata.UserId
}

// GetSessionCacheKey 获取session缓存的Redis key
func GetSessionCacheKey(sessionId string) string {
	return ModerationSessionPrefix + sessionId
}

// TruncateSessionId 截断session ID用于日志显示
// 只显示前8个字符 + "..." 保护隐私
func TruncateSessionId(sessionId string) string {
	if len(sessionId) <= 8 {
		return sessionId
	}
	return sessionId[:8] + "..."
}
