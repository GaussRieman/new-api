# TokenScope 版本路线图

## 版本规划总览

| 版本 | 里程碑 | 核心能力 | 状态 |
|------|--------|---------|------|
| v0.1.0 | L1 Runtime Monitoring | 可观测、可聚合、可对比 | ✅ 已完成 |
| v0.2.0 | L2 Deep Diagnostics | 可诊断、可定位 | 🟡 进行中 |
| v0.3.0 | Modi-Harness Integration | 可优化、可闭环 | ⬜ 规划中 |
| v0.4.0 | Data Backfill & Polish | 历史可用、体验完善 | ⬜ 规划中 |

---

## v0.1.0 — L1 Runtime Monitoring ✅

**目标**：从现有 Gateway 日志数据中产出三项核心运行指标，让企业 Token 消耗从"黑盒"变成"可观测"。

### L1 三项核心指标

| 指标 | 公式 | 回答的问题 |
|------|------|-----------|
| 产出成本 | `quota / completion_tokens` | AI 输出贵不贵？ |
| 上下文负载 | `prompt_tokens / completion_tokens` | 为了得到输出，喂了多少上下文？ |
| 缓存复用率 | `cache_read_tokens / prompt_tokens` | 输入中有多少来自缓存复用？ |

### 已完成变更

**数据层**：
- `logs` 表新增 3 列：`cache_read_tokens`、`cache_write_tokens`、`input_tokens_total`
- GORM AutoMigrate 自动添加，零停机，非破坏性
- 新列与 `other` JSON 双写，不破坏现有功能
- 所有 RecordConsumeLog 调用点补 0 值（8 处）

**后端 API**：
- 新增 `/api/tokenscope/l1/*` 路由组（admin 5 端点 + user-self 2 端点）
- `model/token_scope.go`：L1 聚合查询（按模型分组、总览、时序、用户自查询）
- `controller/tokenscope.go`：5 个 API 端点控制器
- 跨 DB 时序分桶（SQLite/MySQL/PostgreSQL）

**前端**：
- 新增 `token-efficiency` feature 模块
- Summary Cards（3 张核心指标）、Model Table（按模型分组表格）、Filter Bar（筛选栏）
- Sidebar 导航入口：侧边栏"成本分析"菜单项（`/token-efficiency`）
- 使用 `SectionPageLayout` 统一页面布局风格
- i18n 40 个翻译 key（en/zh），含页面标题"成本分析"
- API 响应解包：适配后端 `{ success, data }` 统一响应格式（`res.data.data`）

---

## v0.2.0 — L2 Deep Diagnostics 🟡

**目标**：通过 Raw Body 采样和 Context Parts 解析，找到 Token 浪费的具体来源。

### L2 四项深度指标

| 指标 | 公式 | 回答的问题 |
|------|------|-----------|
| 上下文结构 | 各 `part_type_tokens / input_tokens` | Token 花在 system/history/tool/file/memory 哪些部分？ |
| 重复前缀率 | `repeated_prefix_tokens / input_tokens` | 有多少稳定前缀被反复传入？ |
| 缓存友好度 | `stable_prefix_tokens / input_tokens` | 请求结构是否有利于 Provider 缓存？ |
| 缓存兑现率 | `cache_read_tokens / theoretical_cache_friendly_tokens` | Provider 是否真正返回了缓存收益？ |

### 已完成基础

**数据模型**：
- `model/request_debug_payload.go`：`RequestDebugPayload` 表（raw request body 采样）
- `model/request_context_parts.go`：`RequestContextPart` 表（解析后的上下文结构）
- 已注册到 GORM AutoMigrate（migrateDB + migrateDBFast）
- 辅助函数：CreateRequestDebugPayload, CreateRequestContextParts, MarkDebugPayloadParsed, DeleteOldDebugPayloads

**配置层**：
- `setting/tokenscope_setting/config.go`：L2 采样配置（enabled, sample_rate, max_payload_size, retention_days, capture_models）
- 已注册到 config.GlobalConfig
- 已在 main.go 中 import 初始化

**Relay 捕获**：
- `pkg/tokenscope/capture.go`：MaybeCaptureRequest 采样钩子
  - 确定性哈希采样（基于 request_id）
  - 模型白名单过滤
  - 异步捕获（gopool.Go）
  - 自动截断到 MaxPayloadSize + JSON 边界保护
  - 内联调用 context parser

**Context Parser**：
- `pkg/tokenscope/parser/parser.go`：OpenAI/Claude/Gemini 三种格式解析
  - 自动识别 relay_format 选择解析策略
  - part_type 分类：system, history, tool, file, memory, user
  - SHA-256 前缀 content_hash（去重用）
  - 4-chars-per-token 快速估算
  - 自动分析：is_prefix, is_repeated, is_stable, is_cache_friendly
  - ParseAndStore：一站式解析+入库

**L2 API**：
- 新增 `/api/tokenscope/l2/*` 路由组（4 端点）
- `model/token_scope.go` 追加：L2 聚合查询、Debug Payload 查询、Context Parts 查询
- `controller/tokenscope.go` 追加：4 个 L2 端点
- 前端 api.ts 追加：L2 类型定义 + 4 个 API 函数
- 前端 L2SummaryTable 组件 + Tabs 切换（L1/L2）

### 待完成
- [x] 前端 L2 Tab 完整渲染验证
- [x] i18n L2 翻译 key 补充
- [ ] 清理任务后台 goroutine（retention_days 自动清理）

---

## v0.3.0 — Modi-Harness Integration

**目标**：对接 Modi-Harness 优化层，实现"优化 → 验证"闭环。

### 核心能力

- Modi-Harness 调整上下文结构后，TokenScope 自动验证优化效果
- 对比优化前后的 L1/L2 指标变化
- 优化效果报告自动生成

### 对 Gateway 层的侵入
- 记录优化标识，用于分组对比
---

## v0.4.0 — Data Backfill & Polish

**目标**：让历史数据也具备分析能力，完善用户体验。

### 核心能力

- 历史数据回填脚本：从 `other` JSON 提取 cache token 数据填充新列
- Dashboard 集成：在首页展示 L1 指标摘要
- i18n 完善：所有语言翻译补全
- 性能优化：大数据量下的查询分页与缓存

---

## 版本依赖关系

```text
v0.1.0 (L1 可观测) ✅
  ↓ 数据基础就绪
v0.2.0 (L2 可诊断) 🟡
  ↓ 诊断能力就绪
v0.3.0 (闭环可优化)
  ↓ 验证能力就绪
v0.4.0 (完善)
```

每个版本独立可用，但后续版本依赖前序版本的数据基础。
