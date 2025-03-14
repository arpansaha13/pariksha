export async function createCategory(paperId: number): Promise<void> {
  await $fetch(`/api/papers/${paperId}/categories`, {
    method: 'POST',
    ...getFetchOptions(),
  })
  await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
}
