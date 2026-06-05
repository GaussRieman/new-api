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
import { useEffect } from 'react'
import * as z from 'zod'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useTranslation } from 'react-i18next'
import {
  Form,
  FormControl,
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form'
import { Input } from '@/components/ui/input'
import { SettingsForm } from '../components/settings-form-layout'
import { SettingsPageFormActions } from '../components/settings-page-context'
import { SettingsSection } from '../components/settings-section'
import { useUpdateOption } from '../hooks/use-update-option'

const tokenscopeSchema = z.object({
  sample_rate: z.coerce.number().min(0).max(1),
})

type TokenScopeFormValues = z.infer<typeof tokenscopeSchema>

interface TokenScopeSectionProps {
  defaultSampleRate: number
}

export function TokenScopeSection({
  defaultSampleRate,
}: TokenScopeSectionProps) {
  const { t } = useTranslation()
  const updateOption = useUpdateOption()
  const form = useForm<TokenScopeFormValues>({
    resolver: zodResolver(tokenscopeSchema),
    defaultValues: {
      sample_rate: defaultSampleRate,
    },
  })

  useEffect(() => {
    form.reset({ sample_rate: defaultSampleRate })
  }, [defaultSampleRate, form])

  const onSubmit = async (values: TokenScopeFormValues) => {
    await updateOption.mutateAsync({
      key: 'tokenscope_setting.sample_rate',
      value: values.sample_rate,
    })
  }

  return (
    <SettingsSection title={t('L2 Debug Sampling')}>
      <Form {...form}>
        <SettingsForm onSubmit={form.handleSubmit(onSubmit)}>
          <SettingsPageFormActions
            onSave={form.handleSubmit(onSubmit)}
            isSaving={updateOption.isPending}
            saveLabel={t('Save')}
          />
          <FormField
            control={form.control}
            name='sample_rate'
            render={({ field }) => (
              <FormItem className='space-y-2'>
                <FormLabel>{t('Sample rate')}</FormLabel>
                <FormDescription>
                  {t(
                    'Fraction of requests to capture (0.0–1.0). Plus guaranteed 1-in-10 round-robin.'
                  )}
                </FormDescription>
                <FormControl>
                  <Input
                    type='number'
                    step='0.01'
                    min='0'
                    max='1'
                    {...field}
                    onChange={(e) => field.onChange(parseFloat(e.target.value) || 0)}
                  />
                </FormControl>
                <FormMessage />
              </FormItem>
            )}
          />
        </SettingsForm>
      </Form>
    </SettingsSection>
  )
}
