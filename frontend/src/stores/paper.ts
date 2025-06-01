import { defineStore } from 'pinia'
import { isNullOrUndefined } from '@arpansaha13/utils'

interface PaperStore {
  unsavedCount: Record<CategoryId, number>
  lastVisitedQuestionForCategory: Record<CategoryId, string>
}

export const usePaperStore = defineStore(paperStoreId, {
  state: (): PaperStore => ({
    unsavedCount: {},
    lastVisitedQuestionForCategory: {},
  }),
  getters: {
    getQuestionIdForCategoryId() {
      const route = useRoute()
      const paperId = route.params.paperId as PaperId

      const { data: groupedQuestions } = useNuxtData<
        Record<number, QuestionMinimal[]>
      >(UseAsyncDataKeys.paper_questions(paperId))

      return (categoryId: CategoryId) => {
        const categoryQuestions = groupedQuestions.value?.[categoryId]
        if (isNullOrUndefined(categoryQuestions)) return QUESTION_ID_ADD
        const questionId =
          this.lastVisitedQuestionForCategory[categoryId] ??
          categoryQuestions[0].id
        return questionId
      }
    },
  },
  actions: {
    incUnsavedCount(categoryId: CategoryId) {
      if (!this.unsavedCount[categoryId]) {
        this.unsavedCount[categoryId] = 1
      } else {
        this.unsavedCount[categoryId]++
      }
    },
    decUnsavedCount(categoryId: CategoryId) {
      if (!this.unsavedCount[categoryId]) {
        this.unsavedCount[categoryId] = 0
      } else {
        this.unsavedCount[categoryId]--
      }
    },
  },
})
