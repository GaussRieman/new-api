# TokenScope L2 API 接口文档

**版本**: tokenscope-v0.2.0-l2
**日期**: 2026-06-03

---

## 1. 概述

TokenScope L2 API 提供深度诊断能力，基于 Raw Body 采样和 Context Parts 解析。所有 L2 端点（除 `/self/l2/recent`）需要管理员权限，因为包含原始请求内容。

### 路由前缀

```
/api/tokenscope
```

### 认证

| 端点类型 | 认证方式 |
|----------|----------|
| L2 管理端点 | `middleware.AdminAuth()` — 需要管理员权限 |
| L2 用户端点 | `middleware.UserAuth()` — 仅查看自己的近期采样 |

---

## 2. 管理端点

### 2.1 GET /api/tokenscope/l2/summary

获取 L2 深度诊断指标汇总（按模型分组）。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/l2/summary?start_timestamp=1717200000&end_timestamp=1719792000&model_name=claude-sonnet-4-6"
```

#### 响应示例

```json
{
  "success": true,
  "data": [
    {
      "sample_count": 150,
      "avg_system_tokens": 2500,
      "avg_history_tokens": 8000,
      "avg_tool_tokens": 1200,
      "avg_file_tokens": 500,
      "repeated_prefix_rate": 0.65,
      "cache_friendliness": 0.42,
      "cache_fulfillment_rate": 0.38
    }
  ]
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `sample_count` | int64 | 采样请求数（有 parsed context parts 的请求） |
| `avg_system_tokens` | float64 | 平均 system prompt token 数 |
| `avg_history_tokens` | float64 | 平均 history message token 数 |
| `avg_tool_tokens` | float64 | 平均 tool 定义/调用 token 数 |
| `avg_file_tokens` | float64 | 平均 file content token 数 |
| `repeated_prefix_rate` | float64 | 重复前缀率（0.0~1.0） |
| `cache_friendliness` | float64 | 缓存友好度（0.0~1.0） |
| `cache_fulfillment_rate` | float64 | 缓存兑现率（0.0~1.0） |

---

### 2.2 GET /api/tokenscope/l2/recent

获取近期采样请求列表（按时间倒序）。

#### 额外参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `limit` | int | 否 | 20 | 返回数量 |

#### 响应

返回 `RequestDebugPayload[]` 数组。

---

### 2.3 GET /api/tokenscope/l2/request/:request_id

获取单个请求的完整 debug 信息（raw body + context parts）。

#### 响应示例

```json
{
  "success": true,
  "data": {
    "payload": {
      "id": 1,
      "request_id": "abc-123",
      "model_name": "claude-sonnet-4-6",
      "request_body": "{...}",
      "is_stream": false,
      "relay_format": "claude",
      "parsed": true
    },
    "parts": [
      {
        "id": 1,
        "request_id": "abc-123",
        "part_type": "system",
        "part_name": "system_prompt",
        "content_hash": "a1b2c3d4e5f6...",
        "token_count": 2500,
        "position": 0,
        "is_prefix": true,
        "is_repeated": true,
        "is_stable": true,
        "is_cache_friendly": true
      },
      {
        "id": 2,
        "request_id": "abc-123",
        "part_type": "tool",
        "part_name": "tool_definitions",
        "content_hash": "f6e5d4c3b2a1...",
        "token_count": 1200,
        "position": 1,
        "is_prefix": true,
        "is_repeated": true,
        "is_stable": true,
        "is_cache_friendly": true
      },
      {
        "id": 3,
        "request_id": "abc-123",
        "part_type": "history",
        "part_name": "user_message",
        "content_hash": "1234567890ab...",
        "token_count": 500,
        "position": 2,
        "is_prefix": false,
        "is_repeated": false,
        "is_stable": false,
        "is_cache_friendly": false
      }
    ]
  }
}
```

---

## 3. 用户端点

### 3.1 GET /api/tokenscope/self/l2/recent

当前用户近期的采样请求列表。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/self/l2/recent?limit=10"
```

#### 响应格式

返回 `RequestDebugPayload[]` 数组，仅包含当前用户的数据。

---

## 4. RequestDebugPayload 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 主键 |
| `request_id` | string | 请求唯一标识 |
| `log_id` | int | 关联的 logs 表 ID |
| `user_id` | int | 用户 ID |
| `model_name` | string | 模型名称 |
| `channel_id` | int | 渠道 ID |
| `token_id` | int | 令牌 ID |
| `group` | string | 分组 |
| `created_at` | int64 | 创建时间戳 |
| `request_body` | string | 原始请求体（JSON，可能截断） |
| `is_stream` | bool | 是否流式请求 |
| `relay_format` | string | 中继格式（openai/claude/gemini） |
| `parsed` | bool | 是否已解析为 context parts |
| `sampling_reason` | string | 采样原因 |

---

## 5. RequestContextPart 结构

| 字段 | 类型 | 说明 |
|------|------|------|
| `id` | int | 主键 |
| `request_id` | string | 关联的请求 ID |
| `part_type` | string | 部分类型：system, history, tool, file, memory, user |
| `part_name` | string | 部分名称 |
| `content_hash` | string | 内容 SHA-256 前缀（32 字符），用于去重检测 |
| `token_count` | int | 估算 Token 数（4 chars/token） |
| `position` | int | 在请求中的位置 |
| `is_prefix` | bool | 是否为前缀部分（system prompt、tool definitions 等） |
| `is_repeated` | bool | 是否在多次请求中重复出现 |
| `is_stable` | bool | 是否稳定（内容不变） |
| `is_cache_friendly` | bool | 是否缓存友好（prefix + stable） |
| `created_at` | int64 | 创建时间戳 |

---

## 6. 采样配置

L2 采样由 `tokenscope_setting` 控制，在系统设置中可配置：

| 配置项 | 类型 | 默认值 | 说明 |
|--------|------|--------|------|
| `enabled` | bool | false | 是否启用 L2 采样 |
| `sample_rate` | float64 | 0.01 | 采样率（0.0~1.0） |
| `max_payload_size` | int | 102400 | 最大请求体存储大小（100KB） |
| `retention_days` | int | 7 | 自动清理天数 |
| `capture_models` | string | "" | 模型白名单（逗号分隔，空=全部） |
