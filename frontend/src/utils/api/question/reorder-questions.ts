export async function reorderQuestions(
  paperId: PaperId,
  categoryId: number,
  questionIds: number[]
): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(`/api/categories/${categoryId}/questions/reorder`, {
    method: 'PATCH',
    body: { questions: questionIds },
  })

  await refreshNuxtData(UseAsyncDataKeys.paper_questions(paperId))
}
