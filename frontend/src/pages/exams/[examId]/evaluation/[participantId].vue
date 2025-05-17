<template>
  <div v-if="exam" class="col-span-2 flex items-center gap-2">
    <Icon name="i-heroicons-document-text" size="2rem" />
    <h1 class="text-xl font-semibold">{{ exam.title }}</h1>
  </div>

  <div class="flex items-center justify-end gap-2.5">
    <UButton
      label="Submit evaluation"
      loading-auto
      @click="handleEvaluationSubmit"
    />
  </div>

  <div class="col-span-2 flex h-full flex-col gap-y-4">
    <ExamCategoryNavigation
      v-if="!isNullOrUndefined(sortedCategories)"
      :sorted-categories="sortedCategories"
      :get-question-id-for-category-id="getQuestionIdForCategoryId"
    />
  </div>

  <div class="col-start-3 row-span-2 row-start-2">
    <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

    <UCard v-if="currentCategoryQuestions">
      <EvaluationQuestionList
        v-if="!isNullOrUndefined(currentQuestionId)"
        :current-question-id="currentQuestionId"
        :current-category-questions="currentCategoryQuestions"
      />
    </UCard>
  </div>

  <div
    v-if="currentQuestionAnswer && currentQuestionId"
    class="col-span-2 flex flex-col gap-y-2.5"
  >
    <UCard>
      <p class="font-medium">
        {{ currentQuestionAnswer.question.content.statement }}
      </p>
    </UCard>

    <UCard
      :ui="{
        root:
          isNullOrUndefined(currentQuestionAnswer.answer?.content) &&
          currentQuestionAnswer.type !== QuestionType.MCQ &&
          'grow',
      }"
    >
      <template v-if="currentQuestionAnswer.type === QuestionType.MCQ">
        <URadioGroup
          v-if="!isNullOrUndefined(currentQuestionMcqOptions)"
          v-model="selectedOptionIndex"
          :items="currentQuestionMcqOptions"
          variant="card"
          disabled
          :ui="{
            wrapper: 'ml-3',
            fieldset: 'space-y-1',
            label: 'opacity-100',
            item: 'opacity-100',
          }"
        />
      </template>

      <template v-else>
        <EvaluationUnanswered
          v-if="isNullOrUndefined(currentQuestionAnswer.answer?.content)"
        />

        <p v-else>
          {{ (currentQuestionAnswer.answer.content as SubjectiveAnswer).text }}
        </p>
      </template>
    </UCard>

    <!-- Show evaluation section for MCQ because EvaluationUnanswered for MCQ is shown here -->
    <UCard
      v-if="
        !isNullOrUndefined(currentQuestionAnswer.answer?.content) ||
        currentQuestionAnswer.type === QuestionType.MCQ
      "
      :ui="{ root: 'grow' }"
    >
      <EvaluationUnanswered
        v-if="isNullOrUndefined(currentQuestionAnswer.answer?.content)"
      />
      <UFormField
        v-else
        label="Score"
        description="Score to be awarded for this answer"
        name="score_awarded"
        required
      >
        <UInputNumber
          v-model="
            evaluationStates[currentQuestionAnswer.answer.id].score_awarded
          "
          :min="0"
          :max="currentQuestionAnswer.question.max_score"
          required
        />
      </UFormField>
    </UCard>
  </div>

  <UCard
    v-if="currentCategoryQuestions.length > 1"
    :ui="{ root: 'col-span-2', body: 'flex' }"
  >
    <UButton
      v-if="prevQuestionId"
      label="Previous"
      color="neutral"
      variant="outline"
      :to="{ query: { ...route.query, question: prevQuestionId } }"
      replace
    />
    <UButton
      v-if="nextQuestionId"
      :to="{ query: { ...route.query, question: nextQuestionId } }"
      label="Save and next"
      class="ml-auto"
      replace
    />
  </UCard>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'
import {
  QuestionType,
  type EvaluationAnswer,
  type ExamPermission,
  type SubjectiveAnswer,
  type MCQAnswer,
  type QuestionMcq,
  type Answer,
} from '~/types'

definePageMeta({
  layout: 'paper',
  middleware: [
    'check-exam-permission',
    to => {
      const examId = parseInt(to.params.examId as string)
      const { data: examPermission } = useNuxtData<ExamPermission>(
        AsyncDataKeys.EXAM_PERMISSION(examId)
      )
      if (!examPermission.value!.can_evaluate) {
        return abortNavigation({
          statusCode: HttpStatus.FORBIDDEN,
          message: 'You are not authorized for evaluation in this exam.',
        })
      }
    },
  ],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)
const participantId = parseInt(route.params.participantId as string)

const [
  { data: exam },
  { data: sortedCategories },
  { data: groupedQuestionAnswers },
] = await Promise.all([
  useExam(examId),
  useExamCategories(examId),
  useExamParticipantAnswers(participantId),
])

// _______________LAST VISITED QUESTION FOR CATEGORY________________
function useLastVisitedQuestionForCategory() {
  const lastVisitedQuestionForCategory = ref<Record<number, string>>({})

  watchImmediate(route, newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string)
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  })

  function getQuestionIdForCategoryId(categoryId: number) {
    const categoryItems = groupedQuestionAnswers.value?.[categoryId]
    if (isNullOrUndefined(categoryItems)) return
    const questionId =
      lastVisitedQuestionForCategory.value[categoryId] ??
      categoryItems[0].question.id
    return questionId
  }

  return { getQuestionIdForCategoryId }
}

const { getQuestionIdForCategoryId } = useLastVisitedQuestionForCategory()

// Add initial `category` and `question` queries, if missing
if (!route.query.category && sortedCategories.value?.length) {
  const categoryId = sortedCategories.value[0].id
  const questionId = getQuestionIdForCategoryId(categoryId)
  const query = { category: categoryId, question: questionId }
  await navigateTo({ query }, { replace: true })
}

// ________________________ROUTE QUERY DATA_________________________
const currentCategoryId = computed(() => {
  return route.query.category ? parseInt(route.query.category as string) : null
})

const currentQuestionId = computed(() => {
  return route.query.question ? parseInt(route.query.question as string) : null
})

const currentCategoryQuestions = computed(() => {
  if (!groupedQuestionAnswers.value || !currentCategoryId.value) return []
  return groupedQuestionAnswers.value[currentCategoryId.value] ?? []
})

// ______________________QUESTION NAVIGATION________________________
function useEvaluationQuestionNavigation() {
  const currentQuestionIdx = computed(() => {
    if (!currentCategoryQuestions.value || !currentQuestionId.value) return -1
    return currentCategoryQuestions.value.findIndex(
      item => item.question.id === currentQuestionId.value
    )
  })

  const prevQuestionId = computed(() => {
    if (!currentCategoryQuestions.value || currentQuestionIdx.value <= 0)
      return null
    return currentCategoryQuestions.value[currentQuestionIdx.value - 1].question
      .id
  })

  const nextQuestionId = computed(() => {
    if (
      !currentCategoryQuestions.value ||
      currentQuestionIdx.value === -1 ||
      currentQuestionIdx.value >= currentCategoryQuestions.value.length - 1
    ) {
      return null
    }

    return currentCategoryQuestions.value[currentQuestionIdx.value + 1].question
      .id
  })

  return { prevQuestionId, nextQuestionId, currentQuestionIdx }
}

const { prevQuestionId, nextQuestionId, currentQuestionIdx } =
  useEvaluationQuestionNavigation()

// ___________________HELPER COMPUTED PROPERTIES____________________
const currentQuestionAnswer = computed(() => {
  const qIdx = currentQuestionIdx.value
  if (qIdx === -1) return null
  return currentCategoryQuestions.value[qIdx]
})

// ____________________PREPARE EVALUATION STATES____________________
function useEvaluationStates() {
  const evaluationFetched = ref<Record<string, boolean>>({})
  const evaluationStates = reactive<
    Record<Answer['id'], Partial<EvaluationAnswer>>
  >({})
  const savedEvaluationStates = shallowRef<
    Record<Answer['id'], Partial<EvaluationAnswer>>
  >({})

  for (const items of Object.values(groupedQuestionAnswers.value!)) {
    for (const item of items) {
      if (isNullOrUndefined(item.answer)) continue

      const ansId = item.answer.id
      evaluationStates[ansId] = { score_awarded: undefined }
      savedEvaluationStates.value[ansId] = { ...evaluationStates[ansId] }
    }
  }

  const { $api } = useNuxtApp()
  watchImmediate(currentQuestionAnswer, async qa => {
    if (isNullOrUndefined(qa?.answer)) return

    const qid = qa.question.id
    const ansId = qa.answer.id

    if (evaluationFetched.value[ansId]) return
    evaluationFetched.value[ansId] = true

    const data = await $api<EvaluationAnswer>(
      `/api/participants/${participantId}/questions/${qid}/evaluation-data`
    )

    if (data.question_id !== qid) {
      evaluationFetched.value[ansId] = false
      return
    }

    evaluationStates[ansId].score_awarded = data.score_awarded
    savedEvaluationStates.value[ansId] = { ...evaluationStates[ansId] }
  })

  return { evaluationStates, savedEvaluationStates }
}

const { evaluationStates, savedEvaluationStates } = useEvaluationStates()

// ____________________PERIODIC SAVE EVALUATION_____________________
function usePeriodicSaveEvaluation() {
  /** Save answer for a specific question */
  function saveEvaluation(
    answerId: number,
    savedState: Partial<EvaluationAnswer>,
    currState: Partial<EvaluationAnswer>
  ) {
    if (currState.score_awarded === savedState.score_awarded) {
      return Promise.resolve(null)
    }

    const updateEvaluationBody = {
      new_score: currState.score_awarded,
      evaluated: true,
    }

    return updateAnswerEvaluation(answerId, updateEvaluationBody)
  }

  function saveUpdatedEvaluation() {
    const promises = []

    for (const [ansId, currState] of Object.entries(evaluationStates)) {
      const savedState = savedEvaluationStates.value[parseInt(ansId)]
      if (savedState.score_awarded !== currState.score_awarded) {
        promises.push(
          saveEvaluation(parseInt(ansId), savedState, currState).then(res => {
            if (isNullOrUndefined(res)) return
            savedState.score_awarded = res.score_awarded
          })
        )
      }
    }

    return promises
  }
  useIntervalFn(
    saveUpdatedEvaluation,
    AUTO_SAVE_EVALUATION_INTERVAL_SECONDS * 1000
  )

  return { saveUpdatedEvaluation }
}

const { saveUpdatedEvaluation } = usePeriodicSaveEvaluation()

// ______________________MCQ QUESTIONS DISPLAY______________________
const selectedOptionIndex = ref<number>()

watchImmediate(currentQuestionAnswer, val => {
  if (
    isNullOrUndefined(val?.answer?.content) ||
    val.type !== QuestionType.MCQ
  ) {
    selectedOptionIndex.value = undefined
  } else {
    selectedOptionIndex.value = (val.answer.content as MCQAnswer).optionIndex
  }
})

const currentQuestionMcqOptions = computed(() => {
  if (
    isNullOrUndefined(currentQuestionAnswer.value) ||
    currentQuestionAnswer.value.type !== QuestionType.MCQ
  ) {
    return null
  }

  const mcqQuestion = currentQuestionAnswer.value.question
    .content as QuestionMcq['question']
  return mcqQuestion.options.map((option, i) => ({
    value: i,
    label: option,
  }))
})

// ________________________SUBMIT EVALUATION________________________
const toast = useToast()
async function handleEvaluationSubmit() {
  try {
    await Promise.all(saveUpdatedEvaluation()) // Save any remaining unsaved evaluation data
    await markParticipantAsEvaluated(participantId)
    await navigateTo(`/exams/${examId}`)
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
  } catch (err: any) {
    if (
      err.statusCode === HttpStatus.BAD_REQUEST &&
      err.statusMessage === NuxtErrorStatusMessage.INCOMPLETE_EVALUATION
    ) {
      const toastDescription =
        err.data.unevaluated_count === 1
          ? '1 answer still needs evaluation.'
          : `${err.data.unevaluated_count} answers still need evaluation.`

      toast.add({
        id: ToastId.INCOMPLETE_EVALUATION,
        title: 'Failed to submit evaluation!',
        description: toastDescription,
        color: 'error',
      })
    }
  }
}
</script>
