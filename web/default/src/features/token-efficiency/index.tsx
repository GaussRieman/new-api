import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { L1SummaryCards } from './components/l1-summary-cards'
import { L1ModelTable } from './components/l1-model-table'
import { L2SummaryTable } from './components/l2-summary-table'
import { FilterBar } from './components/filter-bar'
import {
  getTokenScopeL1Summary,
  getTokenScopeL1ByDimension,
  getTokenScopeSelfL1Summary,
  getTokenScopeSelfL1ByDimension,
  getTokenScopeL2Summary,
  getTokenScopeSelfL2Recent,
} from './api'
import type { TokenScopeL1Metrics, L1FilterState, TokenScopeL2Summary, L1Dimension } from './types'
import { SectionPageLayout } from '@/components/layout'
import { useIsAdmin } from '@/hooks/use-admin'

export function TokenEfficiency() {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()

  // Default to last 7 days
  const getDefaultFilters = useCallback((): L1FilterState => {
    const end = Math.floor(Date.now() / 1000)
    const start = end - 7 * 24 * 3600
    return {
      startTimestamp: start,
      endTimestamp: end,
      modelName: '',
      group: '',
    }
  }, [])

  const [filters, setFilters] = useState<L1FilterState>(getDefaultFilters)
  const [dimension, setDimension] = useState<L1Dimension>('model')
  const [summary, setSummary] = useState<TokenScopeL1Metrics | null>(null)
  const [byDimension, setByDimension] = useState<TokenScopeL1Metrics[]>([])
  const [l2Summary, setL2Summary] = useState<TokenScopeL2Summary[]>([])
  const [loading, setLoading] = useState(false)
  const [l2Loading, setL2Loading] = useState(false)

  const fetchL1Data = useCallback(async () => {
    setLoading(true)
    try {
      const params: Record<string, unknown> = {
        start_timestamp: filters.startTimestamp,
        end_timestamp: filters.endTimestamp,
        dimension,
      }
      if (filters.modelName) params.model_name = filters.modelName
      if (filters.group) params.group = filters.group

      const [summaryData, byDimensionData] = await Promise.all([
        isAdmin
          ? getTokenScopeL1Summary(params)
          : getTokenScopeSelfL1Summary(params),
        isAdmin
          ? getTokenScopeL1ByDimension(params)
          : getTokenScopeSelfL1ByDimension(params),
      ])

      setSummary(summaryData)
      setByDimension(byDimensionData)
    } catch {
      // Error handled silently; cards show "-"
    } finally {
      setLoading(false)
    }
  }, [filters, isAdmin, dimension])

  const fetchL2Data = useCallback(async () => {
    setL2Loading(true)
    try {
      const params: Record<string, unknown> = {
        start_timestamp: filters.startTimestamp,
        end_timestamp: filters.endTimestamp,
      }
      if (filters.modelName) params.model_name = filters.modelName
      if (filters.group) params.group = filters.group

      if (isAdmin) {
        const l2Data = await getTokenScopeL2Summary(params)
        setL2Summary(l2Data)
      } else {
        const recent = await getTokenScopeSelfL2Recent({ ...params, limit: 10 })
        // For non-admin, we show recent requests but no summary
        setL2Summary([])
      }
    } catch {
      // Error handled silently
    } finally {
      setL2Loading(false)
    }
  }, [filters, isAdmin])

  useEffect(() => {
    fetchL1Data()
  }, [fetchL1Data])

  useEffect(() => {
    fetchL2Data()
  }, [fetchL2Data])

  return (
    <SectionPageLayout>
      <SectionPageLayout.Title>{t('Cost Analysis')}</SectionPageLayout.Title>
      <SectionPageLayout.Content>
        <div className='space-y-6'>
          <p className='text-muted-foreground'>
            {t('Monitor token usage efficiency across models and time')}
          </p>

          <FilterBar filters={filters} onChange={setFilters} />

          <Tabs defaultValue='l1'>
            <TabsList>
              <TabsTrigger value='l1'>{t('L1 Monitoring')}</TabsTrigger>
              {isAdmin && (
                <TabsTrigger value='l2'>{t('L2 Diagnostics')}</TabsTrigger>
              )}
            </TabsList>

            <TabsContent value='l1' className='space-y-6'>
              <L1SummaryCards data={summary} loading={loading} />
              <div>
                <ToggleGroup
                  value={dimension}
                  onValueChange={(value) => {
                    const newDimension = Array.isArray(value) ? value[0] : value
                    if (newDimension) setDimension(newDimension as L1Dimension)
                  }}
                  variant='outline'
                  size='sm'
                  type='single'
                  className='mb-4'
                >
                  <ToggleGroupItem value='model'>
                    {t('Model')}
                  </ToggleGroupItem>
                  <ToggleGroupItem value='key'>
                    {t('API Key')}
                  </ToggleGroupItem>
                  {isAdmin && (
                    <>
                      <ToggleGroupItem value='user'>
                        {t('User')}
                      </ToggleGroupItem>
                      <ToggleGroupItem value='channel'>
                        {t('Channel')}
                      </ToggleGroupItem>
                    </>
                  )}
                </ToggleGroup>
                <L1ModelTable
                  data={byDimension}
                  loading={loading}
                  dimension={dimension}
                />
              </div>
            </TabsContent>

            <TabsContent value='l2' className='space-y-6'>
              <div>
                <h2 className='mb-4 text-lg font-semibold'>
                  {t('L2 Deep Diagnostics')}
                </h2>
                <L2SummaryTable data={l2Summary} loading={l2Loading} />
              </div>
            </TabsContent>
          </Tabs>
        </div>
      </SectionPageLayout.Content>
    </SectionPageLayout>
  )
}
