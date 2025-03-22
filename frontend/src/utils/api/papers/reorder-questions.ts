export async function reorderQuestions(
  paperId: number,
  categoryId: number,
  questionIds: number[]
): Promise<void> {
  await $fetch(`/api/categories/${categoryId}/questions/reorder`, {
    method: 'PATCH',
    body: { questions: questionIds },
    ...getFetchOptions(),
  })

  await refreshNuxtData(AsyncDataKeys.PAPERS_PAPER_QUESTIONS(paperId))
}
