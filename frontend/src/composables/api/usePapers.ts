import type { Paper } from '~/types'

export function usePapers() {
  const fetchOptions = getFetchOptions()

  return useAsyncData<Paper[]>(AsyncDataKeys.PAPERS, () =>
    $fetch('/api/papers', fetchOptions)
  )
}
