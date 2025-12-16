# 内容审核功能文档 v2.0

## 概述

内容审核系统是一个NSFW（不安全内容）检测和拦截机制，用于保护服务免受不适当内容的影响。系统会在请求发送到上游API之前，对输入内容进行审查。

**v2.0 核心特性**：
- ✨ **Session级审核缓存**：同一session在30分钟内只校验一次
- ✨ **智能内容提取**：用户输入+系统提示词各截取100字符
- ✨ **二级审核机制**：Phase 1快速检测 + Phase 2 Pro模型复核
- ✨ **渠道ID过滤**：只对指定渠道进行审核
- ✨ **宽松技术审核**：技术相关内容宽松处理
- ✨ **重试机制**：审核请求失败时自动重试3次
- ✨ **失败放行**：所有重试失败后自动放行（可配置）

---

## 环境变量配置

在 `.env` 文件中添加以下配置：

```bash
# ========== 基础配置 ==========
CONTENT_MODERATION_ENABLED=true
MODERATION_API_BASE_URL=https://api.siliconflow.cn
MODERATION_API_KEY=sk-your-api-key-here
MODERATION_MODEL=Qwen/Qwen3-Next-80B-A3B-Instruct
MODERATION_MAX_TOKENS=100
MODERATION_TIMEOUT=10

# ========== 渠道过滤配置 ==========
# 只对这些渠道ID进行审核（逗号分隔），空则审核所有渠道
MODERATION_CHANNEL_IDS=1,2,3

# ========== 二级审核配置 (Phase 2 Pro模型复核) ==========
MODERATION_PRO_MODEL=deepseek-ai/DeepSeek-V3
MODERATION_PRO_API_KEY=sk-your-pro-api-key
MODERATION_PRO_API_BASE_URL=https://api.siliconflow.cn

# ========== Session缓存配置 ==========
MODERATION_SESSION_CACHE_ENABLED=true
MODERATION_SESSION_CACHE_TTL=1800

# ========== 内容提取配置 ==========
MODERATION_CONTENT_MAX_CHARS=100

# ========== 重试和失败策略配置 ==========
# 审核API最大重试次数（失败N次后根据策略处理）
MODERATION_MAX_RETRIES=3

# 审核失败时是否放行 (true=放行, false=降级到敏感词检测)
MODERATION_FAIL_OPEN=true
```

---

## 功能说明

### 支持的API格式
1. **Claude格式**: `/v1/messages`
2. **OpenAI格式**: `/v1/chat/completions`

### 1. Session级审核缓存

#### 工作原理

1. **Session识别**：从Claude请求中提取 `metadata.user_id` 作为会话UUID
2. **缓存检查**：检查Redis中是否已有该session的审核记录
3. **首次审核**：未缓存时，提取内容进行审核
4. **缓存结果**：审核通过后缓存30分钟

#### 数据流程

```
请求到达 → 提取sessionId → 检查Redis缓存
                              ↓
                    ┌─ 存在 → 直接放行 ✅
                    │
                    └─ 不存在 → 提取审核内容
                                        ↓
                                   调用审核API (Phase 1 → Phase 2)
                                        ↓
                              ┌─ 通过 → 缓存结果(TTL=30分钟) → 放行 ✅
                              └─ 违规 → 拒绝请求 ❌
```

#### Redis Key 设计

- **Key格式**: `moderation_session:{sessionId}`
- **Value**: `passed`
- **TTL**: 1800秒 (30分钟，可通过 `MODERATION_SESSION_CACHE_TTL` 配置)
- **示例**: `moderation_session:17cf0fd3-d51b-4b59-977d-b899dafb3022`

#### Redis 查询命令

```bash
# 查看所有审核session缓存的key
redis-cli KEYS "moderation_session:*"

# 查看某个session是否已审核
redis-cli EXISTS "moderation_session:17cf0fd3-d51b-4b59-977d-b899dafb3022"

# 查看某个session的剩余TTL（秒）
redis-cli TTL "moderation_session:17cf0fd3-d51b-4b59-977d-b899dafb3022"

# 手动删除某个session的审核缓存（强制下次重新审核）
redis-cli DEL "moderation_session:17cf0fd3-d51b-4b59-977d-b899dafb3022"

# 批量删除所有审核session缓存（谨慎使用）
redis-cli KEYS "moderation_session:*" | xargs redis-cli DEL

# 统计当前缓存的session数量
redis-cli KEYS "moderation_session:*" | wc -l
```

**注意**：如果Redis配置了密码，需要加上 `-a <password>` 参数。

### 2. 智能内容提取

#### 提取规则

| 内容类型 | 提取方式 | 说明 |
|---------|---------|------|
| 用户输入 | 每条user消息前100字符 | 可配置 `MODERATION_CONTENT_MAX_CHARS` |
| 系统提示词 | 每个先截取100字符，再合并 | 确保每个系统提示词都被检查 |

#### 审核内容格式

**单个系统提示词：**
```
[System] You are a coding assistant powered by Claude...

[User] 帮我写一个函数
```

**多个系统提示词和用户消息：**
```
[System] You are a coding assistant powered by Claude...
[User] 帮我写一个函数
[User] 这个函数需要实现排序
```

### 3. 二级审核机制

```
Session首次请求
    ↓
Phase 1: 默认模型快速检测
    ↓
  ┌─ 通过 (status=false) → 缓存结果 → 放行 ✅
  │
  └─ 违规 (status=true) → Phase 2: Pro模型复核
                              ↓
                    ┌─ 通过 → 误判纠正 → 缓存结果 → 放行 ✅
                    └─ 仍违规 → 记录违规日志 → 拒绝请求 ❌
```

### 4. 渠道ID过滤

通过 `MODERATION_CHANNEL_IDS` 配置需要审核的渠道ID列表：

```bash
# 只审核渠道ID为1、2、3的请求
MODERATION_CHANNEL_IDS=1,2,3

# 空值表示审核所有渠道
MODERATION_CHANNEL_IDS=
```

---

## 审核流程完整图

```
请求到达
    ↓
Distribute() → 确定渠道ID
    ↓
ContentModerationMiddleware()
    ├─ 检查是否启用审核 → 否 → 放行
    ├─ 检查渠道ID是否在列表中 → 否 → 放行
    ├─ 检查是否Claude请求 → 是 → 提取sessionId
    │                           ↓
    │                    检查Session缓存 → 存在 → 放行
    │                           ↓
    ├─ 智能内容提取 (每个消息/系统提示词100字符)
    │                           ↓
    ├─ Phase 1: 默认模型快速检测
    │   └─ 通过 → 设置Session缓存 → 放行
    │   └─ 违规 → Phase 2
    │                           ↓
    ├─ Phase 2: Pro模型复核
    │   └─ 通过 → 设置Session缓存 → 放行 (误报纠正)
    │   └─ 违规 → 返回错误
    ↓
继续处理请求
```

---

## 审核策略说明

### 允许的内容（宽松处理）
- ✅ 代码片段（任何编程语言）
- ✅ 技术讨论（软件、API、算法、调试）
- ✅ 错误信息、堆栈跟踪、日志输出
- ✅ 项目规划和架构讨论
- ✅ 安全研究和漏洞讨论（教育性质）
- ✅ 系统管理和DevOps内容
- ✅ 数据库查询和数据处理
- ✅ 任何与软件开发相关的内容

### 严格禁止的内容
- ❌ 纯色情/露骨性内容（无技术上下文）
- ❌ 直接煽动对特定个人的暴力
- ❌ 制造大规模杀伤性武器的详细说明
- ❌ 直接宣传恐怖主义的内容

---

## 错误响应格式

### Claude格式错误响应
```json
{
  "type": "error",
  "error": {
    "type": "invalid_request_error",
    "message": "检测到违规词汇：[xxx]，请修改后重试。在Claude Code中按ESC+ESC可返回上次输入进行修改。"
  }
}
```

### OpenAI格式错误响应
```json
{
  "error": {
    "message": "检测到违规词汇：[xxx]，请修改后重试。在Claude Code中按ESC+ESC可返回上次输入进行修改。",
    "type": "invalid_request_error",
    "param": null,
    "code": "content_policy_violation"
  }
}
```

---

## 功能特性

- ✅ **渠道ID过滤**：只对指定渠道进行审核
- ✅ **Session级缓存**：30分钟内同session只审核一次
- ✅ **二级审核**：Phase 1快速检测 + Phase 2 Pro模型复核
- ✅ **智能内容提取**：每个消息/系统提示词抽取100字符
- ✅ **双重保护**：AI审核 + 敏感词检测降级
- ✅ **格式兼容**：支持Claude和OpenAI两种请求格式
- ✅ **智能跳过**：自动跳过简单问候和测试内容
- ✅ **具体提示**：返回具体的违规词汇列表
- ✅ **用户友好**：提供ESC+ESC快捷键提示
- ✅ **高度可配置**：所有参数均可通过环境变量配置
- ✅ **宽松技术审核**：技术相关内容自动放行

---

## 日志示例

```
🔍 content moderation: session 17cf0fd3... first check
📝 content moderation Phase 1: starting fast detection
✅ content moderation [Qwen/Qwen3-Next-80B-A3B-Instruct]: content passed review
✅ content moderation Phase 1: passed
✅ content moderation: session 17cf0fd3... marked as passed (TTL: 1800s)

# 后续同一session请求
✅ content moderation: session 17cf0fd3... passed (cached)

# 渠道过滤
ℹ️ content moderation: skipped for channel 5 (not in moderation list)
```

---

## 故障排查

### 常见问题
1. **API密钥无效**：检查 `MODERATION_API_KEY` 配置
2. **网络超时**：调整 `MODERATION_TIMEOUT` 设置
3. **模型不存在**：确认 `MODERATION_MODEL` 名称正确
4. **JSON解析失败**：可能是模型返回格式不标准，会自动降级处理
5. **误判问题**：配置Pro模型启用二级审核进行复核
6. **Session缓存不生效**：检查Redis是否正常运行

### 日志查看
```bash
# 查看审核相关日志
docker-compose logs new-api | grep "content moderation"

# 查看违规检测日志
docker-compose logs new-api | grep "violation"

# 查看Session缓存日志
docker-compose logs new-api | grep "session"
```

---

## 常见问题

### Q: 为什么同一session只审核一次？

A: 为了性能优化：
- 减少审核API调用次数
- 降低延迟
- 节省成本
- 30分钟后session过期，会重新审核

### Q: 如果用户在session内切换到违规内容怎么办？

A: 这是权衡的结果：
- 30分钟内确实不会再次检测
- 但大多数用户不会这样做
- 如需更严格，可缩短 `MODERATION_SESSION_CACHE_TTL`

### Q: 没有sessionId时如何处理？

A: 自动回退到原有逻辑：
- 每次请求都进行审核
- 日志会显示相关信息

### Q: 如何禁用session缓存？

A: 设置环境变量：
```bash
MODERATION_SESSION_CACHE_ENABLED=false
```

### Q: 如何只审核特定渠道？

A: 设置环境变量（逗号分隔渠道ID）：
```bash
MODERATION_CHANNEL_IDS=1,2,3
```

---

## 注意事项

1. **性能影响**：首次审核会增加一次额外的API调用延迟（平均1-3秒），后续同session请求直接放行
2. **成本考虑**：使用外部API会产生费用，Session缓存可显著降低成本
3. **隐私保护**：用户内容会发送到审核API，请确保符合隐私政策
4. **日志记录**：违规内容和具体词汇会记录在日志中，注意日志安全管理
5. **Redis依赖**：Session缓存功能需要Redis支持，确保Redis正常运行

---

**最后更新**：2024-12
**版本**：2.0
