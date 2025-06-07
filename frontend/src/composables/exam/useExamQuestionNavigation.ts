interface UseExamQuestionNavigationOptions {
  currentQuestionId: ComputedRef<QuestionId | null>
  currentCategoryQuestions: ComputedRef<ExamQuestionMinimal[]>
}

export function useExamQuestionNavigation(
  args: UseExamQuestionNavigationOptions
) {
  const { currentQuestionId, currentCategoryQuestions } = args

  const currentQuestionIdx = computed(() => {
    if (!currentCategoryQuestions.value || !currentQuestionId.value) return -1
    return currentCategoryQuestions.value.findIndex(
      q => q.id === currentQuestionId.value
    )
  })

  const prevQuestionId = computed(() => {
    if (!currentCategoryQuestions.value || currentQuestionIdx.value <= 0)
      return null
    return currentCategoryQuestions.value[currentQuestionIdx.value - 1].id
  })

  const nextQuestionId = computed(() => {
    if (
      !currentCategoryQuestions.value ||
      currentQuestionIdx.value === -1 ||
      currentQuestionIdx.value >= currentCategoryQuestions.value.length - 1
    ) {
      return null
    }

    return currentCategoryQuestions.value[currentQuestionIdx.value + 1].id
  })

  return { prevQuestionId, nextQuestionId, currentQuestionIdx }
}
