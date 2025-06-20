interface UpsertAnswerBody {
  question_id: QuestionId
  /** `null` answers clears the saved answer */
  answer: MCQAnswer | SubjectiveAnswer | CodingAnswer | null
}

export async function upsertAnswer(examId: ExamId, body: UpsertAnswerBody) {
  const { $api } = useNuxtApp()

  return $api<AnswerMinimal>(`/api/exams/${examId}/answers`, {
    method: 'POST',
    body,
  })
}
