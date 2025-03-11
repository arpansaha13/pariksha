import type { Paper } from '~/types'

export function usePaper(paperId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData<Paper>(AsyncDataKeys.PAPERS_PAPER(paperId), () =>
    $fetch(`/api/papers/${paperId}`, fetchOptions)
  )
}
