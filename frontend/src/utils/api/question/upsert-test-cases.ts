export async function upsertPaperTestCases(
  questionId: QuestionId,
  testCases: ReadonlyArray<PartialQuestionCodingTestCase>
): Promise<void> {
  const { $api } = useNuxtApp()

  // Check for duplicate IDs on client side as well
  const seenIds = new Set<TestCaseId>()
  for (const tc of testCases) {
    if (tc.id) {
      if (seenIds.has(tc.id)) {
        logWarning('Duplicate test case IDs found in upsertPaperTestCases')
        return
      }
      seenIds.add(tc.id)
    }
  }

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
