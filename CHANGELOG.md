# Changelog

## [Unreleased] — TokenScope 成本分析

### 新增

- **成本分析页面**：侧边栏新增"成本分析"菜单项，独立路由 `/token-efficiency`
  - L1 运行监控 Tab：三项核心指标卡片（产出成本、上下文负载、缓存复用率）+ 按模型分组明细表
  - L2 深度诊断 Tab（仅管理员可见）：采样数、各部分平均 Token、重复前缀率、缓存友好度、缓存兑现率
  - 筛选栏：支持 24h / 7d / 30d 快捷时间范围、模型名称筛选、分组筛选

- **后端 API**：
  - `/api/tokenscope/l1/summary` — L1 总览指标（admin）
  - `/api/tokenscope/l1/by-model` — L1 按模型分组（admin）
  - `/api/tokenscope/l1/timeseries` — L1 时序数据（admin）
  - `/api/tokenscope/self/l1/summary` — L1 当前用户总览
  - `/api/tokenscope/self/l1/by-model` — L1 当前用户按模型
  - `/api/tokenscope/l2/summary` — L2 诊断指标（admin）
  - `/api/tokenscope/l2/recent` — L2 最近采样请求（admin）
  - `/api/tokenscope/l2/request/:request_id` — L2 请求详情（admin）
  - `/api/tokenscope/self/l2/recent` — L2 当前用户最近请求

- **数据层**：
  - `logs` 表新增 3 列：`cache_read_tokens`、`cache_write_tokens`、`input_tokens_total`（GORM AutoMigrate 自动添加）
  - `request_debug_payloads` 表：L2 调试数据存储（raw request body 采样）
  - `request_context_parts` 表：解析后的上下文结构（system/history/tool/file 分类 + 缓存分析标记）

- **L2 采样与解析**：
  - 请求体采样钩子（确定性哈希采样 + 模型白名单 + 异步捕获）
  - OpenAI / Claude / Gemini 三种格式自动解析
  - SHA-256 前缀去重、4-chars-per-token 快速估算
  - 自动标记：is_prefix / is_repeated / is_stable / is_cache_friendly

- **配置**：
  - `tokenscope_setting`：L2 采样开关、采样率、最大 payload、保留天数、捕获模型列表

- **i18n**：新增 40 个翻译 key（中/英），涵盖成本分析页面全部文案

### 修复

- 前端 API 调用适配后端 `{ success, data }` 统一响应格式（`res.data.data` 解包）
- 侧边栏"成本分析"菜单项默认可见（`DEFAULT_SIDEBAR_MODULES.console.tokenscope: true`）
- 翻译 key 放入 `translation` 命名空间内，确保中文切换生效
- L1SummaryCards 组件 null 安全处理（`?.` 可选链 + 空值回退）
