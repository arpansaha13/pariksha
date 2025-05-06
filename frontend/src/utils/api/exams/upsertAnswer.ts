import type { GeneralAnswer, MCQAnswer } from '~/types'

interface UpsertAnswerBody {
  question_id: number
  /** `null` answers clears the saved answer */
  answer: MCQAnswer | GeneralAnswer | null
}

export async function upsertAnswer(examId: number, body: UpsertAnswerBody) {
  await $fetch(`/api/exams/${examId}/answers`, {
    method: 'POST',
    body,
    ...getFetchOptions(),
  })
  return refreshNuxtData(AsyncDataKeys.EXAM_ANSWER(examId, body.question_id))
}
