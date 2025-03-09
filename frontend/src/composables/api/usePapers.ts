import type { Paper } from '~/types/models'
import { AsyncDataKeys } from './async-data-keys'

export function usePapers() {
  const fetchOptions = getFetchOptions()

  return useAsyncData<Paper[]>(AsyncDataKeys.PAPERS, () =>
    $fetch('/api/papers', fetchOptions)
  )
}
