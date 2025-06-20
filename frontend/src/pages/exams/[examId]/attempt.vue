<template>
  <UContainer
    v-if="isExamEnded"
    class="col-span-full row-span-full flex flex-col items-center justify-center"
  >
    <p class="flex items-center gap-2">
      Your exam has been submitted.
      <Icon name="twemoji:party-popper" />
    </p>
    <p class="mt-2.5">
      Redirecting in
      <span class="font-medium">
        {{ redirectCountdown === 0 ? 1 : redirectCountdown }}
      </span>
      seconds
    </p>

    <UButton :to="`/exams/${examId}/results`" class="mt-4">
      Go to results
    </UButton>
  </UContainer>

  <template v-else>
    <div v-if="exam" class="col-span-2 flex items-center gap-2">
      <Icon name="i-heroicons-document-text" size="2rem" />
      <h1 class="text-xl font-semibold">{{ exam.title }}</h1>
    </div>

    <div class="flex items-center justify-end gap-2.5">
      <ExamTimer
        v-if="participant"
        :started-at="participant.started_at"
        :scheduled-end-time="participant.scheduled_end_time"
        @timeout="handleExamSubmit"
      />

      <ExamSubmit @submit="handleExamSubmit" />
    </div>

    <div class="col-span-2 flex h-full flex-col gap-y-4">
      <ExamCategoryNavigation
        v-if="!isNullOrUndefined(sortedCategories)"
        :sorted-categories="sortedCategories"
        :get-question-id-for-category-id="getQuestionIdForCategoryId"
      />
    </div>

    <UCard
      v-if="currentCategoryQuestions"
      :ui="{
        root: 'col-start-3 row-span-2 row-start-2 flex flex-col',
        body: 'grow overflow-auto',
        footer: 'flex',
      }"
    >
      <template #header>
        <h3 class="heading">Question Pallet</h3>
      </template>

      <ExamQuestionList
        v-if="!isNullOrUndefined(currentQuestionId)"
        :current-question-id="currentQuestionId"
        :current-category-questions="currentCategoryQuestions"
      />

      <template #footer v-if="currentCategoryQuestions.length > 1">
        <UButton
          v-if="prevQuestionId"
          :to="{ query: { ...route.query, question: prevQuestionId } }"
          replace
          label="Previous"
          color="neutral"
          variant="outline"
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
    </UCard>

    <div
      v-if="!isNullOrUndefined(currentQuestionType) && question"
      class="col-span-2 -m-[2px] flex flex-col gap-y-2.5 overflow-auto p-[2px]"
    >
      <template v-if="currentQuestionType === QuestionType.MCQ">
        <UCard>
          <p class="font-medium">{{ question.question.statement }}</p>
        </UCard>

        <UCard :ui="{ root: 'grow' }">
          <URadioGroup
            v-model="examStore.mcqAnswerStates[question.id]!.optionIndex"
            :items="mcqOptions"
            variant="card"
            :ui="{
              wrapper: 'ml-3',
              fieldset: 'space-y-1',
            }"
          />

          <UButton
            variant="ghost"
            :disabled="
              isNullOrUndefined(
                examStore.mcqAnswerStates[question.id]!.optionIndex
              )
            "
            :ui="{
              base: 'mt-5',
            }"
            @click="examStore.clearMcqSelection"
          >
            Clear selection
          </UButton>
        </UCard>
      </template>

      <template v-else-if="currentQuestionType === QuestionType.SUBJECTIVE">
        <UCard>
          <p class="font-medium">{{ question.question.statement }}</p>
        </UCard>

        <UCard :ui="{ root: 'grow' }">
          <UTextarea
            v-model="examStore.subjectiveAnswerStates[question.id]!.text"
            autoresize
            rows="4"
            placeholder="Write your answer here..."
            :ui="{ root: 'flex' }"
          />
        </UCard>
      </template>

      <template v-else="currentQuestionType === QuestionType.CODING">
        <UCard
          :ui="{
            root: 'grow flex flex-col',
            body: 'grow',
            footer: 'flex justify-end',
          }"
        >
          <DisplayCodingQuestion
            :content="question.question"
            :test-cases="question.test_cases ?? []"
            :editor-link="`/editor/exams/${exam.id}/questions/${question.id}`"
          />

          <template #footer>
            <UButton
              label="Open in editor"
              variant="subtle"
              :to="`/editor/exams/${exam.id}/questions/${question.id}`"
            />
          </template>
        </UCard>
      </template>
    </div>
  </template>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import { ConfirmModal } from '#components'

definePageMeta({
  layout: 'paper',
  middleware: [
    'check-exam-permission',
    to => {
      const examId = to.params.examId as ExamId
      const { data: examPermission } = useNuxtData<ExamPermission>(
        UseAsyncDataKeys.exam_permission(examId)
      )
      if (!examPermission.value!.can_participate) {
        return abortNavigation({
          statusCode: HttpStatus.FORBIDDEN,
          message: 'You are not registered as a participant for this exam.',
        })
      }
      if (
        examPermission.value!.participant_status === ExamParticipantStatus.ENDED
      ) {
        return abortNavigation({
          statusCode: HttpStatus.FORBIDDEN,
          message: 'You have already attempted this exam',
        })
      }
    },
  ],
})

const route = useRoute()
const examId = route.params.examId as ExamId

const examStore = useExamStore()

// If the examId doesn't match, reset the store for the new exam
if (examStore.examId && examStore.examId !== examId) {
  examStore.$reset()
}

examStore.examId = examId

const [
  { data: exam },
  { data: participant },
  { data: groupedQuestions },
  { data: sortedCategories },
] = await Promise.all([
  useExam(examId),
  useExamParticipant(examId),
  useExamQuestions(examId),
  useExamCategories(examId),
])

examStore.prepare(groupedQuestions.value!)

const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal as Component)
provide(InjectionKeys.ConfirmModal, confirmModal)

const getQuestionIdForCategoryId = useExamQuestionIdForCategoryId({
  groupedQuestions,
})

// Add initial `category` and `question` queries, if missing
if (!route.query.category && sortedCategories.value?.length) {
  const categoryId = sortedCategories.value[0].id
  const questionId = getQuestionIdForCategoryId(categoryId)
  const query = { category: categoryId, question: questionId }
  await navigateTo({ query }, { replace: true })
}

const currentCategoryId = computed(() => {
  return route.query.category
    ? (parseInt(route.query.category as string) as CategoryId)
    : null
})

const currentCategoryQuestions = computed(() => {
  if (!groupedQuestions.value || !currentCategoryId.value) return []
  return groupedQuestions.value[currentCategoryId.value] ?? []
})

const currentQuestionId = computed(() => {
  return route.query.question ? (route.query.question as QuestionId) : null
})

const { prevQuestionId, nextQuestionId, currentQuestionIdx } =
  useExamQuestionNavigation({
    currentQuestionId,
    currentCategoryQuestions,
  })

const currentQuestionType = computed(() => {
  const qIdx = currentQuestionIdx.value
  if (qIdx === -1) return null
  return currentCategoryQuestions.value[qIdx].type
})

const { data: question } = await useExamQuestion(currentQuestionId)

const mcqOptions = computed(() => {
  if (
    isNullOrUndefined(question.value) ||
    currentQuestionType.value !== QuestionType.MCQ
  ) {
    return
  }
  return (question.value.question as QuestionMcqContent).options.map(
    (option, i) => ({
      value: i,
      label: option,
    })
  )
})

//__________________________LOAD ANSWER___________________________

const { $api } = useNuxtApp()

watchImmediate(currentQuestionId, async qid => {
  if (isNullOrUndefined(qid)) return

  if (examStore.answerFetched[qid]) return
  examStore.answerFetched[qid] = true

  const data = await $api<AnswerMinimal>(
    `/api/exams/${examId}/questions/${qid}/answer`
  )

  if (isNullOrUndefined(data.answer)) return
  if (data.question_id !== qid) {
    examStore.answerFetched[qid] = false
    return
  }

  storeAnswerFromResponse(qid, data)
})

function storeAnswerFromResponse(
  qid: QuestionId,
  answerResponse: AnswerMinimal
) {
  if (isNullOrUndefined(currentCategoryQuestions.value)) {
    logWarning('currentCategoryQuestions is null or undefined')
    return
  }

  const qIdx = currentQuestionIdx.value
  if (qIdx === -1) {
    logWarning('currentQuestionIdx is -1')
    return
  }

  const qType = currentCategoryQuestions.value[qIdx].type
  examStore.setAnswer(qid, qType, answerResponse.answer)
}

useIntervalFn(
  examStore.saveUpdatedAnswers,
  AUTO_SAVE_EXAM_ANSWER_INTERVAL_SECONDS * 1000
)

// ___________________AUTO-END EXAM ON TIMEOUT____________________
const isExamEnded = ref(false)
const { remaining: redirectCountdown, start: startRedirectCountdown } =
  useCountdown(5, {
    onComplete() {
      navigateTo(`/exams/${examId}/results`)
    },
  })

async function handleExamSubmit() {
  await Promise.all(examStore.saveUpdatedAnswers()) // Save any remaining unsaved answers
  await endExam(examId)
  isExamEnded.value = true
  startRedirectCountdown()
}
</script>
