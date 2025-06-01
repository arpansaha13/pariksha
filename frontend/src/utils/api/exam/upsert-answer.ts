interface UpsertAnswerBody {
  question_id: number
  /** `null` answers clears the saved answer */
  answer: MCQAnswer | SubjectiveAnswer | null
}

export async function upsertAnswer(examId: ExamId, body: UpsertAnswerBody) {
  const { $api } = useNuxtApp()

  return $api<AnswerMinimal>(`/api/exams/${examId}/answers`, {
    method: 'POST',
    body,
  })
}
