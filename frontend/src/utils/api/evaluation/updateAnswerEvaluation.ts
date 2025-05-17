import type { EvaluationAnswer } from '~/types'

interface UpdateAnswerEvaluationBody {
  new_score?: number
  evaluated?: boolean
  comment?: string
}

export async function updateAnswerEvaluation(
  answerId: number,
  body: UpdateAnswerEvaluationBody
) {
  const { $api } = useNuxtApp()

  return $api<EvaluationAnswer>(`/api/answers/${answerId}`, {
    method: 'PATCH',
    body,
  })
}
