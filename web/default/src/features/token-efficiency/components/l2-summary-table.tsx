import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import type { TokenScopeL2Summary } from '../api'

interface L2SummaryTableProps {
  data: TokenScopeL2Summary[]
  loading: boolean
}

export function L2SummaryTable({ data, loading }: L2SummaryTableProps) {
  const { t } = useTranslation()

  if (loading) {
    return <div className='h-32 animate-pulse rounded bg-muted' />
  }

  if (data.length === 0) {
    return (
      <div className='py-8 text-center text-muted-foreground'>
        {t('No L2 data available')}
      </div>
    )
  }

  return (
    <div className='rounded-md border'>
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>{t('Samples')}</TableHead>
            <TableHead className='text-right'>{t('Avg System Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Avg History Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Avg Tool Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Avg File Tokens')}</TableHead>
            <TableHead className='text-right'>{t('Repeated Prefix Rate')}</TableHead>
            <TableHead className='text-right'>{t('Cache Friendliness')}</TableHead>
            <TableHead className='text-right'>{t('Cache Fulfillment')}</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {data.map((row, i) => (
            <TableRow key={i}>
              <TableCell>{row.sample_count}</TableCell>
              <TableCell className='text-right'>
                {row.avg_system_tokens.toFixed(0)}
              </TableCell>
              <TableCell className='text-right'>
                {row.avg_history_tokens.toFixed(0)}
              </TableCell>
              <TableCell className='text-right'>
                {row.avg_tool_tokens.toFixed(0)}
              </TableCell>
              <TableCell className='text-right'>
                {row.avg_file_tokens.toFixed(0)}
              </TableCell>
              <TableCell className='text-right'>
                {(row.repeated_prefix_rate * 100).toFixed(1)}%
              </TableCell>
              <TableCell className='text-right'>
                {(row.cache_friendliness * 100).toFixed(1)}%
              </TableCell>
              <TableCell className='text-right'>
                {(row.cache_fulfillment_rate * 100).toFixed(1)}%
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
