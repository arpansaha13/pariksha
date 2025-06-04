export async function upsertPaperTestCases(
  questionId: QuestionId,
  testCases: ReadonlyArray<QuestionCodingTestCase>
): Promise<void> {
  const { $api } = useNuxtApp()

  const { data: questionData } = useNuxtData<Question>(
    UseAsyncDataKeys.paper_question(questionId)
  )

  if (questionData.value?.type !== QuestionType.CODING) {
    logWarning('upsertPaperTestCases called without QuestionType.CODING')
    return
  }

  await $api(`/api/questions/${questionId}/test-cases`, {
    method: 'PUT',
    body: {
      test_cases: testCases,
    },
  })

  return refreshNuxtData(UseAsyncDataKeys.paper_question(questionId))
}
