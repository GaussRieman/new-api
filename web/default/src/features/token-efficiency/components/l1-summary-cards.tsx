import { useTranslation } from 'react-i18next'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { TokenScopeL1Metrics } from '../api'
import {
  formatOutputCost,
  formatContextLoad,
  formatCacheReuseRate,
  formatTokenCount,
} from '../lib/format'

interface L1SummaryCardsProps {
  data: TokenScopeL1Metrics | null
  loading: boolean
}

export function L1SummaryCards({ data, loading }: L1SummaryCardsProps) {
  const { t } = useTranslation()

  const cards = [
    {
      title: t('Output Cost'),
      value: data ? formatOutputCost(data.output_cost) : '-',
      description: t('Cost per 1K output tokens'),
      subValue:
        data?.request_count != null
          ? t('{{count}} requests', { count: data.request_count.toLocaleString() })
          : '',
    },
    {
      title: t('Context Load'),
      value: data ? formatContextLoad(data.context_load) : '-',
      description: t('Input tokens per output token'),
      subValue:
        data?.total_prompt_tokens != null && data?.total_output_tokens != null
          ? t('Input {{input}} · Output {{output}}', {
              input: formatTokenCount(data.total_prompt_tokens),
              output: formatTokenCount(data.total_output_tokens),
            })
          : '',
    },
    {
      title: t('Cache Reuse Rate'),
      value: data ? formatCacheReuseRate(data.cache_reuse_rate) : '-',
      description: t('Fraction of input from cache'),
      subValue:
        data?.total_cache_read != null
          ? t('Cache read {{count}} tokens', {
              count: formatTokenCount(data.total_cache_read),
            })
          : '',
    },
  ]

  return (
    <div className='grid gap-4 md:grid-cols-3'>
      {cards.map((card) => (
        <Card key={card.title}>
          <CardHeader className='flex flex-row items-center justify-between space-y-0 pb-2'>
            <div>
              <CardTitle className='text-sm font-medium'>{card.title}</CardTitle>
              <p className='text-xs text-muted-foreground'>{card.description}</p>
            </div>
          </CardHeader>
          <CardContent>
            {loading ? (
              <div className='h-8 w-24 animate-pulse rounded bg-muted' />
            ) : (
              <>
                <div className='text-2xl font-bold'>{card.value}</div>
                {card.subValue && (
                  <p className='mt-1 text-xs text-muted-foreground'>
                    {card.subValue}
                  </p>
                )}
              </>
            )}
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
