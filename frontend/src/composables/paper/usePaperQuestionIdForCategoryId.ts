import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type QuestionMinimal } from '~/types'

interface UsePaperQuestionIdForCategoryIdArgs {
  groupedQuestions: Ref<Record<number, QuestionMinimal[]> | null>
}

export function usePaperQuestionIdForCategoryId(
  args: UsePaperQuestionIdForCategoryIdArgs
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
    if (isNullOrUndefined(categoryQuestions)) return QuestionId.ADD
    const questionId =
      lastVisitedQuestionForCategory.value[categoryId] ??
      categoryQuestions[0].id
    return questionId
  }

  return getQuestionIdForCategoryId
}
