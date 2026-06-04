# Gateway 层指标
## L1：运行监控指标
数据源：`request_log`
默认全量采集。
```text
1. 产出成本
= total_cost / output_tokens
2. 上下文负载
= input_tokens / output_tokens
3. 缓存复用率
= cache_read_tokens / input_tokens
```
作用：
```text
看输出贵不贵
看上下文重不重
看缓存有没有复用
```
---
## L2：深度诊断指标
数据源：`raw_request_body + context_parts`
Debug / 采样开启。
```text
1. 上下文结构
= 各类 context_part_tokens / input_tokens
2. 重复前缀率
= repeated_prefix_tokens / input_tokens
3. 缓存友好度
= stable_repeated_prefix_tokens / input_tokens
4. 缓存兑现率
= cache_read_tokens / theoretical_cache_friendly_tokens
```
作用：
```text
看 Token 花在哪里
看哪些前缀重复
看请求是否缓存友好
看 Provider 是否兑现缓存收益
```
---
# Gateway 数据结构
## 1. request_log
支撑 L1，默认采集。
```json
{
  "request_id": "uuid",
  "timestamp": "2026-06-02T10:00:00Z",
  "provider": "anthropic",
  "model": "claude-sonnet-4-6",
  "api_key_id": "key_001",
  "user_id": "user_001",
  "project_id": "project_001",
  "app_id": "modi-coding",
  "input_tokens": 1500,
  "output_tokens": 300,
  "cache_read_tokens": 1000,
  "cache_write_tokens": 500,
  "total_cost": 0.0111,
  "latency_ms": 4200,
  "status": "success"
}
```
---
## 2. request_debug_payload
支撑 L2，Debug / 采样开启。
```json
{
  "request_id": "uuid",
  "raw_request_body": {},
  "raw_response_body": {},
  "sampling_reason": "high_context_load"
}
```
---
## 3. request_context_parts
支撑 L2，由 Raw Body 解析生成。
```json
{
  "request_id": "uuid",
  "part_id": "uuid",
  "part_type": "tool_definition",
  "part_name": "case_search_tool",
  "content_hash": "sha256_xxx",
  "token_count": 800,
  "position": 3,
  "is_prefix": true,
  "is_repeated": true,
  "is_stable": true,
  "is_cache_friendly": true
}
```
---
# 最小关系
```text
request_log
  ├─ 0/1 request_debug_payload
  └─ 0/N request_context_parts
```
# 一句话
```text
L1 用 request_log 看问题；
L2 用 Raw Body + Context Parts 解释问题。
```