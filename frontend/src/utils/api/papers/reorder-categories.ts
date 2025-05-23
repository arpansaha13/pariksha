export async function reorderCategories(
  paperId: PaperId,
  categoryIds: number[]
): Promise<void> {
  const { $api } = useNuxtApp()

  const { data: storedCategories } = useNuxtData<QuestionCategory[]>(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)
  )

  const previousCategories = storedCategories.value!

  // Update orders based on new array position
  storedCategories.value = categoryIds.map(
    categoryId => storedCategories.value!.find(cat => cat.id === categoryId)!
  )

  try {
    await $api(`/api/papers/${paperId}/categories/reorder`, {
      method: 'PATCH',
      body: { categories: categoryIds },
    })
    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
  } catch {
    storedCategories.value = previousCategories
  }
}
