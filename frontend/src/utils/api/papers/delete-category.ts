export async function deleteCategory(
  categoryId: number,
  paperId: number
): Promise<void> {
  await $fetch(`/api/categories/${categoryId}`, {
    method: 'DELETE',
    ...getFetchOptions(),
  })
  await Promise.all([
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)),
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
  ])
}
