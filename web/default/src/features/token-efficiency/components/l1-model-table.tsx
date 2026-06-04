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
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { TokenScopeL1Metrics, L1Dimension } from '../api'
import {
  formatOutputCost,
  formatContextLoad,
  formatCacheReuseRate,
  formatTokenCount,
} from '../lib/format'
import { formatQuota } from '@/lib/format'

type SortKey = keyof TokenScopeL1Metrics
type SortDirection = 'asc' | 'desc'

interface L1ModelTableProps {
  data: TokenScopeL1Metrics[]
  loading: boolean
  dimension: L1Dimension
}

/** Numeric columns that support sorting */
const SORTABLE_COLUMNS: SortKey[] = [
  'request_count',
  'output_cost',
  'context_load',
  'cache_reuse_rate',
  'total_quota',
  'total_prompt_tokens',
  'total_output_tokens',
  'total_cache_read',
]

export function L1ModelTable({ data, loading, dimension }: L1ModelTableProps) {
  const { t } = useTranslation()
  const [sortKey, setSortKey] = useState<SortKey | null>(null)
  const [sortDirection, setSortDirection] = useState<SortDirection>('desc')

  const handleSort = (key: SortKey) => {
    if (sortKey === key) {
      setSortDirection((prev) => (prev === 'asc' ? 'desc' : 'asc'))
    } else {
      setSortKey(key)
      setSortDirection('desc')
    }
  }

  const sortedData = useMemo(() => {
    if (!sortKey) return data
    return [...data].sort((a, b) => {
      const av = a[sortKey] ?? 0
      const bv = b[sortKey] ?? 0
      return sortDirection === 'asc'
        ? (av as number) - (bv as number)
        : (bv as number) - (av as number)
    })
  }, [data, sortKey, sortDirection])

  const dimensionLabel: Record<L1Dimension, string> = {
    model: t('Model'),
    user: t('User'),
    key: t('API Key'),
    channel: t('Channel'),
  }

  const SortIcon = ({ columnKey }: { columnKey: SortKey }) => {
    const active = sortKey === columnKey
    return (
      <span className='ml-0.5 inline-flex flex-col leading-none -space-y-1'>
        <span
          style={{ fontSize: 10 }}
          className={
            active && sortDirection === 'asc'
              ? 'text-foreground'
              : 'text-muted-foreground opacity-40'
          }
        >
          ▲
        </span>
        <span
          style={{ fontSize: 10 }}
          className={
            active && sortDirection === 'desc'
              ? 'text-foreground'
              : 'text-muted-foreground opacity-40'
          }
        >
          ▼
        </span>
      </span>
    )
  }

  const SortableHead = ({
    columnKey,
    label,
  }: {
    columnKey: SortKey
    label: string
  }) => (
    <TableHead
      className='text-right cursor-pointer select-none'
      onClick={() => handleSort(columnKey)}
    >
      <span className='inline-flex items-center'>
        {label}
        <SortIcon columnKey={columnKey} />
      </span>
    </TableHead>
  )

  const renderNameCell = (row: TokenScopeL1Metrics) => {
    if (dimension === 'channel') {
      return (
        <TableCell className='font-medium'>
          {row.sub_name || row.name}
          {row.sub_id && (
            <span className='ml-1 text-xs text-muted-foreground'>
              #{row.sub_id}
            </span>
          )}
        </TableCell>
      )
    }
    if (dimension === 'user' || dimension === 'key') {
      return (
        <TableCell className='font-medium'>
          {row.name}
          {row.sub_id && (
            <span className='ml-1 text-xs text-muted-foreground'>
              #{row.sub_id}
            </span>
          )}
        </TableCell>
      )
    }
    return <TableCell className='font-medium'>{row.name}</TableCell>
  }

  if (loading) {
    return (
      <div className='space-y-2'>
        {Array.from({ length: 3 }).map((_, i) => (
          <div key={i} className='h-10 animate-pulse rounded bg-muted' />
        ))}
      </div>
    )
  }

  if (data.length === 0) {
    return (
      <div className='py-8 text-center text-muted-foreground'>
        {t('No data available')}
      </div>
    )
  }

  return (
    <div className='rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{dimensionLabel[dimension]}</TableHead>
            <SortableHead columnKey='request_count' label={t('Requests')} />
            <SortableHead columnKey='output_cost' label={t('Output Cost')} />
            <SortableHead columnKey='context_load' label={t('Context Load')} />
            <SortableHead columnKey='cache_reuse_rate' label={t('Cache Reuse')} />
            <SortableHead columnKey='total_quota' label={t('Total Quota')} />
            <SortableHead columnKey='total_prompt_tokens' label={t('Input Tokens')} />
            <SortableHead columnKey='total_output_tokens' label={t('Output Tokens')} />
            <SortableHead columnKey='total_cache_read' label={t('Cache Read')} />
          </TableRow>
        </TableHeader>
        <TableBody>
          {sortedData.map((row) => (
            <TableRow key={row.name + (row.sub_id ?? '')}>
              {renderNameCell(row)}
              <TableCell className='text-right'>
                {row.request_count?.toLocaleString() ?? '-'}
              </TableCell>
              <TableCell className='text-right'>
                {formatOutputCost(row.output_cost)}
              </TableCell>
              <TableCell className='text-right'>
                {formatContextLoad(row.context_load)}
              </TableCell>
              <TableCell className='text-right'>
                {formatCacheReuseRate(row.cache_reuse_rate)}
              </TableCell>
              <TableCell className='text-right'>
                {formatQuota(row.total_quota)}
              </TableCell>
              <TableCell className='text-right'>
                {formatTokenCount(row.total_prompt_tokens ?? 0)}
              </TableCell>
              <TableCell className='text-right'>
                {formatTokenCount(row.total_output_tokens ?? 0)}
              </TableCell>
              <TableCell className='text-right'>
                {formatTokenCount(row.total_cache_read ?? 0)}
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
