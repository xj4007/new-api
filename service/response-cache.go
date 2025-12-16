package service

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// ResponseCachePrefix Redis key prefix for response cache
const ResponseCachePrefix = "response_cache:"

// DefaultCacheTTL 默认缓存过期时间 180秒（3分钟）
const DefaultCacheTTL = 180 * time.Second

// MaxCacheSize 最大缓存大小 5MB
const MaxCacheSize = 5 * 1024 * 1024

// CachedResponse 缓存的响应数据结构
type CachedResponse struct {
	StatusCode int               `json:"status_code"`
	Headers    map[string]string `json:"headers"`
	Body       []byte            `json:"body"`
	Usage      json.RawMessage   `json:"usage,omitempty"`
	CachedAt   int64             `json:"cached_at"`
}

// GenerateCacheKey 生成缓存键
// sessionHash: 会话哈希（从 metadata.user_id 提取或使用 sessionHelper）
// requestBodyHash: 请求体的哈希值
func GenerateCacheKey(sessionHash string, requestBodyHash string) string {
	if sessionHash == "" || requestBodyHash == "" {
		return ""
	}

	// 组合 sessionHash 和 requestBodyHash 生成最终缓存键
	combined := sessionHash + ":" + requestBodyHash
	hash := sha256.Sum256([]byte(combined))
	return hex.EncodeToString(hash[:])
}

// GenerateRequestBodyHash 生成请求体哈希
// 对整个请求体（排除 stream 字段）进行哈希
func GenerateRequestBodyHash(requestBody []byte) string {
	if len(requestBody) == 0 {
		return ""
	}

	// 解析请求体并排除 stream 字段
	var requestMap map[string]interface{}
	if err := json.Unmarshal(requestBody, &requestMap); err != nil {
		// 如果解析失败，直接对原始内容哈希
		hash := sha256.Sum256(requestBody)
		return hex.EncodeToString(hash[:])
	}

	// 删除 stream 字段，因为它不影响响应内容
	delete(requestMap, "stream")

	// 重新序列化为规范化的 JSON
	normalizedBody, err := json.Marshal(requestMap)
	if err != nil {
		// 失败时使用原始内容
		hash := sha256.Sum256(requestBody)
		return hex.EncodeToString(hash[:])
	}

	hash := sha256.Sum256(normalizedBody)
	return hex.EncodeToString(hash[:])
}

// GetCachedResponse 从 Redis 获取缓存的响应
func GetCachedResponse(cacheKey string) (*CachedResponse, error) {
	if !common.RedisEnabled || cacheKey == "" {
		return nil, fmt.Errorf("redis not enabled or invalid cache key")
	}

	key := ResponseCachePrefix + cacheKey
	value, err := common.RedisGet(key)
	if err != nil {
		return nil, err
	}

	var cached CachedResponse
	if err := json.Unmarshal([]byte(value), &cached); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached response: %w", err)
	}

	return &cached, nil
}

// CacheResponse 缓存响应到 Redis
func CacheResponse(cacheKey string, response *CachedResponse, ttl time.Duration) error {
	if !common.RedisEnabled || cacheKey == "" {
		return nil
	}

	// 检查响应体大小
	if len(response.Body) > MaxCacheSize {
		common.SysLog(fmt.Sprintf("[ResponseCache] Response too large to cache: %d bytes > %d bytes", len(response.Body), MaxCacheSize))
		return fmt.Errorf("response too large to cache")
	}

	// 设置缓存时间戳
	response.CachedAt = time.Now().Unix()

	// 序列化响应
	data, err := json.Marshal(response)
	if err != nil {
		return fmt.Errorf("failed to marshal response: %w", err)
	}

	key := ResponseCachePrefix + cacheKey
	if err := common.RedisSet(key, string(data), ttl); err != nil {
		return fmt.Errorf("failed to set cache: %w", err)
	}

	// 记录缓存存储日志
	sizeKB := float64(len(response.Body)) / 1024.0
	common.SysLog(fmt.Sprintf("[ResponseCache] 💾 Cached response: %s | Size: %.2fKB | TTL: %ds",
		TruncateCacheKey(cacheKey), sizeKB, int(ttl.Seconds())))

	return nil
}

// TruncateCacheKey 截断缓存键用于日志显示
func TruncateCacheKey(cacheKey string) string {
	if len(cacheKey) <= 16 {
		return cacheKey
	}
	return cacheKey[:8] + "..." + cacheKey[len(cacheKey)-8:]
}
