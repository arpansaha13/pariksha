import { isNullOrUndefined } from '@arpansaha13/utils'
import type { ExamQuestionMinimal } from '~/types'

interface UseExamQuestionIdForCategoryIdArgs {
  groupedQuestions: Ref<Record<number, ExamQuestionMinimal[]> | null>
}

export function useExamQuestionIdForCategoryId(
  args: UseExamQuestionIdForCategoryIdArgs
) {
  const { groupedQuestions } = args

  const route = useRoute()
  const lastVisitedQuestionForCategory = ref<Record<number, string>>({})

  watchImmediate(route, newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string)
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  })

  function getQuestionIdForCategoryId(categoryId: number) {
    const categoryQuestions = groupedQuestions.value?.[categoryId]
    if (isNullOrUndefined(categoryQuestions)) return
    const questionId =
      lastVisitedQuestionForCategory.value[categoryId] ??
      categoryQuestions[0].id
    return questionId
  }

  return getQuestionIdForCategoryId
}
