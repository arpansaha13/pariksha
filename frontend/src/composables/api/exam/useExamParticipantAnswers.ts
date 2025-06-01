export function useExamParticipantAnswers(participantId: number) {
  const { $api } = useNuxtApp()

  return useAsyncData(
    UseAsyncDataKeys.exam_participant_answers(participantId),
    () => $api<QuestionAnswer[]>(`/api/participants/${participantId}/answers`),
    {
      transform: questionAnswers => {
        const byCategory = {} as Record<number, QuestionAnswer[]>

        for (const item of questionAnswers) {
          const categoryId = item.question.category_id
          if (!byCategory[categoryId]) {
            byCategory[categoryId] = []
          }
          byCategory[categoryId].push(item)
        }

        return byCategory
      },
    }
  )
}
