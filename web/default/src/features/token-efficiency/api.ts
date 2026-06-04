import { api } from '@/lib/api'

export interface TokenScopeFilterOptions {
  model_names: string[]
  groups: string[]
}

export async function getTokenScopeFilters(): Promise<TokenScopeFilterOptions> {
  const res = await api.get<{ success: boolean; data: TokenScopeFilterOptions }>(
    '/api/tokenscope/filters'
  )
  return res.data.data
}

export async function getTokenScopeSelfFilters(): Promise<TokenScopeFilterOptions> {
  const res = await api.get<{ success: boolean; data: TokenScopeFilterOptions }>(
    '/api/tokenscope/self/filters'
  )
  return res.data.data
}

/** L1 grouping dimension */
export type L1Dimension = 'model' | 'user' | 'key' | 'channel'

export interface TokenScopeL1Metrics {
  name: string
  dimension: L1Dimension
  sub_id?: string
  sub_name?: string
  request_count: number
  output_cost: number
  context_load: number
  cache_reuse_rate: number
  total_quota: number
  total_prompt_tokens: number
  total_output_tokens: number
  total_cache_read: number
  total_cache_write: number
  total_input_tokens: number
}

export interface TokenScopeL1TimePoint {
  bucket: string
  request_count: number
  output_cost: number
  context_load: number
  cache_reuse_rate: number
  total_quota: number
  total_prompt_tokens: number
  total_output_tokens: number
  total_cache_read: number
  total_cache_write: number
  total_input_tokens: number
}

interface L1Params {
  start_timestamp?: number
  end_timestamp?: number
  model_name?: string
  group?: string
  channel?: number
  username?: string
}

interface L1ByDimensionParams extends L1Params {
  dimension?: L1Dimension
}

interface TimeSeriesParams extends L1Params {
  bucket?: 'hour' | 'day'
}

export async function getTokenScopeL1Summary(
  params: L1Params = {}
): Promise<TokenScopeL1Metrics> {
  const res = await api.get<{ success: boolean; data: TokenScopeL1Metrics }>(
    '/api/tokenscope/l1/summary',
    { params }
  )
  return res.data.data
}

export async function getTokenScopeL1ByModel(
  params: L1Params = {}
): Promise<TokenScopeL1Metrics[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL1Metrics[]
  }>('/api/tokenscope/l1/by-model', { params })
  return res.data.data
}

/** @deprecated Use getTokenScopeL1ByDimension instead */
export { getTokenScopeL1ByModel as getTokenScopeL1ByDimensionOld }

export async function getTokenScopeL1ByDimension(
  params: L1ByDimensionParams = {}
): Promise<TokenScopeL1Metrics[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL1Metrics[]
  }>('/api/tokenscope/l1/by-dimension', { params })
  return res.data.data
}

export async function getTokenScopeSelfL1Summary(
  params: Omit<L1Params, 'channel' | 'username'> = {}
): Promise<TokenScopeL1Metrics> {
  const res = await api.get<{ success: boolean; data: TokenScopeL1Metrics }>(
    '/api/tokenscope/self/l1/summary',
    { params }
  )
  return res.data.data
}

export async function getTokenScopeSelfL1ByModel(
  params: Omit<L1Params, 'channel' | 'username'> = {}
): Promise<TokenScopeL1Metrics[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL1Metrics[]
  }>('/api/tokenscope/self/l1/by-model', { params })
  return res.data.data
}

export async function getTokenScopeSelfL1ByDimension(
  params: Omit<L1ByDimensionParams, 'channel' | 'username'> = {}
): Promise<TokenScopeL1Metrics[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL1Metrics[]
  }>('/api/tokenscope/self/l1/by-dimension', { params })
  return res.data.data
}

export async function getTokenScopeL1TimeSeries(
  params: TimeSeriesParams = {}
): Promise<TokenScopeL1TimePoint[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL1TimePoint[]
  }>('/api/tokenscope/l1/timeseries', { params })
  return res.data.data
}

// ---------------------------------------------------------------------------
// L2 Deep Diagnostics API
// ---------------------------------------------------------------------------

export interface TokenScopeL2Summary {
  name: string
  sub_id?: string
  sub_name?: string
  sample_count: number
  avg_system_tokens: number
  avg_history_tokens: number
  avg_tool_tokens: number
  avg_file_tokens: number
  repeated_prefix_rate: number
  cache_friendliness: number
  cache_fulfillment_rate: number
}

export interface RequestContextPart {
  id: number
  request_id: string
  part_type: string
  part_name: string
  content_hash: string
  token_count: number
  position: number
  is_prefix: boolean
  is_repeated: boolean
  is_stable: boolean
  is_cache_friendly: boolean
}

export interface RequestDebugPayload {
  id: number
  request_id: string
  log_id: number
  user_id: number
  model_name: string
  channel_id: number
  token_id: number
  group: string
  created_at: number
  request_body: string
  is_stream: boolean
  relay_format: string
  parsed: boolean
  sampling_reason: string
}

export interface L2TimeSeriesParams {
  start_timestamp?: number
  end_timestamp?: number
  model_name?: string
  group?: string
  dimension?: L1Dimension
  channel?: number
  username?: string
}

export async function getTokenScopeL2Summary(
  params: L2TimeSeriesParams = {}
): Promise<TokenScopeL2Summary[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL2Summary[]
  }>('/api/tokenscope/l2/summary', { params })
  return res.data.data || []
}

export async function getTokenScopeL2ByDimension(
  params: L2TimeSeriesParams = {}
): Promise<TokenScopeL2Summary[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL2Summary[]
  }>('/api/tokenscope/l2/by-dimension', { params })
  return res.data.data || []
}

export async function getTokenScopeSelfL2ByDimension(
  params: Omit<L2TimeSeriesParams, 'channel' | 'username'> = {}
): Promise<TokenScopeL2Summary[]> {
  const res = await api.get<{
    success: boolean
    data: TokenScopeL2Summary[]
  }>('/api/tokenscope/self/l2/by-dimension', { params })
  return res.data.data || []
}

export async function getTokenScopeL2Recent(
  params: L2TimeSeriesParams & { limit?: number } = {}
): Promise<RequestDebugPayload[]> {
  const res = await api.get<{
    success: boolean
    data: RequestDebugPayload[]
  }>('/api/tokenscope/l2/recent', { params })
  return res.data.data || []
}

export async function getTokenScopeL2RequestDetail(
  requestId: string
): Promise<{ payload: RequestDebugPayload; parts: RequestContextPart[] }> {
  const res = await api.get<{
    success: boolean
    data: { payload: RequestDebugPayload; parts: RequestContextPart[] }
  }>(`/api/tokenscope/l2/request/${requestId}`)
  return res.data.data
}

export async function getTokenScopeSelfL2Recent(
  params: L2TimeSeriesParams & { limit?: number } = {}
): Promise<RequestDebugPayload[]> {
  const res = await api.get<{
    success: boolean
    data: RequestDebugPayload[]
  }>('/api/tokenscope/self/l2/recent', { params })
  return res.data.data || []
}
