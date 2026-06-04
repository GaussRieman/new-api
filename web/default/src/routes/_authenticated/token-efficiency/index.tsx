import { createFileRoute } from '@tanstack/react-router'
import { TokenEfficiency } from '@/features/token-efficiency'

export const Route = createFileRoute('/_authenticated/token-efficiency/')({
  component: TokenEfficiency,
})
