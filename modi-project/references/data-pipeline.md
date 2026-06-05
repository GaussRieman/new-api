# TokenScope 数据来龙去脉

## 全局架构

```
HTTP 请求
  │
  ├─① BodyStorage 缓存（内存/磁盘临时文件，请求结束即释放）
  │
  ├─② Relay 处理（转发请求到上游 Provider）
  │
  ├─③ 计费结算 → 写 logs 表 → L1 数据源
  │
  ├─④ L2 采样判断 → 截断 → 写 request_debug_payloads + request_context_parts
  │
  └─⑤ 响应返回 → BodyStorage.Close() 释放

前端查询
  │
  ├─ L1: SELECT ... FROM logs WHERE ... GROUP BY ...
  │
  └─ L2: SELECT ... FROM request_debug_payloads
           INNER JOIN request_context_parts ON request_id
           LEFT JOIN logs ON request_id
```

---

## L1：运行监控指标

### 数据源：`logs` 表

L1 没有专用表，直接从 Gateway 原有的消费日志表聚合。每条请求完成计费后写一条记录。

### 表结构

```sql
CREATE TABLE logs (
  id                INTEGER PRIMARY KEY,
  user_id           INTEGER,           -- 用户 ID
  created_at        INTEGER,           -- 请求时间戳（秒）
  type              INTEGER,           -- 日志类型（L1 只看 type=2 即消费记录）
  content           TEXT,              -- 日志内容
  username          TEXT DEFAULT '',   -- 用户名
  token_name        TEXT DEFAULT '',   -- API Token 名称
  model_name        TEXT DEFAULT '',   -- 模型名称
  quota             INTEGER DEFAULT 0,-- 消耗配额
  prompt_tokens     INTEGER DEFAULT 0,-- 输入 Token 数（不含缓存）
  completion_tokens INTEGER DEFAULT 0,-- 输出 Token 数
  use_time          INTEGER DEFAULT 0,-- 耗时（ms）
  is_stream         NUMERIC,           -- 是否流式
  channel_id        INTEGER,           -- 渠道 ID
  channel_name      TEXT,              -- 渠道名称
  token_id          INTEGER DEFAULT 0,-- API Token ID
  `group`           TEXT,              -- 分组
  ip                TEXT DEFAULT '',   -- 来源 IP
  request_id        VARCHAR(64),       -- 请求唯一标识
  upstream_request_id VARCHAR(128),    -- 上游请求 ID
  other             TEXT,              -- 扩展 JSON 字段
  cache_read_tokens   INTEGER DEFAULT 0, -- 缓存命中 Token（TokenScope 新增）
  cache_write_tokens  INTEGER DEFAULT 0, -- 缓存写入 Token（TokenScope 新增）
  input_tokens_total  INTEGER DEFAULT 0  -- 总输入 Token = prompt + cache_read（TokenScope 新增）
);
```

### 数据链路

```
请求进入 Relay
  → 转发到上游 Provider
  → 收到响应
  → PostConsumeQuota() 计费
  → RecordConsumeLog() 写 logs 表
  → 每条请求一条记录，全量写入，无采样
```

### L1 三项核心指标

| 指标 | 公式 | 回答的问题 | 正常范围 |
|------|------|-----------|---------|
| **产出成本** | `SUM(quota) / SUM(completion_tokens)` | 每输出 1 Token 花多少钱？ | 模型定价决定 |
| **上下文负载** | `SUM(prompt_tokens) / SUM(completion_tokens)` | 为了得到输出，喂了多少输入？ | 3-50，>100 说明上下文浪费 |
| **缓存复用率** | `SUM(cache_read_tokens) / SUM(input_tokens_total)` | 输入中有多少来自缓存？ | 0-1，越高越好 |

### 指标计算细节

产出成本和上下文负载用 `prompt_tokens` 而非 `input_tokens_total`，因为分母是 `completion_tokens`，分子应该是"实际发出的 prompt"而非"含缓存的总输入"。

缓存复用率的分母用 `input_tokens_total = prompt_tokens + cache_read_tokens`，代表"Provider 看到的全部输入"，分子是其中来自缓存的部分。

### SQL 聚合逻辑

```sql
-- 按 model 分组的 L1 指标
SELECT
  model_name AS name,
  COUNT(*) AS request_count,
  COALESCE(SUM(quota), 0) AS total_quota,
  COALESCE(SUM(prompt_tokens), 0) AS total_prompt_tokens,
  COALESCE(SUM(completion_tokens), 0) AS total_output_tokens,
  COALESCE(SUM(cache_read_tokens), 0) AS total_cache_read,
  COALESCE(SUM(cache_write_tokens), 0) AS total_cache_write,
  COALESCE(SUM(input_tokens_total), 0) AS total_input_tokens
FROM logs
WHERE type = 2                       -- 只看消费记录
  AND created_at >= :start_timestamp
  AND created_at <= :end_timestamp
  AND model_name != ''
GROUP BY model_name;
```

应用层计算派生指标（Go 代码 `aggToMetrics`）：

```go
OutputCost     = total_quota / total_output_tokens       // 产出成本
ContextLoad    = total_prompt_tokens / total_output_tokens // 上下文负载
CacheReuseRate = total_cache_read / (total_prompt_tokens + total_cache_read) // 缓存复用率
```

### 分组维度

L1 支持按 4 个维度分组，通过切换 `GROUP BY` 列实现：

| 维度 | GROUP BY 列 | name 列 | 说明 |
|------|------------|---------|------|
| 模型 | `model_name` | 模型名称 | 默认维度 |
| 用户 | `username` | 用户名 | 仅管理员 |
| API Key | `token_name` | Token 名称 | |
| 渠道 | `channel_id` | 渠道 ID → 名称 | 仅管理员，需二次查询 channel 表 |

### 时序查询

按天或小时分桶，支持 SQLite / MySQL / PostgreSQL 三种时间函数：

```sql
-- SQLite: 按天分桶
strftime('%Y-%m-%d', datetime(created_at, 'unixepoch')) AS bucket
-- MySQL: 按天分桶
DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d') AS bucket
-- PostgreSQL: 按天分桶
date_trunc('day', to_timestamp(created_at)) AS bucket
```

### 常见分析逻辑

1. **产出成本飙升**：按模型分组 → 找到哪个模型变贵了 → 可能是模型调价或配额计算变化
2. **上下文负载异常高**（>100）：说明大量输入只产出很少输出 → 检查是否有应用在发巨型 system prompt 但只问简单问题
3. **缓存复用率低**：对比同模型不同渠道 → 可能某个渠道不支持缓存，或缓存配置有误
4. **趋势分析**：时序查询看 7 天内各指标走势，识别渐进恶化

---

## L2：深度诊断指标

### 数据源：`request_debug_payloads` + `request_context_parts`

L2 需要原始请求体来分析"Token 花在哪里"，这是 `logs` 表不存的。通过采样捕获。

### 表结构

**`request_debug_payloads`** — 采样的原始请求体

```sql
CREATE TABLE request_debug_payloads (
  id              INTEGER PRIMARY KEY,
  request_id      VARCHAR(64) DEFAULT '',  -- 请求唯一标识（唯一索引）
  log_id          INTEGER DEFAULT 0,       -- 关联 logs 表 ID（当前未实现，永远为 0）
  user_id         INTEGER,                 -- 用户 ID
  model_name      TEXT DEFAULT '',         -- 模型名称
  channel_id      INTEGER DEFAULT 0,       -- 渠道 ID
  token_id        INTEGER DEFAULT 0,       -- API Token ID
  `group`         TEXT DEFAULT '',          -- 分组
  created_at      INTEGER,                 -- 采样时间戳
  request_body    TEXT,                     -- 原始请求体 JSON（最大 100KB，截断后闭合括号）
  is_stream       NUMERIC,                  -- 是否流式
  relay_format    TEXT DEFAULT '',          -- 中继格式（openai/claude/gemini）
  parsed          NUMERIC DEFAULT false,    -- 是否已解析为 context parts
  sampling_reason TEXT DEFAULT ''            -- 采样原因
);
```

**`request_context_parts`** — parser 解析出的上下文分段

```sql
CREATE TABLE request_context_parts (
  id               INTEGER PRIMARY KEY,
  request_id       VARCHAR(64) DEFAULT '',  -- 关联 payload
  part_type        TEXT DEFAULT '',          -- system / history / tool / file / memory / user
  part_name        TEXT DEFAULT '',          -- system_prompt / user_message / Bash / ...
  content_hash     TEXT DEFAULT '',          -- SHA-256 前缀（32字符），去重检测用
  token_count      INTEGER DEFAULT 0,        -- 估算 Token 数（4 字符/token）
  position         INTEGER DEFAULT 0,        -- 在请求中的位置顺序
  is_prefix        NUMERIC DEFAULT false,     -- 是否为前缀部分（system prompt、tool definitions）
  is_repeated      NUMERIC DEFAULT false,     -- 是否跨请求重复出现
  is_stable        NUMERIC DEFAULT false,     -- 是否内容不变
  is_cache_friendly NUMERIC DEFAULT false,    -- 是否缓存友好（prefix + stable）
  created_at       INTEGER                   -- 创建时间戳
);

-- 索引
CREATE INDEX idx_request_context_parts_request_id ON request_context_parts(request_id);
CREATE INDEX idx_request_context_parts_created_at ON request_context_parts(created_at);
CREATE INDEX idx_request_context_parts_content_hash ON request_context_parts(content_hash);
CREATE INDEX idx_request_context_parts_part_type ON request_context_parts(part_type);
```

### 关系

```
request_debug_payloads  1 : N  request_context_parts
        (1 条采样)              (10-30 条分段)

request_debug_payloads  N : 0..1  logs
      (request_id)                  (request_id 关联，LEFT JOIN，可能无匹配)
```

### 数据链路

```
① 请求进入 Relay
    │
    ├─ BodyStorage 缓存请求体（内存/磁盘，请求期间存活）
    │  common.CreateBodyStorageFromReader(c.Request.Body)
    │  → 小请求: memoryStorage ([]byte)
    │  → 大请求: diskStorage (临时文件)
    │
② Relay 处理请求，body 可被多次读取（解析、转发）
    │
③ 计费完成 → PostConsumeQuota()
    │
    ├─ 写 logs 表（L1 数据源，全量）
    │
    ├─ 从 BodyStorage 读取 body 字节（必须在 goroutine 外，因为 gin 会回收 context）
    │  l2Body, _ = common.GetBodyStorage(ctx).Bytes()
    │
    └─ goroutine 内: MaybeCaptureRequestBody(relayInfo, l2Body)
         │
         ├─ 采样判断
         │  ├─ SampleRate > 0?（设为 0 等效关闭）
         │  ├─ 模型白名单过滤
         │  └─ 混合采样策略
         │     ├─ 轮询保底：每 10 个请求采样 1 个（原子计数器）
         │     └─ 随机采样：基于 request_id 哈希，概率 = SampleRate
         │
         ├─ 截断处理（body > MaxPayloadSize 时）
         │  truncateAndCloseJSON(body, 102400)
         │  ├─ 正向扫描找到最后一个完整 value 闭合位置（} 或 ]）
         │  ├─ 统计未闭合的 { 和 [ 数量
         │  ├─ 补上闭合括号
         │  └─ 验证 json.Valid，不合法则回退
         │
         ├─ 写入 request_debug_payloads 表
         │
         ├─ 调用 parser.ParseAndStore()
         │  ├─ 根据 relay_format 选择解析器：
         │  │  ├─ openai → parseOpenAIFormat
         │  │  ├─ claude → parseClaudeFormat（兼容 OpenAI 格式中的 system 消息）
         │  │  └─ gemini → parseGeminiFormat
         │  ├─ 解析 messages 为分段（part_type + part_name + content_hash + token_count）
         │  ├─ analyzeParts() 后处理：
         │  │  ├─ part_type=system 或 part_name=tool_definitions → is_prefix=true
         │  │  ├─ is_prefix 且 content_hash 在多条 part 中出现 → is_repeated=true, is_stable=true
         │  │  └─ is_prefix + is_stable → is_cache_friendly=true
         │  └─ 写入 request_context_parts 表
         │
         └─ cleanupOldSamples(100)
            DELETE FROM request_debug_payloads
            WHERE request_id NOT IN (
              SELECT request_id FROM request_debug_payloads
              ORDER BY created_at DESC LIMIT 100
            )
            -- 同时删除对应的 context_parts
④ 请求结束 → BodyStorage.Close()
    ├─ memoryStorage: 释放 []byte
    └─ diskStorage: 关闭文件 + 删除临时文件
```

### L2 四项深度指标

| 指标 | 公式 | 回答的问题 | 正常范围 |
|------|------|-----------|---------|
| **上下文结构** | `AVG(token_count) WHERE part_type='system'` 等 | Token 花在 system/history/tool/file 哪些部分？ | 视场景 |
| **重复前缀率** | `AVG(CASE WHEN is_repeated AND is_prefix THEN 1.0 ELSE 0.0)` | 有多少稳定前缀被反复传入？ | 0-1，越高浪费越大 |
| **缓存友好度** | `AVG(CASE WHEN is_prefix AND is_stable THEN 1.0 ELSE 0.0)` | 请求结构是否有利于 Provider 缓存？ | 0-1，越高越好 |
| **缓存兑现率** | `AVG(CASE WHEN is_cache_friendly THEN 1.0 ELSE 0.0)` | 实际被缓存覆盖的比例 | 0-1，应接近缓存友好度 |

### 指标计算细节

**上下文结构**是按 `part_type` 分组 AVG(token_count)，展示每个请求平均在各类 context 上花多少 Token。

**重复前缀率**的逻辑：一个 part 如果 `is_prefix=true`（属于 system prompt 或 tool definitions）且 `is_repeated=true`（content_hash 在多个 part 中出现），说明这个前缀内容在多次请求中被重复发送。这个值高说明有优化空间——应该利用 Provider 的缓存机制避免重复传输。

**缓存友好度**：`is_prefix AND is_stable` 的比例。一个前缀如果每次请求内容都不变（stable），那么 Provider 的 prompt caching 可以缓存它。这个值高说明"技术上可缓存"的部分多。

**缓存兑现率**：`is_cache_friendly` 的比例。结合 L1 的 `cache_read_tokens`，可以判断 Provider 是否真正兑现了缓存收益。如果缓存友好度高但 L1 缓存复用率低，说明 Provider 没有兑现缓存。

### SQL 聚合逻辑

```sql
SELECT
  rdp.model_name AS name,
  COUNT(DISTINCT rdp.request_id) AS sample_count,
  COALESCE(AVG(CASE WHEN rc.part_type = 'system' THEN rc.token_count END), 0) AS avg_system_tokens,
  COALESCE(AVG(CASE WHEN rc.part_type = 'history' THEN rc.token_count END), 0) AS avg_history_tokens,
  COALESCE(AVG(CASE WHEN rc.part_type = 'tool' THEN rc.token_count END), 0) AS avg_tool_tokens,
  COALESCE(AVG(CASE WHEN rc.part_type = 'file' THEN rc.token_count END), 0) AS avg_file_tokens,
  COALESCE(AVG(CASE WHEN rc.is_repeated = 1 AND rc.is_prefix = 1 THEN 1.0 ELSE 0.0 END), 0) AS repeated_prefix_rate,
  COALESCE(AVG(CASE WHEN rc.is_prefix = 1 AND rc.is_stable = 1 THEN 1.0 ELSE 0.0 END), 0) AS cache_friendliness,
  COALESCE(AVG(CASE WHEN rc.is_cache_friendly = 1 THEN 1.0 ELSE 0.0 END), 0) AS cache_fulfillment_rate
FROM request_debug_payloads rdp
INNER JOIN request_context_parts rc ON rc.request_id = rdp.request_id
LEFT JOIN logs lg ON lg.request_id = rdp.request_id AND lg.type = 2
WHERE rdp.created_at >= :start AND rdp.created_at <= :end
GROUP BY rdp.model_name;
```

**注意 `INNER JOIN`**：如果 parser 解析失败（body 无效），payload 有记录但 parts 为空，这条 payload 不会出现在 L2 聚合结果中。

### 分组维度

L2 支持与 L1 相同的 4 个维度，但主表是 `request_debug_payloads`，非管理员维度需要 LEFT JOIN logs 表获取 username/token_name。

### 常见分析逻辑

1. **Token 花在哪了？**：看上下文结构 → avg_system_tokens 异常高 → system prompt 太大，应精简或利用缓存
2. **重复前缀率高**：说明每次请求都传了相同的 system prompt + tool definitions → 应启用 Provider 的 prompt caching（如 Anthropic 的 cache_control）
3. **缓存友好度高，但 L1 缓存复用率低**：Provider 没有兑现缓存收益 → 检查渠道配置，可能缓存未启用或不支持
4. **tool tokens 占比高**：应用可能在发送大量工具定义 → 考虑按需加载工具，而非全量发送
5. **某个用户 history tokens 异常高**：该用户的对话上下文过长 → 考虑截断历史或使用摘要

---

## 两层指标的关系

```
L1 告诉你"有问题" → L2 告诉你"问题在哪"

L1: 缓存复用率 30%          → 缓存利用不够
L2: 缓存友好度 85%          → 技术上可缓存的部分很多
    缓存兑现率 10%           → 但实际只有 10% 被缓存
    重复前缀率 70%           → 70% 的前缀在反复传输
    avg_system_tokens 8000  → system prompt 很大
结论: system prompt 很大且每次重复，技术上完全可缓存，但 Provider 只兑现了 10%
行动: 检查渠道的缓存配置，或换一个支持 prompt caching 的 Provider
```

---

## 采样与存储管理

### 采样策略

| 策略 | 实现 | 说明 |
|------|------|------|
| 轮询保底 | 原子计数器 `% 10 == 0` | 每 10 个请求保证采样 1 个 |
| 随机采样 | request_id 哈希 < SampleRate | 默认 5%，确定性哈希保证同一请求结果一致 |
| 模型白名单 | `capture_models` 配置 | 空=全部，逗号分隔指定模型 |

### 当前存储管理

| 配置项 | 默认值 | 实际生效 |
|--------|--------|---------|
| `sample_rate` | 0.05 | ✅ 生效 |
| `max_payload_size` | 102400 (100KB) | ✅ 生效，截断后闭合 JSON |
| `retention_days` | 7 | ❌ **未实现**，没有后台清理任务 |
| `capture_models` | "" (全部) | ✅ 生效 |

**清理机制**：每次新采样写入后，硬编码 `maxSamples = 100`，删除最老记录只保留最新 100 条。这是当前唯一的清理方式。

### 已知限制

| 限制 | 影响 | 建议 |
|------|------|------|
| `maxSamples = 100` 硬编码 | 高流量场景 100 条很快被替换 | 改为可配置，或按时间保留 |
| `RetentionDays` 未实现 | 数据只靠条数上限清理 | 实现后台 goroutine 定时清理 |
| `LinkDebugPayloadToLog` 未调用 | `log_id` 永远为 0 | 在 RecordConsumeLog 后调用关联 |
| body 存 TEXT 列 | 大 JSON 直接入库，无压缩 | 改为文件存储 + 数据库存元数据 |
| INNER JOIN 无 fallback | parser 失败的 payload 在 L2 汇总中消失 | 改为 LEFT JOIN 或单独查询 payload |
| `maxPayloadSize = 100KB` | Claude Code 的 system prompt 常超 50KB | 提高上限或按需调整 |
