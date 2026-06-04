export type {
  TokenScopeL1Metrics,
  TokenScopeL1TimePoint,
  TokenScopeL2Summary,
  RequestContextPart,
  RequestDebugPayload,
  L1Dimension,
} from './api'

export interface L1FilterState {
  startTimestamp: number | undefined
  endTimestamp: number | undefined
  modelName: string
  group: string
}
