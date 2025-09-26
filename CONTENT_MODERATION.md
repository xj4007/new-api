# 内容审核功能配置示例

## 环境变量配置

在 `.env` 文件中添加以下配置：

```env
# 内容审核配置
CONTENT_MODERATION_ENABLED=true
MODERATION_API_BASE_URL=https://api.siliconflow.cn
MODERATION_API_KEY=sk-your-api-key-here
MODERATION_MODEL=Qwen/Qwen3-Next-80B-A3B-Instruct
MODERATION_MAX_TOKENS=100
MODERATION_TIMEOUT=10
```

## 功能说明

### 支持的API格式
1. **Claude格式**: `/v1/messages`
2. **OpenAI格式**: `/v1/chat/completions`

### 审核请求格式

系统会向审核API发送如下格式的请求：

```json
{
  "top_p": 0.7,
  "model": "Qwen/Qwen3-Next-80B-A3B-Instruct",
  "messages": [
    {
      "content": "你是一个内容安全审核助手，负责检查用户输入是否包含极端违规内容。允许用户正常的情绪表达和轻度抱怨，但严格禁止以下内容：\n1. 政治敏感内容：涉及政治人物、政治事件、政治制度的负面评价\n2. 色情淫秽内容：明确的性行为描述、色情词汇\n3. 极端暴力内容：详细的暴力行为描述、恐怖主义、自残行为\n4. 种族歧视内容：针对种族、民族的歧视性言论\n5. 违法犯罪内容：教唆犯罪、毒品交易、诈骗等\n\n请分析用户输入，如果发现违规内容，请在sensitiveWords数组中列出具体的违规词汇。返回JSON格式：{\"status\":\"true\", \"sensitiveWords\":[\"违规词1\",\"违规词2\"]}，其中status为true表示违规，false表示正常",
      "role": "system"
    },
    {
      "content": "用户的实际输入内容",
      "role": "user"
    }
  ],
  "response_format": {
    "type": "json_object"
  },
  "enable_thinking": false
}
```

### 预期的API响应格式

```json
{
  "id": "ad386958a8664d0dbbb407b41ef24337",
  "object": "chat.completion",
  "created": 1758851853,
  "model": "Qwen/Qwen3-Next-80B-A3B-Instruct",
  "choices": [
    {
      "index": 0,
      "message": {
        "role": "assistant",
        "content": "{\"status\":\"false\", \"sensitiveWords\":[]}"
      },
      "finish_reason": "stop"
    }
  ]
}
```

### 审核流程
1. 提取用户输入的最后一条消息内容
2. 调用配置的审核模型进行内容检测
3. 解析返回的违规词汇列表
4. 如果API调用失败，自动降级到原有敏感词检测
5. 返回具体的违规词汇和修改建议

### 错误响应格式

#### Claude格式错误响应
```json
{
  "type": "error",
  "error": {
    "type": "invalid_request_error",
    "message": "检测到违规词汇：[操、废物]，请修改后重试。在Claude Code中按ESC+ESC可返回上次输入进行修改。"
  }
}
```

#### OpenAI格式错误响应
```json
{
  "error": {
    "message": "检测到违规词汇：[操、废物]，请修改后重试。在Claude Code中按ESC+ESC可返回上次输入进行修改。",
    "type": "invalid_request_error",
    "param": null,
    "code": "content_policy_violation"
  }
}
```

### 功能特性
- ✅ 双重保护：AI审核 + 敏感词检测降级
- ✅ 格式兼容：支持Claude和OpenAI两种请求格式
- ✅ 智能跳过：自动跳过简单问候和测试内容
- ✅ 具体提示：返回具体的违规词汇列表
- ✅ 用户友好：提供ESC+ESC快捷键提示
- ✅ 高度可配置：域名、模型、超时等均可配置
- ✅ 容错机制：API失败时自动降级
- ✅ 情绪宣泄：允许用户正常情绪表达，仅禁止极端内容

## 测试建议

### 1. 基础功能测试
```bash
# 测试正常内容
curl -X POST http://localhost:3000/v1/chat/completions \
  -H "Authorization: Bearer your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "Hello, how are you?"}]
  }'

# 测试情绪宣泄内容（应该通过）
curl -X POST http://localhost:3000/v1/messages \
  -H "x-api-key: your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "今天工作真累，压力好大"}]
  }'

# 测试违规内容（用测试词汇）
curl -X POST http://localhost:3000/v1/messages \
  -H "x-api-key: your-token" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "claude-3-sonnet",
    "messages": [{"role": "user", "content": "test_sensitive"}]
  }'
```

### 2. 配置验证
- 确保 `MODERATION_API_KEY` 配置正确
- 测试不同的 `MODERATION_MODEL` 设置（默认：Qwen/Qwen3-Next-80B-A3B-Instruct）
- 验证超时机制是否正常工作

### 3. 降级测试
- 临时设置错误的API密钥，确保降级到敏感词检测
- 测试网络超时情况下的处理

## 审核策略说明

### 允许的内容
- ✅ 正常情绪表达：抱怨、不满、发泄情绪
- ✅ 温和批评：对事物的负面评价（非极端）
- ✅ 日常用语：包含轻微俚语的正常表达

### 严格禁止的内容
- ❌ 政治敏感：涉及政治人物、制度的负面评价
- ❌ 色情淫秽：明确的性行为描述、色情词汇
- ❌ 极端暴力：详细暴力描述、恐怖主义、自残
- ❌ 种族歧视：针对种族、民族的歧视性言论
- ❌ 违法犯罪：教唆犯罪、毒品、诈骗等内容

## 注意事项

1. **性能影响**：每次审核会增加一次额外的API调用延迟（平均1-3秒）
2. **成本考虑**：使用外部API会产生费用，根据使用量计费
3. **隐私保护**：用户内容会发送到审核API，请确保符合隐私政策
4. **日志记录**：违规内容和具体词汇会记录在日志中，注意日志安全管理

## 故障排查

### 常见问题
1. **API密钥无效**：检查 `MODERATION_API_KEY` 配置
2. **网络超时**：调整 `MODERATION_TIMEOUT` 设置
3. **模型不存在**：确认 `MODERATION_MODEL` 名称正确
4. **JSON解析失败**：可能是模型返回格式不标准，会自动降级处理
5. **误判问题**：可以通过修改提示词来调整审核策略

### 日志查看
```bash
# 查看审核相关日志
docker-compose logs new-api | grep "content moderation"

# 查看违规检测日志
docker-compose logs new-api | grep "violation detected"
```

## 高级配置

### 自定义审核模型
可以切换到其他支持JSON格式输出的模型，如：
- `Pro/THUDM/glm-4-9b-chat`
- `deepseek-ai/DeepSeek-V2.5`
- `01-ai/Yi-Large`

### 自定义提示词
如需调整审核策略，可以修改 `setting/content-moderation.go` 中的 `GetModerationSystemPrompt()` 函数。