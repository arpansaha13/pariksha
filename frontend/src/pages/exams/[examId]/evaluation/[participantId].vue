<template>
  <NuxtLayout name="evaluation">
    <template #title>
      <h1 class="text-lg font-semibold sm:text-xl">{{ exam.title }}</h1>
    </template>

    <template #submit>
      <UButton
        label="Submit"
        icon="heroicons:cloud-arrow-up"
        loading-auto
        class="hidden sm:flex"
        @click="handleEvaluationSubmit"
      />
      <UButton
        icon="heroicons:cloud-arrow-up"
        size="sm"
        loading-auto
        class="sm:hidden"
        @click="handleEvaluationSubmit"
      />
    </template>

    <template #category-nav v-if="currentCategoryQuestions">
      <ExamCategoryNavigation
        v-if="!isNullOrUndefined(sortedCategories)"
        :sorted-categories="sortedCategories"
        :get-question-id-for-category-id="getQuestionIdForCategoryId"
      />
    </template>

    <template #question-nav v-if="currentCategoryQuestions">
      <EvaluationQuestionList
        v-if="!isNullOrUndefined(currentQuestionId)"
        :current-question-id="currentQuestionId"
        :current-category-questions="currentCategoryQuestions"
      />
    </template>
    <template #body v-if="currentQuestionAnswer && currentQuestionId">
      <UCard>
        <p
          v-if="currentQuestionAnswer.type === QuestionType.CODING"
          class="mb-2 text-lg font-bold"
        >
          {{ currentQuestionAnswer.question.content.title }}
        </p>
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

          <p v-else-if="currentQuestionAnswer.type === QuestionType.SUBJECTIVE">
            {{
              (currentQuestionAnswer.answer.content as SubjectiveAnswer).text
            }}
          </p>

          <Shiki
            v-else-if="currentQuestionAnswer.type === QuestionType.CODING"
            :code="(currentQuestionAnswer.answer.content as CodingAnswer).code"
            class="text-sm"
          />
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
    </template>

    <template #footer v-if="currentCategoryQuestions.length > 1">
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
        label="Next"
        variant="subtle"
        class="ml-auto"
        replace
      />
    </template>
  </NuxtLayout>
</template>

<script lang="ts" setup>
import { ConfirmModal } from '#components'
import { isNullOrUndefined } from '@arpansaha13/utils'

definePageMeta({
  layout: false,
  middleware: [
    'check-exam-permission',
    to => {
      const examId = to.params.examId as ExamId
      const { data: examPermission } = useNuxtData<ExamPermission>(
        UseAsyncDataKeys.exam_permission(examId)
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
const examId = route.params.examId as ExamId
const participantId = parseInt(
  route.params.participantId as string
) as ExamParticipantId

// If participant has not ended their exam, or if they are already evaluated
// then disallow access to this route
await callOnce(
  async () => {
    const { data: participant } = await useExamParticipantById(participantId)
    if (
      isNullOrUndefined(participant.value) ||
      participant.value.status !== ExamParticipantStatus.ENDED
    ) {
      throw createError({
        statusCode: HttpStatus.FORBIDDEN,
        message: 'Participant cannot be evaluated now.',
      })
    }
  },
  { mode: 'navigation' }
)

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
  const lastVisitedQuestionForCategory = ref<Record<CategoryId, string>>({})

  watchImmediate(route, newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string) as CategoryId
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  })

  function getQuestionIdForCategoryId(categoryId: CategoryId) {
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
  return route.query.category
    ? (parseInt(route.query.category as string) as CategoryId)
    : null
})

const currentQuestionId = computed(() => {
  return route.query.question ? (route.query.question as QuestionId) : null
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
  const evaluationFetched = ref<Record<AnswerId, boolean>>({})
  const evaluationStates = reactive<
    Record<AnswerId, Partial<EvaluationAnswer>>
  >({})
  const savedEvaluationStates = shallowRef<
    Record<AnswerId, Partial<EvaluationAnswer>>
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
      const savedState =
        savedEvaluationStates.value[ansId as unknown as AnswerId]
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
    .content as QuestionMcqContent
  return mcqQuestion.options.map((option, i) => ({
    value: i,
    label: option,
  }))
})

// ________________________SUBMIT EVALUATION________________________
const toast = useToast()
const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal)

async function handleEvaluationSubmit() {
  const instance = confirmModal.open({
    title: 'Submit evaluation',
    description:
      'Once submitted, this evaluation cannot be modified. Please review carefully before proceeding.',
    confirmLabel: 'Submit',
    variant: 'primary',
  })

  const shouldSubmit = await instance.result
  if (!shouldSubmit) return

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
