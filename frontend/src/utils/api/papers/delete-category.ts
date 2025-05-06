export async function deleteCategory(
  categoryId: number,
  paperId: number
): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(`/api/categories/${categoryId}`, {
    method: 'DELETE',
  })
  await Promise.all([
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId)),
    refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId)),
  ])
}
