import type { QuestionCategory } from '~/types'

export async function reorderCategories(
  paperId: number,
  categoryIds: number[]
): Promise<void> {
  const { data: storedCategories } = useNuxtData<QuestionCategory[]>(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)
  )

  const previousCategories = storedCategories.value!

  // Update orders based on new array position
  storedCategories.value = categoryIds.map(
    categoryId => storedCategories.value!.find(cat => cat.id === categoryId)!
  )

  try {
    await $fetch(`/api/papers/${paperId}/categories/reorder`, {
      method: 'PATCH',
      body: { categories: categoryIds },
      ...getFetchOptions(),
    })
    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
  } catch {
    storedCategories.value = previousCategories
  }
}
