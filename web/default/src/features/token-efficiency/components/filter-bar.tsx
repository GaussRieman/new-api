import { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { Combobox } from '@/components/ui/combobox'
import type { L1FilterState } from '../types'
import {
  getTokenScopeFilters,
  getTokenScopeSelfFilters,
} from '../api'
import type { TokenScopeFilterOptions } from '../api'
import { useIsAdmin } from '@/hooks/use-admin'

interface FilterBarProps {
  filters: L1FilterState
  onChange: (filters: L1FilterState) => void
}

export function FilterBar({ filters, onChange }: FilterBarProps) {
  const { t } = useTranslation()
  const isAdmin = useIsAdmin()
  const [filterOptions, setFilterOptions] = useState<TokenScopeFilterOptions>({
    model_names: [],
    groups: [],
  })

  useEffect(() => {
    const fetchFilters = async () => {
      try {
        const data = isAdmin
          ? await getTokenScopeFilters()
          : await getTokenScopeSelfFilters()
        setFilterOptions(data)
      } catch {
        // Error handled silently; dropdowns will be empty
      }
    }
    fetchFilters()
  }, [isAdmin])

  const handleQuickRange = (days: number) => {
    const end = Math.floor(Date.now() / 1000)
    const start = end - days * 24 * 3600
    onChange({ ...filters, startTimestamp: start, endTimestamp: end })
  }

  const modelOptions = filterOptions.model_names.map((name) => ({
    value: name,
    label: name,
  }))

  const groupOptions = filterOptions.groups.map((group) => ({
    value: group,
    label: group,
  }))

  const activeRange =
    filters.endTimestamp && filters.startTimestamp
      ? filters.endTimestamp - filters.startTimestamp
      : 0

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Time Range')}:</span>
        <button
          className={`rounded-md border px-2 py-1 text-xs hover:bg-accent ${activeRange === 1 * 24 * 3600 ? 'bg-accent font-medium' : ''}`}
          onClick={() => handleQuickRange(1)}
        >
          {t('24h')}
        </button>
        <button
          className={`rounded-md border px-2 py-1 text-xs hover:bg-accent ${activeRange === 7 * 24 * 3600 ? 'bg-accent font-medium' : ''}`}
          onClick={() => handleQuickRange(7)}
        >
          {t('7d')}
        </button>
        <button
          className={`rounded-md border px-2 py-1 text-xs hover:bg-accent ${activeRange === 30 * 24 * 3600 ? 'bg-accent font-medium' : ''}`}
          onClick={() => handleQuickRange(30)}
        >
          {t('30d')}
        </button>
      </div>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Model')}:</span>
        <Combobox
          options={modelOptions}
          value={filters.modelName || null}
          onValueChange={(value) =>
            onChange({ ...filters, modelName: value ?? '' })
          }
          placeholder={t('All Models')}
          searchPlaceholder={t('Search models...')}
          emptyText={t('No model found.')}
          allowCustomValue
          className='w-48'
        />
      </div>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Group')}:</span>
        <Combobox
          options={groupOptions}
          value={filters.group || null}
          onValueChange={(value) =>
            onChange({ ...filters, group: value ?? '' })
          }
          placeholder={t('All Groups')}
          searchPlaceholder={t('Search groups...')}
          emptyText={t('No group found.')}
          allowCustomValue
          className='w-40'
        />
      </div>
    </div>
  )
}
