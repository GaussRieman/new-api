import { useTranslation } from 'react-i18next'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { TokenScopeL1Metrics } from '../api'
import {
  formatOutputCost,
  formatContextLoad,
  formatCacheReuseRate,
  formatTokenCount,
} from '../lib/format'
import { formatQuota } from '@/lib/format'

interface L1ModelTableProps {
  data: TokenScopeL1Metrics[]
  loading: boolean
}

export function L1ModelTable({ data, loading }: L1ModelTableProps) {
  const { t } = useTranslation()

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
            <TableHead>{t('Model')}</TableHead>
            <TableHead className='text-right'>{t('Requests')}</TableHead>
            <TableHead className='text-right'>{t('Output Cost')}</TableHead>
            <TableHead className='text-right'>{t('Context Load')}</TableHead>
            <TableHead className='text-right'>{t('Cache Reuse')}</TableHead>
            <TableHead className='text-right'>{t('Total Quota')}</TableHead>
            <TableHead className='text-right'>{t('Input Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Output Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Cache Read')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row) => (
            <TableRow key={row.model_name}>
              <TableCell className='font-medium'>{row.model_name}</TableCell>
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
