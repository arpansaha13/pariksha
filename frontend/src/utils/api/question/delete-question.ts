export async function deleteQuestion(
  questionId: QuestionId,
  paperId: PaperId,
  categoryId: CategoryId
): Promise<void> {
  const { $api } = useNuxtApp()

  const { data: groupedQuestions } = useNuxtData<
    Record<number, QuestionMinimal[]>
  >(UseAsyncDataKeys.paper_questions(paperId))

  // Store minimal data for potential rollback
  const categoryQuestions = groupedQuestions.value![categoryId]
  const questionIdx = categoryQuestions.findIndex(q => q.id === questionId)
  const deletedQuestion = categoryQuestions[questionIdx]

  // Optimistically remove the question
  categoryQuestions.splice(questionIdx, 1)

  try {
    await $api(`/api/paper/${paperId}/questions/${questionId}`, {
      method: 'DELETE',
    })

    // Refresh paper data since max_score and question_counts change
    await refreshNuxtData(UseAsyncDataKeys.paper(paperId))
  } catch {
    // Rollback on error
    categoryQuestions.splice(questionIdx, 0, deletedQuestion)
  }
}
