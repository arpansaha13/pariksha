export async function deleteCategory(
  categoryId: number,
  paperId: PaperId
): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(`/api/categories/${categoryId}`, {
    method: 'DELETE',
  })
  await Promise.all([
    refreshNuxtData(UseAsyncDataKeys.paper_categories(paperId)),
    refreshNuxtData(UseAsyncDataKeys.paper_questions(paperId)),
  ])
}
