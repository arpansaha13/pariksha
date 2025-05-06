import type { Paper } from '~/types'

export function usePaper(paperId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData<Paper>(AsyncDataKeys.PAPERS_PAPER(paperId), () =>
    $api(`/api/papers/${paperId}`)
  )
}
