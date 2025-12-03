# Responses API URL 路径优化

## 问题背景

当使用第三方代理服务（非官方 OpenAI）转发 `/v1/responses` 请求时，系统会自动在 base_url 后拼接 `/v1/responses` 路径，导致最终 URL 包含 `/v1` 前缀。

例如：
- base_url: `http://154.219.123.250:3000/openai`
- 修改前: `http://154.219.123.250:3000/openai/v1/responses`
- 修改后: `http://154.219.123.250:3000/openai/responses`

部分第三方代理服务的 Responses API 路径不包含 `/v1` 前缀，导致请求失败。

## 解决方案

在 `relay/channel/openai/adaptor.go` 的 `GetRequestURL` 函数中，针对 Responses API 请求做特殊处理：

- 如果 base_url **不包含** `api.openai.com`（即非官方 OpenAI 地址）
- 则自动将 `/v1/responses` 替换为 `/responses`

## 代码修改

**文件**: `relay/channel/openai/adaptor.go`

**位置**: `GetRequestURL` 函数的 `default` case 中

```go
// 对于 Responses API，检查是否需要去掉 /v1 前缀
// 如果 base_url 不是官方 OpenAI 地址，则去掉 /v1 前缀，使用 /responses
if info.RelayMode == relayconstant.RelayModeResponses {
    requestPath := info.RequestURLPath
    // 非官方 OpenAI 地址，去掉 /v1 前缀
    if !strings.Contains(info.ChannelBaseUrl, "api.openai.com") {
        requestPath = strings.Replace(requestPath, "/v1/responses", "/responses", 1)
    }
    return relaycommon.GetFullRequestURL(info.ChannelBaseUrl, requestPath, info.ChannelType), nil
}
```

## 影响范围

| 场景 | base_url | 最终路径 |
|------|----------|----------|
| 官方 OpenAI | `https://api.openai.com` | `/v1/responses` (不变) |
| 第三方代理 | `http://xxx.xxx.xxx/openai` | `/responses` (去掉 /v1) |

**其他 API 不受影响**：
- `/v1/chat/completions` 保持原样
- `/v1/embeddings` 等其他接口不受影响
- 仅针对 `/v1/responses` 请求生效

## 修改日期

2025-12-03
