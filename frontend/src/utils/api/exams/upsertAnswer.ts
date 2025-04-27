import type { GeneralAnswer, MCQAnswer } from '~/types'

interface UpsertAnswerBody {
  question_id: number
  answer: MCQAnswer | GeneralAnswer
}

export async function upsertAnswer(examId: number, body: UpsertAnswerBody) {
  console.log(body)
  return $fetch(`/api/exams/${examId}/answers`, {
    method: 'POST',
    body,
    ...getFetchOptions(),
  })
}
