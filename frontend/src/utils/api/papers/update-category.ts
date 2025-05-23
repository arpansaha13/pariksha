interface UpdateCategoryBody {
  name: string
}

export async function updateCategory(
  categoryId: number,
  paperId: PaperId,
  body: UpdateCategoryBody
): Promise<void> {
  const { $api } = useNuxtApp()

  const { data: categories } = useNuxtData<QuestionCategory[]>(
    AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)
  )

  const previousCategories = categories.value!

  categories.value = categories.value!.map(category =>
    category.id === categoryId ? { ...category, ...body } : category
  )

  try {
    await $api(`/api/categories/${categoryId}`, {
      method: 'PATCH',
      body,
    })

    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
  } catch {
    categories.value = previousCategories
  }
}
