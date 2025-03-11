import type { QuestionCategory } from '~/types'

export function usePaperCategories(paperId: number) {
  const fetchOptions = getFetchOptions()

  return useAsyncData(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId),
    () =>
      $fetch<QuestionCategory[]>(
        `/api/papers/${paperId}/categories`,
        fetchOptions
      ),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
