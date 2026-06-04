# TokenScope v0.1.0 — L1 Runtime Monitoring 详细设计

**版本**: tokenscope-v0.1.0-l1
**里程碑**: L1 Runtime Monitoring — 可观测、可聚合、可对比
**日期**: 2026-06-03

---

## 1. 背景与目标

### 1.1 问题

当前 new-api Gateway 的 `logs` 表记录了每次 API 调用的 Token 消耗和成本数据，但缺乏**效率维度的指标**。运营者只能看到"花了多少钱"、"用了多少 Token"，无法回答：

- **输出贵不贵？** — 同样的输出，哪些模型更贵？
- **上下文重不重？** — 为了得到一个输出，喂了多少上下文？
- **缓存有没有生效？** — 输入中有多少来自缓存复用？

### 1.2 目标

从现有 `logs` 表数据中产出三项核心运行指标，实现：

| 目标 | 实现 |
|------|------|
| 看得见 | 统一采集模型调用、Token、成本、缓存等运行数据 |
| 算得清 | 用统一指标衡量不同模型、项目、用户的 Token 使用效率 |
| 可对比 | 支持按模型、按时间、按用户分组的指标对比 |

### 1.3 约束

- **最小侵入**：L1 不修改 Relay 热路径逻辑，仅扩展数据写入和新增查询 API
- **向后兼容**：不破坏现有 `other` JSON 的读取逻辑
- **跨 DB 兼容**：SQLite / MySQL / PostgreSQL 三种数据库均需支持

---

## 2. 数据模型变更

### 2.1 现状

`logs` 表当前结构（关键字段）：

| 字段 | 类型 | 说明 |
|------|------|------|
| `prompt_tokens` | int | 输入 Token 数 |
| `completion_tokens` | int | 输出 Token 数 |
| `quota` | int | 消耗配额（内部单位） |
| `other` | text | 额外信息 JSON |

缓存相关数据**仅**存储在 `other` JSON 字段中：

```json
{
  "cache_tokens": 1000,          // 缓存读取 Token 数
  "cache_ratio": 0.5,            // 缓存计费比率
  "cache_write_tokens": 500,     // 缓存写入 Token 数
  "cache_creation_tokens": 500,  // 缓存创建 Token 数
  "input_tokens_total": 2000     // 归一化总输入 Token 数
}
```

**问题**：`other` 是 TEXT 类型 JSON 字符串，无法用 SQL 聚合函数（`SUM`、`AVG`）直接查询。SQLite 不支持 JSON 函数，MySQL/PostgreSQL 的 JSON 提取语法不同且性能差。

### 2.2 变更方案

在 `logs` 表新增 3 个整数列：

| 新增列 | 类型 | 默认值 | 对应 `other` Key | 说明 |
|--------|------|--------|-------------------|------|
| `cache_read_tokens` | int | 0 | `cache_tokens` | 缓存读取 Token 数 |
| `cache_write_tokens` | int | 0 | `cache_write_tokens` | 缓存写入 Token 数（归一化） |
| `input_tokens_total` | int | 0 | `input_tokens_total` | 归一化总输入 Token 数 |

**命名说明**：
- `cache_read_tokens` 而非 `cache_tokens`：与 Modi Token Efficiency Stack 术语对齐（`cache_read` vs `cache_write`）
- `other` 中的 `cache_tokens` 保持不变，前端已有代码读取该 key

### 2.3 迁移策略

GORM `AutoMigrate(&Log{})` 自动添加新列，特性：

- **零停机**：`ALTER TABLE ADD COLUMN` 不锁表
- **非破坏性**：不修改已有列，不删除数据
- **旧数据兼容**：新列默认值为 0，语义正确（未采集 = 无缓存 Token）
- **双写**：新数据同时写入新列和 `other` JSON，确保向后兼容

### 2.4 数据流

```text
Relay 请求完成
  ↓
service/text_quota.go: PostTextConsumeQuota()
  ↓ 计算 summary.CacheTokens, cacheWriteTokensTotal(), usage.InputTokens
  ↓
model.RecordConsumeLog(params)
  ↓ params.CacheReadTokens = summary.CacheTokens
  ↓ params.CacheWriteTokens = cacheWriteTokensTotal(summary)
  ↓ params.InputTokensTotal = usage.InputTokens
  ↓
LOG_DB.Create(&Log{...})
  → logs.cache_read_tokens  = 1000
  → logs.cache_write_tokens = 500
  → logs.input_tokens_total = 2000
  → logs.other              = {"cache_tokens": 1000, ...}  // 双写
```

---

## 3. L1 指标算法

### 3.1 三项核心指标

#### 产出成本 (Output Cost)

```text
output_cost = total_quota / total_completion_tokens
```

- **含义**：每产出 1 个 Token 需要消耗多少配额
- **单位**：内部配额单位 / Token
- **解读**：数值越高，产出越贵。可跨模型对比成本效率
- **注意**：`total_quota` 和 `total_completion_tokens` 是聚合值（`SUM`），不是单条记录的比

#### 上下文负载 (Context Load)

```text
context_load = total_prompt_tokens / total_completion_tokens
```

- **含义**：为了产出 1 个 Token，平均需要输入多少 Token
- **单位**：比率（无量纲）
- **解读**：数值越高，上下文越"重"。高负载可能意味着 History 过长、Tool 定义过多、System Prompt 臃肿
- **注意**：这是聚合比，不是单条记录比的均值

#### 缓存复用率 (Cache Reuse Rate)

```text
cache_reuse_rate = total_cache_read_tokens / total_prompt_tokens
```

- **含义**：输入 Token 中有多少来自缓存复用
- **单位**：比率（0.0 ~ 1.0）
- **解读**：数值越高，缓存复用越好。低复用率可能意味着请求结构不稳定，或 Provider 不支持缓存
- **注意**：仅当 `cache_read_tokens > 0` 时有意义。旧数据该列为 0，复用率也为 0

### 3.2 计算层级

```text
SQL 层（跨 DB 兼容）:
  SELECT model_name,
         COUNT(*) as request_count,
         SUM(quota) as total_quota,
         SUM(prompt_tokens) as total_prompt_tokens,
         SUM(completion_tokens) as total_output_tokens,
         SUM(cache_read_tokens) as total_cache_read,
         SUM(cache_write_tokens) as total_cache_write,
         SUM(input_tokens_total) as total_input_tokens
  FROM logs
  WHERE type = 2 AND created_at BETWEEN ? AND ?
  GROUP BY model_name

Go 层（除零保护）:
  if total_output_tokens > 0:
    output_cost = total_quota / total_output_tokens
  else:
    output_cost = 0

  if total_output_tokens > 0:
    context_load = total_prompt_tokens / total_output_tokens
  else:
    context_load = 0

  if total_prompt_tokens > 0:
    cache_reuse_rate = total_cache_read / total_prompt_tokens
  else:
    cache_reuse_rate = 0
```

### 3.3 时序分桶

时序数据按小时或天分桶，用于折线图展示。

**跨 DB 兼容**：

| 数据库 | 小时分桶 | 天分桶 |
|--------|---------|--------|
| SQLite | `strftime('%Y-%m-%d %H:00', datetime(created_at, 'unixepoch'))` | `strftime('%Y-%m-%d', datetime(created_at, 'unixepoch'))` |
| MySQL | `DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d %H:00')` | `DATE_FORMAT(FROM_UNIXTIME(created_at), '%Y-%m-%d')` |
| PostgreSQL | `date_trunc('hour', to_timestamp(created_at))` | `date_trunc('day', to_timestamp(created_at))` |

使用 `common.UsingPostgreSQL` / `common.UsingSQLite` 分支选择对应 SQL。

---

## 4. API 设计

详见 `modi-project/api/tokenscope-l1-api.md`。

路由前缀：`/api/tokenscope`

| 端点 | 方法 | 权限 | 说明 |
|------|------|------|------|
| `/api/tokenscope/l1/summary` | GET | Admin | 所有模型汇总指标 |
| `/api/tokenscope/l1/by-model` | GET | Admin | 按模型分组指标 |
| `/api/tokenscope/l1/timeseries` | GET | Admin | 时序数据（折线图） |
| `/api/tokenscope/self/l1/summary` | GET | User | 当前用户汇总指标 |
| `/api/tokenscope/self/l1/by-model` | GET | User | 当前用户按模型分组指标 |

---

## 5. 前端设计

### 5.1 Feature 模块

位置：`web/default/src/features/token-efficiency/`

```text
features/token-efficiency/
  api.ts                     — API 请求函数
  types.ts                   — TypeScript 类型定义
  index.tsx                  — 主页面
  section-registry.tsx       — Section 注册（sidebar 用）
  lib/
    format.ts                — 指标格式化（百分比、成本显示）
    metrics.ts               — 前端指标计算辅助
  components/
    l1-summary-cards.tsx     — 三张核心指标卡片
    l1-model-table.tsx       — 按模型分组的指标表格
    l1-timeseries-chart.tsx  — 时序折线图
    filter-bar.tsx           — 时间范围、模型、分组筛选
```

### 5.2 核心展示

#### Summary Cards

三张卡片，展示全局汇总指标：

| 卡片 | 指标 | 格式 | 辅助信息 |
|------|------|------|----------|
| 产出成本 | output_cost | `$X.XXXX / 1K tokens` | 环比变化 |
| 上下文负载 | context_load | `X.XX : 1` | 环比变化 |
| 缓存复用率 | cache_reuse_rate | `XX.X%` | 环比变化 |

#### Model Table

各模型的 L1 指标对比表格：

| 列 | 说明 |
|----|------|
| 模型名称 | model_name |
| 请求数 | request_count |
| 产出成本 | output_cost |
| 上下文负载 | context_load |
| 缓存复用率 | cache_reuse_rate |
| 总配额 | total_quota |
| 总输入 | total_prompt_tokens |
| 总输出 | total_output_tokens |
| 缓存读取 | total_cache_read |

支持按任意列排序。

#### Time Series Chart

产出成本和缓存复用率随时间变化的折线图：
- X 轴：时间（按小时或天）
- Y 轴左：产出成本
- Y 轴右：缓存复用率
- 支持多模型对比

### 5.3 导航

Sidebar 新增 "Token Efficiency" 入口：
- Admin 用户：可看到全局数据
- 普通用户：只能看到自己的数据

---

## 6. 文件变更清单

### 修改

| 文件 | 变更 |
|------|------|
| `model/log.go` | Log struct 加 3 列 + RecordConsumeLogParams 加 3 字段 + RecordConsumeLog 赋值 |
| `service/text_quota.go` | PostTextConsumeLog 调用处传入 cache token 值 + 新增 usageInputTokens 辅助函数 |
| `service/quota.go` | 其他 RecordConsumeLog 调用处传入 0 值 |
| `service/violation_fee.go` | RecordConsumeLog 调用处传入 0 值 |
| `service/task_billing.go` | RecordConsumeLog 调用处传入 0 值 |
| `relay/mjproxy_handler.go` | RecordConsumeLog 调用处传入 0 值 |
| `controller/channel-test.go` | RecordConsumeLog 调用处传入 0 值 |
| `router/api-router.go` | 新增 /tokenscope 路由组 |
| `web/default/src/i18n/locales/en.json` | 新增翻译 key |
| `web/default/src/i18n/locales/zh.json` | 新增翻译 key |

### 新建

| 文件 | 用途 |
|------|------|
| `model/token_scope.go` | L1 聚合查询函数和结构体 |
| `controller/tokenscope.go` | L1 API 控制器 |
| `web/default/src/features/token-efficiency/*` | 前端 Feature 模块 |

---

## 7. 验证方案

| 验证项 | 方法 |
|--------|------|
| 后端编译 | `go build ./...` 通过 |
| 数据库迁移 | 启动后 GORM AutoMigrate 自动新增 3 列，旧数据不受影响 |
| API 功能 | `curl /api/tokenscope/l1/summary?start_timestamp=xxx&end_timestamp=xxx` 返回正确聚合数据 |
| 数据一致性 | 新产生的 Log 记录同时在新列和 Other JSON 中包含 cache token 数据 |
| 前端展示 | 访问 `/token-efficiency` 页面可看到三张指标卡片和模型表格 |
| 跨 DB | 在 SQLite / MySQL / PostgreSQL 上均通过上述测试 |

---

## 8. 未来演进

v0.1.0 完成后，L1 指标数据基础就绪，可进入 v0.2.0 L2 Deep Diagnostics：

- 新增 `request_debug_payload` 和 `request_context_parts` 表
- 实现 Context Parser（OpenAI/Claude/Gemini 格式）
- 采样与 Debug 配置
- 四项深度诊断指标

详见 `modi-project/design/roadmap.md`。
