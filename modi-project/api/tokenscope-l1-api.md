# TokenScope L1 API 接口文档

**版本**: tokenscope-v0.1.0-l1
**日期**: 2026-06-03

---

## 1. 概述

TokenScope L1 API 提供 Token 使用效率的聚合查询能力。所有端点基于现有 `logs` 表数据计算，无需额外配置即可使用（新产生的日志会包含缓存 Token 列，旧日志默认为 0）。

### 路由前缀

```
/api/tokenscope
```

### 认证

| 端点类型 | 认证方式 |
|----------|----------|
| Admin 端点 | `middleware.AdminAuth()` — 需要管理员权限 |
| User 端点 | `middleware.UserAuth()` — 需要登录，只能查看自己的数据 |

---

## 2. 通用查询参数

以下参数在所有 L1 端点中通用：

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| `start_timestamp` | int64 | 否 | 起始时间戳（Unix 秒），默认最近 7 天 |
| `end_timestamp` | int64 | 否 | 结束时间戳（Unix 秒），默认当前时间 |
| `model_name` | string | 否 | 模型名称过滤，精确匹配 |
| `group` | string | 否 | 分组过滤，精确匹配 |
| `channel` | int | 否 | 渠道 ID 过滤（仅 Admin 端点） |
| `username` | string | 否 | 用户名过滤（仅 Admin 端点） |

---

## 3. Admin 端点

### 3.1 GET /api/tokenscope/l1/summary

获取所有模型的汇总 L1 指标。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/l1/summary?start_timestamp=1717200000&end_timestamp=1719792000"
```

#### 响应示例

```json
{
  "success": true,
  "data": {
    "model_name": "(all)",
    "request_count": 15000,
    "output_cost": 0.0085,
    "context_load": 12.3,
    "cache_reuse_rate": 0.42,
    "total_quota": 8500000,
    "total_prompt_tokens": 45000000,
    "total_output_tokens": 3650000,
    "total_cache_read": 18900000,
    "total_cache_write": 5000000,
    "total_input_tokens": 48000000
  }
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `model_name` | string | 固定为 `"(all)"` |
| `request_count` | int64 | 请求总数 |
| `output_cost` | float64 | 产出成本 = total_quota / total_output_tokens（配额单位/Token） |
| `context_load` | float64 | 上下文负载 = total_prompt_tokens / total_output_tokens |
| `cache_reuse_rate` | float64 | 缓存复用率 = total_cache_read / total_prompt_tokens（0.0~1.0） |
| `total_quota` | int64 | 总消耗配额 |
| `total_prompt_tokens` | int64 | 总输入 Token |
| `total_output_tokens` | int64 | 总输出 Token |
| `total_cache_read` | int64 | 总缓存读取 Token |
| `total_cache_write` | int64 | 总缓存写入 Token |
| `total_input_tokens` | int64 | 总归一化输入 Token |

---

### 3.2 GET /api/tokenscope/l1/by-model

按模型分组的 L1 指标。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/l1/by-model?start_timestamp=1717200000&end_timestamp=1719792000&group=default"
```

#### 响应示例

```json
{
  "success": true,
  "data": [
    {
      "model_name": "claude-sonnet-4-6",
      "request_count": 8000,
      "output_cost": 0.0092,
      "context_load": 15.6,
      "cache_reuse_rate": 0.55,
      "total_quota": 4500000,
      "total_prompt_tokens": 30000000,
      "total_output_tokens": 1920000,
      "total_cache_read": 16500000,
      "total_cache_write": 3000000,
      "total_input_tokens": 32000000
    },
    {
      "model_name": "gpt-4o",
      "request_count": 5000,
      "output_cost": 0.0068,
      "context_load": 8.2,
      "cache_reuse_rate": 0.18,
      "total_quota": 2500000,
      "total_prompt_tokens": 10000000,
      "total_output_tokens": 1200000,
      "total_cache_read": 1800000,
      "total_cache_write": 1500000,
      "total_input_tokens": 11000000
    },
    {
      "model_name": "deepseek-chat",
      "request_count": 2000,
      "output_cost": 0.0012,
      "context_load": 20.5,
      "cache_reuse_rate": 0.0,
      "total_quota": 500000,
      "total_prompt_tokens": 5000000,
      "total_output_tokens": 530000,
      "total_cache_read": 0,
      "total_cache_write": 0,
      "total_input_tokens": 5000000
    }
  ]
}
```

---

### 3.3 GET /api/tokenscope/l1/timeseries

时序数据，用于折线图。

#### 额外参数

| 参数 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `bucket` | string | 否 | `"day"` | 分桶粒度：`"hour"` 或 `"day"` |

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/l1/timeseries?start_timestamp=1717200000&end_timestamp=1719792000&bucket=day&model_name=claude-sonnet-4-6"
```

#### 响应示例

```json
{
  "success": true,
  "data": [
    {
      "bucket": "2026-06-01",
      "request_count": 1200,
      "output_cost": 0.0095,
      "context_load": 14.8,
      "cache_reuse_rate": 0.52,
      "total_quota": 600000,
      "total_prompt_tokens": 4000000,
      "total_output_tokens": 270000,
      "total_cache_read": 2080000,
      "total_cache_write": 450000,
      "total_input_tokens": 4200000
    },
    {
      "bucket": "2026-06-02",
      "request_count": 1500,
      "output_cost": 0.0088,
      "context_load": 16.2,
      "cache_reuse_rate": 0.58,
      "total_quota": 750000,
      "total_prompt_tokens": 5200000,
      "total_output_tokens": 320000,
      "total_cache_read": 3016000,
      "total_cache_write": 520000,
      "total_input_tokens": 5500000
    }
  ]
}
```

#### 响应字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `bucket` | string | 时间桶标签（`"2026-06-01"` 或 `"2026-06-01 14:00"`） |
| 其余字段 | — | 同 summary 响应字段 |

---

## 4. User 端点

### 4.1 GET /api/tokenscope/self/l1/summary

当前用户的汇总 L1 指标。仅返回当前登录用户的数据。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/self/l1/summary?start_timestamp=1717200000&end_timestamp=1719792000"
```

#### 响应格式

同 Admin summary 响应，但数据范围限定为当前用户。

---

### 4.2 GET /api/tokenscope/self/l1/by-model

当前用户按模型分组的 L1 指标。

#### 请求示例

```bash
curl -H "Authorization: Bearer <token>" \
  "https://api.example.com/api/tokenscope/self/l1/by-model?start_timestamp=1717200000&end_timestamp=1719792000"
```

#### 响应格式

同 Admin by-model 响应，但数据范围限定为当前用户。

---

## 5. 错误响应

所有端点遵循项目统一的错误响应格式：

```json
{
  "success": false,
  "message": "错误描述"
}
```

### 常见错误

| HTTP 状态码 | 说明 |
|-------------|------|
| 401 | 未认证（缺少或无效的 Token） |
| 403 | 权限不足（非 Admin 访问 Admin 端点） |
| 500 | 服务器内部错误（数据库查询失败等） |

---

## 6. 数据说明

### 6.1 指标计算注意事项

1. **产出成本**使用聚合除法（`SUM(quota) / SUM(completion_tokens)`），不是单条记录比值的平均。聚合除法更能反映整体成本效率。

2. **缓存复用率**在旧数据上为 0，因为 `cache_read_tokens` 列在 v0.1.0 新增，历史数据该列为 0。这不是"缓存复用率为 0"，而是"数据缺失"。前端应区分这两种情况。

3. **除零保护**：所有除法在 Go 层计算，分母为 0 时返回 0。

### 6.2 配额单位

`quota` 和 `output_cost` 中的配额单位为系统内部单位（1 美元 = 500000 配额单位，可由 `common.QuotaPerUnit` 配置）。前端展示时应转换为用户可读的货币单位。

### 6.3 数据新鲜度

L1 指标基于 `logs` 表实时查询，无缓存延迟。数据反映的是当前数据库中最新的日志记录。

---

## 7. 未来版本 API 预留

v0.2.0 将新增以下端点（不在 v0.1.0 范围内）：

| 端点 | 方法 | 说明 |
|------|------|------|
| `/api/tokenscope/l2/summary` | GET | L2 深度诊断指标汇总 |
| `/api/tokenscope/l2/by-model` | GET | 按模型分组的 L2 指标 |
| `/api/tokenscope/l2/request/:request_id` | GET | 单个请求的 Context Parts 详情 |
| `/api/tokenscope/setting` | GET/PUT | L2 采样配置 |
