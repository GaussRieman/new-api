import { formatQuota } from '@/lib/format'

/**
 * Format output cost from internal quota units to a human-readable cost per 1K tokens.
 */
export function formatOutputCost(cost: number): string {
  if (cost === 0) return '$0.0000'
  // cost is quota_per_output_token; multiply by 1000 for per-1K display
  const costPer1K = cost * 1000
  return formatQuota(costPer1K) + ' / 1K'
}

/**
 * Format context load ratio (e.g. 12.3 → "12.3 : 1").
 */
export function formatContextLoad(ratio: number): string {
  if (ratio === 0) return '0 : 1'
  return `${ratio.toFixed(1)} : 1`
}

/**
 * Format cache reuse rate as a percentage (0.42 → "42.0%").
 */
export function formatCacheReuseRate(rate: number): string {
  if (rate === 0) return '0.0%'
  return `${(rate * 100).toFixed(1)}%`
}

/**
 * Format a large token count with K/M suffix.
 */
export function formatTokenCount(count: number): string {
  if (count >= 1_000_000) {
    return `${(count / 1_000_000).toFixed(1)}M`
  }
  if (count >= 1_000) {
    return `${(count / 1_000).toFixed(1)}K`
  }
  return count.toString()
}
