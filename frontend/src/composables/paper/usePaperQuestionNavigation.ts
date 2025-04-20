import { QuestionId, type QuestionMinimal } from '~/types'

interface UsePaperQuestionNavigationArgs {
  currentQuestionId: ComputedRef<number | null>
  currentCategoryQuestions: ComputedRef<QuestionMinimal[]>
}

export function usePaperQuestionNavigation(
  args: UsePaperQuestionNavigationArgs
) {
  enum QuestionIndex {
    NON_EXISTENT = -1,

    /** Special case for add question */
    ADD = -2,
  }

  const { currentQuestionId, currentCategoryQuestions } = args

  const currentQuestionIdx = computed(() => {
    if (!currentQuestionId.value) return QuestionIndex.NON_EXISTENT
    if (currentQuestionId.value === QuestionId.ADD) return QuestionIndex.ADD
    return (
      currentCategoryQuestions.value?.findIndex(
        q => q.id === currentQuestionId.value
      ) ?? QuestionIndex.NON_EXISTENT
    )
  })

  const questionNavigation = computed(() => {
    if (!currentCategoryQuestions.value) {
      return { prev: null, next: null }
    }

    // If on add question page
    if (currentQuestionId.value === QuestionId.ADD) {
      return {
        prev: currentCategoryQuestions.value.at(-1)?.id ?? null, // show last question as prev
        next: null, // there is no next
      }
    }

    // Non-existent question check should be after QuestionId.ADD
    if (currentQuestionIdx.value < 0) {
      return { prev: null, next: null }
    }

    // First question
    if (currentQuestionIdx.value === 0) {
      return {
        prev: null,
        next: currentCategoryQuestions.value[1]?.id ?? QuestionId.ADD,
      }
    }

    // Last question
    if (
      currentQuestionIdx.value ===
      currentCategoryQuestions.value.length - 1
    ) {
      return {
        prev: currentCategoryQuestions.value.at(-2)?.id ?? null,
        next: QuestionId.ADD, // show add question as next
      }
    }

    return {
      prev: currentCategoryQuestions.value[currentQuestionIdx.value - 1].id,
      next: currentCategoryQuestions.value[currentQuestionIdx.value + 1].id,
    }
  })

  return { questionNavigation }
}
