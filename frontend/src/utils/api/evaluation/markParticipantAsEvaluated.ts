interface EvaluationStatusResponse {
  unevaluated_count: number
}

export async function markParticipantAsEvaluated(
  participantId: number
): Promise<EvaluationStatusResponse> {
  const { $api } = useNuxtApp()

  const response = await $api<EvaluationStatusResponse>(
    `/api/participants/${participantId}/evaluate`,
    {
      method: 'PATCH',
    }
  )

  if (response.unevaluated_count > 0) {
    throw createError({
      statusCode: HttpStatus.BAD_REQUEST,
      statusMessage: NuxtErrorStatusMessage.INCOMPLETE_EVALUATION,
      data: response,
    })
  }

  return response
}
