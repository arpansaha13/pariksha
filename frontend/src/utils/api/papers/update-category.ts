interface UpdateCategoryBody {
  name: string
}

export async function updateCategory(
  categoryId: number,
  paperId: number,
  body: UpdateCategoryBody
): Promise<void> {
  await $fetch(`/api/categories/${categoryId}`, {
    method: 'PATCH',
    body,
    ...getFetchOptions(),
  })
  await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
}
