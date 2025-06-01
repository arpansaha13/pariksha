export function useQuestionCodingBoilerplate(
  questionId: QuestionId | ComputedRef<QuestionId | null>,
  languageId: LanguageId | ComputedRef<LanguageId | null>
) {
  const { $api } = useNuxtApp()

  return useAsyncData<Boilerplate>(
    () => UseAsyncDataKeys.boilerplate(unref(questionId), unref(languageId)),
    async () => {
      if (unref(questionId) === QUESTION_ID_ADD) return Promise.resolve(null)

      const data = await $api<Boilerplate>(
        `/api/questions/${unref(questionId)}/languages/${unref(languageId)}/boilerplate`
      )

      if (typeof data === 'string') {
        return JSON.parse(data)
      }

      return data
    }
  )
}
