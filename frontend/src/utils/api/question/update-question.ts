import { isEqual } from 'lodash-es'
import { isNullOrUndefined } from '@arpansaha13/utils'
import { extractQuestionContent, type MergedQuestion } from './utils'

type UpdateQuestionBody = Partial<Omit<Question, 'category_id'>>
type UpdateQuestionReturn = Pick<Question, 'id'>
type MergedQuestionOmitCategory = Omit<MergedQuestion, 'category_id'>

export async function updateQuestion(
  questionId: QuestionId,
  paperId: PaperId,
  mergedQuestion: MergedQuestionOmitCategory
) {
  const { $api } = useNuxtApp()

  const { data: question } = useNuxtData<Question>(
    UseAsyncDataKeys.paper_question(questionId)
  )

  const previousQuestion = question.value!
  const requestBody = getRequestBody(previousQuestion, mergedQuestion)

  // If no changes, return early
  if (Object.keys(requestBody).length === 0) return

  const res = await $api<UpdateQuestionReturn>(
    `/api/papers/${paperId}/questions/${questionId}`,
    {
      method: 'PATCH',
      body: requestBody,
    }
  )

  // Update the particular question in paperQuestions
  const { data: groupedQuestions } = useNuxtData<
    Record<number, QuestionMinimal[]>
  >(UseAsyncDataKeys.paper_questions(paperId))

  const categoryQuestions =
    groupedQuestions.value![previousQuestion.category_id]
  const minimalQuestion = categoryQuestions?.find(q => q.id === questionId)
  if (minimalQuestion && requestBody.question) {
    minimalQuestion.question = requestBody.question
  }

  const refreshPromises = []

  if (res.id === questionId) {
    refreshPromises.push(
      refreshNuxtData(UseAsyncDataKeys.paper_question(questionId))
    )
  } else {
    // If a locked question is updated, a new questionId will be returned
    const route = useRoute()
    const replaceOldQuestionIdWithNew = async () => {
      await navigateTo(
        { query: { ...route.query, question: res.id } }, // update the route query
        { replace: true }
      )
      if (minimalQuestion) minimalQuestion.id = res.id // update questionId in groupedQuestions list
      clearNuxtData(UseAsyncDataKeys.paper_question(questionId)) // clear old question data
    }
    refreshPromises.push(replaceOldQuestionIdWithNew())
  }

  // requestBody will only have the updated fields
  const isMaxScoreUpdated = !isNullOrUndefined(requestBody.max_score)
  const isTypeUpdated = !isNullOrUndefined(requestBody.type)

  // Refresh paper data if max_score or type changed
  if (isMaxScoreUpdated || isTypeUpdated) {
    refreshPromises.push(refreshNuxtData(UseAsyncDataKeys.paper(paperId)))
  }

  return Promise.all(refreshPromises)
}

/** Only include fields that are updated */
function getRequestBody(
  previousQuestion: Readonly<Question>,
  mergedQuestion: MergedQuestionOmitCategory
): UpdateQuestionBody {
  const requestBody: UpdateQuestionBody = {}

  const isTypeUpdated =
    !isNullOrUndefined(mergedQuestion.type) &&
    mergedQuestion.type !== previousQuestion.type
  if (isTypeUpdated) {
    if (isNullOrUndefined(mergedQuestion.question)) {
      logWarning('Must include question data when type changes')
      return {}
    }
    requestBody.type = mergedQuestion.type
    requestBody.question = extractQuestionContent(mergedQuestion)!
  } else if (
    !isNullOrUndefined(mergedQuestion.question) &&
    isQuestionContentUpdated(
      mergedQuestion.type,
      mergedQuestion,
      previousQuestion
    )
  ) {
    requestBody.question = extractQuestionContent(mergedQuestion)!
  }

  if (
    !isNullOrUndefined(mergedQuestion.max_score) &&
    mergedQuestion.max_score !== previousQuestion.max_score
  ) {
    requestBody.max_score = mergedQuestion.max_score
  }

  // if (
  //   !isNullOrUndefined(mergedQuestion.tags) &&
  //   !isEqual(mergedQuestion.tags, previousQuestion.tags)
  // ) {
  //   requestBody.tags = mergedQuestion.tags
  // }

  // if (
  //   !isNullOrUndefined(mergedQuestion.correct_answer) &&
  //   mergedQuestion.correct_answer !== previousQuestion.correct_answer
  // ) {
  //   requestBody.correct_answer = mergedQuestion.correct_answer
  // }

  return requestBody
}

const isQuestionContentUpdated = (
  type: QuestionType,
  newQ: MergedQuestion,
  oldQ: Question
) => {
  const newQContent = newQ.question
  const oldQContent = oldQ.question

  if (type === QuestionType.MCQ) {
    return isMcqQuestionContentUpdated(
      newQContent,
      oldQContent as QuestionMcqContent
    )
  }
  if (type === QuestionType.SUBJECTIVE) {
    return isSubjectiveQuestionContentUpdated(
      newQContent,
      oldQContent as QuestionSubjectiveContent
    )
  }
  if (type === QuestionType.CODING) {
    return isCodingQuestionContentUpdated(
      newQContent,
      oldQContent as QuestionCodingContent
    )
  }
  logWarning(`Invalid mergedQuestion type: ${type}`)
  return false
}

const isMcqQuestionContentUpdated = (
  newQContent: QuestionMcqContent,
  oldQContent: QuestionMcqContent
) => {
  return (
    newQContent.statement.length !== oldQContent.statement.length ||
    newQContent.statement !== oldQContent.statement ||
    !isEqual(newQContent.options ?? [], oldQContent.options ?? [])
  )
}

const isSubjectiveQuestionContentUpdated = (
  newQContent: QuestionSubjectiveContent,
  oldQContent: QuestionSubjectiveContent
) => {
  return (
    newQContent.statement.length !== oldQContent.statement.length ||
    newQContent.statement !== oldQContent.statement
  )
}

const isCodingQuestionContentUpdated = (
  newQContent: QuestionCodingContent,
  oldQContent: QuestionCodingContent
) => {
  return (
    newQContent.title.length !== oldQContent.title.length ||
    newQContent.title !== oldQContent.title ||
    newQContent.statement.length !== oldQContent.statement.length ||
    newQContent.statement !== oldQContent.statement ||
    !isEqual(
      newQContent.output_definition ?? [],
      oldQContent.output_definition ?? []
    ) ||
    !isEqual(
      newQContent.input_definitions ?? [],
      oldQContent.input_definitions ?? []
    )
  )
}
