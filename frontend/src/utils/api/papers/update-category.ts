import type { QuestionCategory } from '~/types'

interface UpdateCategoryBody {
  name: string
}

export async function updateCategory(
  categoryId: number,
  paperId: number,
  body: UpdateCategoryBody
): Promise<void> {
  const { data: categories } = useNuxtData<QuestionCategory[]>(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)
  )

  const previousCategories = categories.value!

  categories.value = categories.value!.map(category =>
    category.id === categoryId ? { ...category, ...body } : category
  )

  try {
    await $fetch(`/api/categories/${categoryId}`, {
      method: 'PATCH',
      body,
      ...getFetchOptions(),
    })

    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
  } catch {
    categories.value = previousCategories
  }
}
