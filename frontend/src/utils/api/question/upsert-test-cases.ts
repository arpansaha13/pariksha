export async function upsertPaperTestCases(
  questionId: QuestionId,
  testCases: ReadonlyArray<PartialQuestionCodingTestCase>
): Promise<void> {
  const { $api } = useNuxtApp()

  // Check for duplicate IDs on client side as well
  const seenIds = new Set<number>()
  for (const tc of testCases) {
    if (tc.id) {
      if (seenIds.has(tc.id)) {
        throw new Error('Duplicate test case IDs found')
      }
      seenIds.add(tc.id)
    }
  }

  await $api(`/api/questions/${questionId}/test-cases`, {
    method: 'PUT',
    body: {
      test_cases: testCases,
    },
  })

  return refreshNuxtData(UseAsyncDataKeys.paper_question(questionId))
}
