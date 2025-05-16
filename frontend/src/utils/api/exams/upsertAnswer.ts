import type { AnswerMinimal, GeneralAnswer, MCQAnswer } from '~/types'

interface UpsertAnswerBody {
  question_id: number
  /** `null` answers clears the saved answer */
  answer: MCQAnswer | GeneralAnswer | null
}

export async function upsertAnswer(examId: number, body: UpsertAnswerBody) {
  const { $api } = useNuxtApp()

  return $api<AnswerMinimal>(`/api/exams/${examId}/answers`, {
    method: 'POST',
    body,
  })
}
