# 非流式响应缓存功能

## 📋 功能概述

为了解决**客户端超时断开但上游 API 已成功返回**导致的重复计费问题，实现了非流式响应缓存机制。该功能可以：
- ✅ 缓存所有非流式 Claude API 响应（有效期 3 分钟）
- ✅ 客户端重试时直接从缓存返回，**不扣费**
- ✅ 显著提升用户体验，节省 API 成本
- ✅ 自动处理会话隔离，确保用户数据安全

---

## 🎯 解决的问题

### 问题场景

```
第1次请求：
客户端 → 15秒超时断开
服务器 → 继续等待上游
上游   → 54秒返回结果 ✅（但客户端已经走了，结果被丢弃，并已扣费）

客户端重试（第2次请求）：
客户端 → 又发起同样的请求
服务器 → 又要等上游54秒 ❌（再次扣费！）
上游   → 返回结果（又扣了一次费）
```

**结果**：同一个请求扣了多次费用，用户体验差，成本高昂。

### 解决方案

```
第1次请求：
客户端 → 15秒超时断开
服务器 → 继续等待上游
上游   → 54秒返回结果 → 💾 缓存起来（TTL: 3分钟）→ 扣费1次

客户端重试（第2次请求）：
客户端 → 发起同样的请求
服务器 → 🚀 检测到缓存 → 直接返回！（秒回，不扣费）
```

**结果**：只扣费一次，重试时秒级响应，用户体验好，成本节省。

---

## 🔑 核心设计

### 1. 缓存策略

- **缓存对象**：所有非流式（`stream: false`）Claude API 响应
- **缓存时机**：上游成功返回后立即缓存
- **缓存时长**：180 秒（3 分钟）TTL 自动过期
- **缓存大小限制**：单个响应最大 5MB
- **存储位置**：Redis

### 2. 缓存键生成机制

```go
cacheKey = SHA256(sessionHash + ":" + requestBodyHash)
```

**组成部分**：

1. **sessionHash**：从 `metadata.user_id` 提取的会话标识
   - 确保不同会话的请求不会互相干扰
   - 同一会话的相同请求可以命中缓存

2. **requestBodyHash**：请求体的 SHA256 哈希（排除 `stream` 字段）
   - 包含：`model`, `messages`, `system`, `max_tokens`, `temperature`, `top_p`, `top_k`, `stop_sequences`
   - 排除：`metadata`, `stream`（这些不影响响应内容）

**特性**：
- ✅ 完全相同的请求 → 相同的缓存键 → 缓存命中
- ✅ 任何参数差异 → 不同的缓存键 → 独立缓存
- ✅ 用户隔离：不同用户/会话的缓存完全独立
- ✅ 子代理隔离：Tool 调用产生的子请求有不同的 `messages`，因此不会与主请求混淆

### 3. 缓存数据结构

```go
type CachedResponse struct {
    StatusCode int               `json:"status_code"` // HTTP 状态码
    Headers    map[string]string `json:"headers"`     // 响应头
    Body       []byte            `json:"body"`        // 完整响应体
    Usage      json.RawMessage   `json:"usage"`       // Token 使用信息
    CachedAt   int64             `json:"cached_at"`   // 缓存时间戳
}
```

**Redis 存储格式**：
```
键名：response_cache:{cacheKey}
类型：String (JSON)
TTL：180秒（3分钟）
```

---

## 🔧 实现细节

### 涉及文件（Git Commit: 8a39e40）

| 文件 | 类型 | 说明 |
|------|------|------|
| `service/response-cache.go` | **新建** | 响应缓存服务核心逻辑 |
| `relay/claude_handler.go` | **修改** | 请求入口添加缓存检查 |
| `relay/channel/claude/relay-claude.go` | **修改** | 响应处理添加缓存存储 |
| `service/http.go` | **修改** | `IOCopyBytesGracefully` 返回错误 |
| `router/relay-router.go` | **修改** | 中间件顺序调整 |

### 关键函数

#### 1. `service/response-cache.go` - 缓存服务

```go
// 生成缓存键
func GenerateCacheKey(sessionHash string, requestBodyHash string) string

// 生成请求体哈希
func GenerateRequestBodyHash(requestBody []byte) string

// 获取缓存的响应
func GetCachedResponse(cacheKey string) (*CachedResponse, error)

// 缓存响应
func CacheResponse(cacheKey string, response *CachedResponse, ttl time.Duration) error

// 截断缓存键用于日志显示
func TruncateCacheKey(cacheKey string) string
```

#### 2. `relay/claude_handler.go` - 缓存检查（请求前）

**位置**：第 138-179 行

**逻辑**：
1. 检查是否为非流式请求
2. 生成缓存键
3. 查询 Redis 缓存
4. 如果命中，直接返回缓存内容（**不请求上游，不扣费**）
5. 如果未命中，继续正常流程

```go
if !info.IsStream && common.RedisEnabled {
    // 生成缓存键
    cacheKey := service.GenerateCacheKey(sessionHash, requestBodyHash)

    // 检查缓存
    cached, err := service.GetCachedResponse(cacheKey)
    if err == nil && cached != nil {
        // 🎯 缓存命中！直接返回
        returnCachedResponseToClient(c, cached)
        return nil // 不扣费
    }
}
```

#### 3. `relay/channel/claude/relay-claude.go` - 缓存存储（响应后）

**位置**：第 795-837 行

**逻辑**：
1. 上游返回响应后
2. 写入客户端
3. **无论写入是否成功，都缓存响应**（TTL: 3分钟）

```go
// 💾 缓存所有非流式响应
if cacheKey, exists := c.Get("response_cache_key"); exists {
    cachedResp := &service.CachedResponse{
        StatusCode: httpResp.StatusCode,
        Headers:    headers,
        Body:       responseData,
        Usage:      usageJSON,
    }
    service.CacheResponse(cacheKeyStr, cachedResp, service.DefaultCacheTTL)
}
```

---

## 📊 日志标识

### 缓存检查阶段（`ClaudeHelper`）

```log
[ResponseCache] Checking cache for non-stream request
[ResponseCache] Request body size: 311358 bytes
[ResponseCache] Extracted sessionHash: user_161...
[ResponseCache] Generated requestBodyHash: 8e1ca84a...74f64c8b
[ResponseCache] Generated cacheKey: 18be76eb...06b300ad
[ResponseCache] Cache MISS - will proceed with upstream request
```

**或者缓存命中时：**
```log
[ResponseCache] 🎯 Cache HIT | Key: 18be76eb...06b300ad | SessionHash: user_161...
```

### 响应写入阶段（`ClaudeHandler`）

```log
[ClaudeHandler] About to write response, size: 16938 bytes
[ClaudeHandler] ✅ Response written successfully to client
[ClaudeHandler] 💾 Caching response with key: 18be76eb...06b300ad
[ClaudeHandler] ✅ Response cached successfully (TTL: 3min)
```

### 缓存存储记录（`response-cache.go`）

```log
[ResponseCache] 💾 Cached response: 18be76eb...06b300ad | Size: 16.54KB | TTL: 180s
```

---

## 🚀 使用效果

### 场景1：正常请求（第一次）

```
用户请求 → 缓存 MISS → 请求上游 → 上游返回 → 缓存响应 → 返回用户 → ✅ 扣费
耗时：~45秒
费用：正常扣费
```

### 场景2：重试请求（3分钟内）

```
用户重试 → 缓存 HIT → 直接返回 → ✅ 不扣费
耗时：<1秒
费用：0 元
```

### 实际数据对比

**优化前**（无缓存）：
- 5次重试 → 5次扣费（每次 $0.135）
- 总费用：$0.675
- 总耗时：~225秒

**优化后**（有缓存）：
- 第1次：扣费 $0.135，耗时 45秒
- 第2-5次：缓存命中，$0 费用，耗时 <1秒
- **总费用：$0.135**
- **总耗时：~49秒**
- **节省费用：80%**
- **节省时间：78%**

---

## ⚙️ 配置参数

### 环境变量

无需额外配置，自动使用现有的 Redis 连接。

### 代码常量

位置：`service/response-cache.go`

```go
const (
    ResponseCachePrefix = "response_cache:"  // Redis 键前缀
    DefaultCacheTTL     = 180 * time.Second  // 默认 TTL：3分钟
    MaxCacheSize        = 5 * 1024 * 1024    // 最大缓存大小：5MB
)
```

**修改 TTL**（如果需要）：
```go
// 改为 5 分钟
const DefaultCacheTTL = 300 * time.Second
```

---

## 🔍 Redis 调试命令

### 查看缓存键

```bash
# 查看所有缓存键
redis-cli KEYS "response_cache:*"

# 查看缓存数量
redis-cli KEYS "response_cache:*" | wc -l

# 查看某个缓存的详细信息
redis-cli GET "response_cache:{cacheKey}"

# 查看缓存剩余 TTL
redis-cli TTL "response_cache:{cacheKey}"
```

### 手动清理缓存

```bash
# 清除所有响应缓存
redis-cli --scan --pattern "response_cache:*" | xargs redis-cli DEL

# 清除特定缓存
redis-cli DEL "response_cache:{cacheKey}"
```

### 查看缓存大小

```bash
# 查看某个缓存的大小
redis-cli --raw GET "response_cache:{cacheKey}" | wc -c
```

---

## 🛡️ 安全性与隔离

### 用户隔离

- **sessionHash 包含用户标识**：不同用户的缓存完全隔离
- 即使两个用户发送完全相同的请求，也不会共享缓存

### 会话隔离

- **sessionHash 包含会话 ID**：同一用户的不同会话有不同的缓存
- Claude Code 的不同 session（不同任务）不会互相干扰

### 子代理隔离

- **requestBodyHash 包含完整 messages**：Tool 调用的子请求与主请求有不同的 messages
- 子代理请求自动生成不同的缓存键，不会与主请求混淆

---

## 📈 性能优化

### 缓存命中率优化

1. **增加 TTL**：如果用户重试间隔较长，可以适当增加 TTL
2. **监控缓存使用**：通过日志统计缓存命中率

```bash
# 统计缓存命中次数
grep "Cache HIT" logs/*.log | wc -l

# 统计缓存未命中次数
grep "Cache MISS" logs/*.log | wc -l
```

### 内存优化

- **自动过期**：TTL 到期后 Redis 自动删除
- **大小限制**：超过 5MB 的响应不缓存（避免占用过多内存）

---

## 🔧 故障排除

### 问题1：缓存未命中（重复请求仍扣费）

**症状**：
```log
[ResponseCache] Cache MISS - will proceed with upstream request
```
但确定是重复请求。

**可能原因**：
1. **请求参数有微小差异**（如 temperature、max_tokens）
2. **会话 ID 不同**（不同 session）
3. **缓存已过期**（超过 3 分钟）

**解决**：
```bash
# 检查缓存键是否存在
redis-cli KEYS "response_cache:*"

# 对比两次请求的 cacheKey 是否相同（查看日志）
grep "Generated cacheKey" logs/*.log | tail -5
```

### 问题2：Redis 连接失败

**症状**：
```log
[ResponseCache] Redis not enabled, skipping cache
```

**原因**：
- `REDIS_CONN_STRING` 环境变量未配置或配置错误

**解决**：
```bash
# 检查环境变量
docker-compose exec new-api env | grep REDIS

# 检查 Redis 连接
docker-compose exec redis redis-cli ping
```

### 问题3：缓存过大导致失败

**症状**：
```log
[ResponseCache] Response too large to cache: 6780000 bytes > 5242880 bytes
```

**原因**：
- 响应体超过 5MB 限制

**解决**：
- 增加 `MaxCacheSize` 常量（如果需要缓存更大的响应）

---

## 📝 Git Commit 信息

**Commit ID**: `8a39e40bbf6af8be0f8fbe319e3776a353954007`

**提交信息**:
```
fix: enhance response caching and error handling in Claude relay

- Added logic to cache responses when client disconnects during response write
- Implemented response cache retrieval in Claude handler to serve cached responses directly
- Improved response writing error handling to cache responses gracefully
- Refined IOCopyBytesGracefully to return errors for better error management
- Introduced response cache utility functions for generating cache keys
```

**修改文件**:
- ✅ `service/response-cache.go` (新建，135 行)
- ✅ `relay/claude_handler.go` (+67 行)
- ✅ `relay/channel/claude/relay-claude.go` (+52 行修改)
- ✅ `service/http.go` (+6 行修改)
- ✅ `router/relay-router.go` (+5 行修改)

---

## 🎯 未来扩展

### 可能的优化方向

1. **流式响应缓存**：目前只支持非流式，未来可以考虑缓存流式响应
2. **智能 TTL**：根据请求类型和内容动态调整 TTL
3. **缓存统计接口**：添加管理后台接口查看缓存使用情况
4. **LRU 淘汰策略**：当缓存数量过多时，自动淘汰最少使用的缓存

---

## 📚 相关文档

- [Redis 配置文档](../README.md#redis-配置)
- [Claude API 文档](https://docs.anthropic.com/claude/reference)
- [CLAUDE.md 项目说明](../../CLAUDE.md)

---

**最后更新**：2025-12-17
**作者**：fenglangyuan
**版本**：v1.0
