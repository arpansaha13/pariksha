export async function deleteQuestion(
  questionId: number,
  paperId: PaperId,
  categoryId: number
): Promise<void> {
  const { $api } = useNuxtApp()

  const { data: groupedQuestions } = useNuxtData<
    Record<number, QuestionMinimal[]>
  >(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId))

  // Store minimal data for potential rollback
  const categoryQuestions = groupedQuestions.value![categoryId]
  const questionIdx = categoryQuestions.findIndex(q => q.id === questionId)
  const deletedQuestion = categoryQuestions[questionIdx]

  // Optimistically remove the question
  categoryQuestions.splice(questionIdx, 1)

  try {
    await $api(`/api/questions/${questionId}`, {
      method: 'DELETE',
    })

    // Refresh paper data since max_score and question_counts change
    await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER(paperId))
  } catch {
    // Rollback on error
    categoryQuestions.splice(questionIdx, 0, deletedQuestion)
  }
}
