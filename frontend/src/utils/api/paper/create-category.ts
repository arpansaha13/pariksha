export async function createCategory(paperId: PaperId): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(`/api/papers/${paperId}/categories`, {
    method: 'POST',
  })
  await refreshNuxtData(UseAsyncDataKeys.paper_categories(paperId))
}
