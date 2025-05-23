import { isNullOrUndefined } from '@arpansaha13/utils'

interface UseExamQuestionIdForCategoryIdArgs {
  groupedQuestions: Ref<Record<CategoryId, ExamQuestionMinimal[]> | null>
}

export function useExamQuestionIdForCategoryId(
  args: UseExamQuestionIdForCategoryIdArgs
) {
  const { groupedQuestions } = args

  const route = useRoute()
  const lastVisitedQuestionForCategory = ref<Record<CategoryId, string>>({})

  watchImmediate(route, newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string) as CategoryId
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  })

  function getQuestionIdForCategoryId(categoryId: CategoryId) {
    const categoryQuestions = groupedQuestions.value?.[categoryId]
    if (isNullOrUndefined(categoryQuestions)) return
    const questionId =
      lastVisitedQuestionForCategory.value[categoryId] ??
      categoryQuestions[0].id
    return questionId
  }

  return getQuestionIdForCategoryId
}
