/*
Copyright (C) 2023-2026 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/
import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import { ChevronDown, ChevronRight } from 'lucide-react'
import {
  getTokenScopeL2Recent,
  getTokenScopeSelfL2Recent,
  getTokenScopeL2RequestDetail,
  type RequestContextPart,
  type RequestDebugPayload,
} from '../api'
import { formatTokenCount } from '../lib/format'

interface L2RequestTableProps {
  filters: {
    startTimestamp: number
    endTimestamp: number
    modelName: string
    group: string
  }
  dimension: string
  loading: boolean
}

const PART_TYPE_ICONS: Record<string, string> = {
  system: '',
  history: '💬',
  tool: '🔧',
  file: '📄',
  memory: '🧠',
  user: '👤',
}

const PART_TYPE_LABELS: Record<string, string> = {
  system: 'System',
  history: 'History',
  tool: 'Tool',
  file: 'File',
  memory: 'Memory',
  user: 'User',
}

export function L2RequestTable({
  filters,
  dimension,
  loading,
}: L2RequestTableProps) {
  const { t } = useTranslation()
  const [requests, setRequests] = useState<RequestDebugPayload[]>([])
  const [expandedId, setExpandedId] = useState<string | null>(null)
  const [parts, setParts] = useState<Record<string, RequestContextPart[]>>({})
  const [detailLoading, setDetailLoading] = useState<Set<string>>(new Set())

  const fetchRequests = useCallback(async () => {
    try {
      const params: Record<string, unknown> = {
        start_timestamp: filters.startTimestamp,
        end_timestamp: filters.endTimestamp,
        limit: 50,
      }
      if (filters.modelName) params.model_name = filters.modelName
      if (filters.group) params.group = filters.group

      const data = await getTokenScopeL2Recent(params)
      setRequests(data)
    } catch {
      // Error handled silently
    }
  }, [filters])

  const fetchRequestDetail = useCallback(async (requestId: string) => {
    if (parts[requestId]) return

    setDetailLoading((prev) => new Set(prev).add(requestId))
    try {
      const result = await getTokenScopeL2RequestDetail(requestId)
      setParts((prev) => ({ ...prev, [requestId]: result.parts }))
    } catch {
      // Error handled silently
    } finally {
      setDetailLoading((prev) => {
        const next = new Set(prev)
        next.delete(requestId)
        return next
      })
    }
  }, [parts])

  useEffect(() => {
    fetchRequests()
  }, [fetchRequests])

  const toggleExpand = async (requestId: string) => {
    if (expandedId === requestId) {
      setExpandedId(null)
    } else {
      setExpandedId(requestId)
      await fetchRequestDetail(requestId)
    }
  }

  const formatDate = (ts: number) => {
    const d = new Date(ts * 1000)
    return d.toLocaleDateString() + ' ' + d.toLocaleTimeString()
  }

  if (loading) {
    return <div className='h-32 animate-pulse rounded bg-muted' />
  }

  if (requests.length === 0) {
    return (
      <div className='py-8 text-center text-muted-foreground'>
        {t('No sampled requests available')}
        <p className='mt-1 text-sm'>
          {t('Enable L2 sampling in settings or adjust the time range')}
        </p>
      </div>
    )
  }

  return (
    <div className='space-y-4'>
      <div className='rounded-md border'>
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className='w-8'></TableHead>
              <TableHead>Request ID</TableHead>
              <TableHead>{t('Model')}</TableHead>
              <TableHead className='text-right'>{t('Parts')}</TableHead>
              <TableHead className='text-right'>
                {t('Total Tokens (est)')}
              </TableHead>
              <TableHead>{t('Time')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {requests.map((row) => {
              const isExpanded = expandedId === row.request_id
              const rowParts = parts[row.request_id] || []
              const isDetailLoading = detailLoading.has(row.request_id)
              const totalTokens = rowParts.reduce(
                (sum, p) => sum + (p.token_count || 0),
                0
              )

              return (
                <>
                  <TableRow
                    key={row.request_id}
                    className='cursor-pointer hover:bg-muted/50'
                    onClick={() => toggleExpand(row.request_id)}
                  >
                    <TableCell className='w-8'>
                      {isDetailLoading ? (
                        <div className='h-4 w-4 animate-spin rounded-full border-2 border-muted-foreground border-t-transparent' />
                      ) : isExpanded ? (
                        <ChevronDown className='h-4 w-4' />
                      ) : (
                        <ChevronRight className='h-4 w-4 text-muted-foreground' />
                      )}
                    </TableCell>
                    <TableCell className='font-mono text-xs'>
                      {row.request_id.slice(0, 12)}…
                    </TableCell>
                    <TableCell className='font-medium'>
                      {row.model_name}
                    </TableCell>
                    <TableCell className='text-right'>
                      {rowParts.length > 0 ? rowParts.length : '—'}
                    </TableCell>
                    <TableCell className='text-right'>
                      {rowParts.length > 0
                        ? formatTokenCount(totalTokens)
                        : '—'}
                    </TableCell>
                    <TableCell className='text-xs text-muted-foreground'>
                      {formatDate(row.created_at)}
                    </TableCell>
                  </TableRow>
                  {isExpanded && (
                    <TableRow>
                      <TableCell colSpan={6} className='bg-muted/30 p-0'>
                        <div className='p-4'>
                          {rowParts.length === 0 ? (
                            <p className='text-sm text-muted-foreground'>
                              {t('No context parts parsed')}
                            </p>
                          ) : (
                            <div className='space-y-2'>
                              <div className='grid grid-cols-6 gap-2 text-xs font-medium text-muted-foreground px-2'>
                                <span>{t('Type')}</span>
                                <span>{t('Name')}</span>
                                <span className='text-right'>
                                  {t('Tokens (est)')}
                                </span>
                                <span className='text-center'>
                                  {t('Prefix')}
                                </span>
                                <span className='text-center'>
                                  {t('Cached')}
                                </span>
                                <span className='text-center'>
                                  {t('Repeated')}
                                </span>
                              </div>
                              {rowParts.map((part) => (
                                <div
                                  key={part.id}
                                  className='grid grid-cols-6 gap-2 items-center rounded px-2 py-1.5 text-sm'
                                  style={{
                                    backgroundColor: part.is_prefix
                                      ? 'hsl(var(--muted))'
                                      : 'transparent',
                                    opacity: part.is_cache_friendly ? 1 : 0.7,
                                  }}
                                >
                                  <span className='flex items-center gap-1.5'>
                                    <span className='text-base'>
                                      {PART_TYPE_ICONS[part.part_type] || '?'}
                                    </span>
                                    <span className='text-xs font-medium'>
                                      {PART_TYPE_LABELS[part.part_type] ||
                                        part.part_type}
                                    </span>
                                  </span>
                                  <span className='truncate text-xs'>
                                    {part.part_name || '—'}
                                  </span>
                                  <span className='text-right font-mono text-xs'>
                                    {formatTokenCount(part.token_count || 0)}
                                  </span>
                                  <span className='text-center'>
                                    {part.is_prefix ? (
                                      <span className='rounded bg-blue-100 px-1.5 py-0.5 text-xs text-blue-700 dark:bg-blue-900 dark:text-blue-300'>
                                        ✓
                                      </span>
                                    ) : (
                                      <span className='text-muted-foreground'>
                                        —
                                      </span>
                                    )}
                                  </span>
                                  <span className='text-center'>
                                    {part.is_cache_friendly ? (
                                      <span className='rounded bg-green-100 px-1.5 py-0.5 text-xs text-green-700 dark:bg-green-900 dark:text-green-300'>
                                        ✓
                                      </span>
                                    ) : (
                                      <span className='text-muted-foreground'>
                                        —
                                      </span>
                                    )}
                                  </span>
                                  <span className='text-center'>
                                    {part.is_repeated ? (
                                      <span className='rounded bg-yellow-100 px-1.5 py-0.5 text-xs text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300'>
                                        ✓
                                      </span>
                                    ) : (
                                      <span className='text-muted-foreground'>
                                        —
                                      </span>
                                    )}
                                  </span>
                                </div>
                              ))}
                              <p className='pt-1 text-xs text-muted-foreground'>
                                {t('Token counts are estimates (4 chars ≈ 1 token)')}
                              </p>
                            </div>
                          )}
                        </div>
                      </TableCell>
                    </TableRow>
                  )}
                </>
              )
            })}
          </TableBody>
        </Table>
      </div>
    </div>
  )
}
