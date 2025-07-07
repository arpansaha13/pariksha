export async function reorderQuestions(
  paperId: PaperId,
  categoryId: CategoryId,
  questionIds: QuestionId[]
): Promise<void> {
  const { $api } = useNuxtApp()

  await $api(
    `/api/papers/${paperId}/categories/${categoryId}/questions/reorder`,
    {
      method: 'PATCH',
      body: { questions: questionIds },
    }
  )

  await refreshNuxtData(UseAsyncDataKeys.paper_questions(paperId))
}
