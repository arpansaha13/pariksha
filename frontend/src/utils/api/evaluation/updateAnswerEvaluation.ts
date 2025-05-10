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

  await $api(`/api/answers/${answerId}`, {
    method: 'PATCH',
    body,
  })
}
