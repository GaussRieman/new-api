import { useTranslation } from 'react-i18next'
import type { L1FilterState } from '../types'

interface FilterBarProps {
  filters: L1FilterState
  onChange: (filters: L1FilterState) => void
}

export function FilterBar({ filters, onChange }: FilterBarProps) {
  const { t } = useTranslation()

  // Default to last 7 days if no timestamps set
  const getDefaultRange = () => {
    const end = Math.floor(Date.now() / 1000)
    const start = end - 7 * 24 * 3600
    return { start, end }
  }

  const handleQuickRange = (days: number) => {
    const end = Math.floor(Date.now() / 1000)
    const start = end - days * 24 * 3600
    onChange({ ...filters, startTimestamp: start, endTimestamp: end })
  }

  return (
    <div className='flex flex-wrap items-center gap-3'>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Time Range')}:</span>
        <button
          className='rounded-md border px-2 py-1 text-xs hover:bg-accent'
          onClick={() => handleQuickRange(1)}
        >
          {t('24h')}
        </button>
        <button
          className='rounded-md border px-2 py-1 text-xs hover:bg-accent'
          onClick={() => handleQuickRange(7)}
        >
          {t('7d')}
        </button>
        <button
          className='rounded-md border px-2 py-1 text-xs hover:bg-accent'
          onClick={() => handleQuickRange(30)}
        >
          {t('30d')}
        </button>
      </div>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Model')}:</span>
        <input
          type='text'
          className='h-8 w-40 rounded-md border bg-background px-2 text-sm'
          placeholder={t('Filter by model name')}
          value={filters.modelName}
          onChange={(e) =>
            onChange({ ...filters, modelName: e.target.value })
          }
        />
      </div>
      <div className='flex items-center gap-2'>
        <span className='text-sm text-muted-foreground'>{t('Group')}:</span>
        <input
          type='text'
          className='h-8 w-32 rounded-md border bg-background px-2 text-sm'
          placeholder={t('Filter by group')}
          value={filters.group}
          onChange={(e) => onChange({ ...filters, group: e.target.value })}
        />
      </div>
    </div>
  )
}
