export async function createCategory(paperId: number): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(`/api/papers/${paperId}/categories`, {
    method: 'POST',
  })
  await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_CATEGORIES(paperId))
}
