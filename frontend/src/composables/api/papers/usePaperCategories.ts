import type { QuestionCategory } from '~/types'

export function usePaperCategories(paperId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId),
    () => $api<QuestionCategory[]>(`/api/papers/${paperId}/categories`),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
