export function usePaperCategories(paperId: PaperId) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    UseAsyncDataKeys.paper_categories(paperId),
    () => $api<QuestionCategory[]>(`/api/papers/${paperId}/categories`),
    {
      transform: categories => categories.toSorted((a, b) => a.order - b.order),
    }
  )
}
