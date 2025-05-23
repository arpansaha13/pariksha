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

    <div class="col-start-3 row-span-2 row-start-2">
      <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

      <UCard v-if="currentCategoryQuestions">
        <ExamQuestionList
          v-if="!isNullOrUndefined(currentQuestionId)"
          :current-question-id="currentQuestionId"
          :current-category-questions="currentCategoryQuestions"
        />
      </UCard>
    </div>

    <div
      v-if="!isNullOrUndefined(currentQuestionType) && question"
      class="col-span-2 flex flex-col gap-y-2.5"
    >
      <template v-if="currentQuestionType === QuestionType.MCQ">
        <UCard>
          <p class="font-medium">{{ question.question.statement }}</p>
        </UCard>

        <UCard :ui="{ root: 'grow' }">
          <URadioGroup
            v-model="mcqAnswerStates[question.id].optionIndex"
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
              isNullOrUndefined(mcqAnswerStates[question.id].optionIndex)
            "
            :ui="{
              base: 'mt-5',
            }"
            @click="clearMcqSelection"
          >
            Clear selection
          </UButton>
        </UCard>
      </template>

      <template v-else>
        <UCard>
          <p class="font-medium">{{ question.question.statement }}</p>
        </UCard>

        <UCard :ui="{ root: 'grow' }">
          <UTextarea
            v-model="subjectiveAnswerStates[question.id].text"
            autoresize
            placeholder="Write your answer here..."
            :ui="{ root: 'flex' }"
          />
        </UCard>
      </template>
    </div>

    <UCard
      v-if="currentCategoryQuestions.length > 1"
      :ui="{ root: 'col-span-2', body: 'flex' }"
    >
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
        label="Save and next"
        class="ml-auto"
        replace
      />
    </UCard>
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
      const examId = parseInt(to.params.examId as string)
      const { data: examPermission } = useNuxtData<ExamPermission>(
        AsyncDataKeys.EXAM_PERMISSION(examId)
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
const examId = parseInt(route.params.examId as string)

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
  return route.query.question
    ? (parseInt(route.query.question as string) as QuestionId)
    : null
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

function clearMcqSelection() {
  if (isNullOrUndefined(currentQuestionId.value)) return
  const qid = currentQuestionId.value
  mcqAnswerStates[qid].optionIndex = undefined
}

//__________________________LOAD ANSWER___________________________
const answerFetched = ref<Record<string, boolean>>({})
const savedMcqAnswerStates = shallowRef<Record<string, MCQAnswer>>({})
const savedSubjectiveAnswerStates = shallowRef<
  Record<string, SubjectiveAnswer>
>({})
const mcqAnswerStates = reactive<Record<string, MCQAnswer>>({})
const subjectiveAnswerStates = reactive<Record<string, SubjectiveAnswer>>({})

for (const questionMinimals of Object.values(groupedQuestions.value!)) {
  for (const questionMinimal of questionMinimals) {
    const qid = questionMinimal.id
    if (questionMinimal.type === QuestionType.MCQ) {
      mcqAnswerStates[qid] = {
        optionIndex: undefined,
      }
      savedMcqAnswerStates.value[qid] = { ...mcqAnswerStates[qid] }
    } else {
      subjectiveAnswerStates[qid] = {
        text: '',
      }
      savedSubjectiveAnswerStates.value[qid] = {
        ...subjectiveAnswerStates[qid],
      }
    }
  }
}

const { $api } = useNuxtApp()

watchImmediate(currentQuestionId, async qid => {
  if (isNullOrUndefined(qid)) return

  if (answerFetched.value[qid]) return
  answerFetched.value[qid] = true

  const data = await $api<AnswerMinimal>(
    `/api/exams/${examId}/questions/${qid}/answer`
  )

  if (isNullOrUndefined(data.answer)) return
  if (data.question_id !== qid) {
    answerFetched.value[qid] = false
    return
  }

  storeAnswerFromResponse(qid, data)
})

function storeAnswerFromResponse(qid: number, answerResponse: AnswerMinimal) {
  if (isNullOrUndefined(currentCategoryQuestions.value)) {
    console.warn('currentCategoryQuestions is null or undefined')
    return
  }

  const qIdx = currentQuestionIdx.value
  if (qIdx === -1) {
    console.warn('currentQuestionIdx is -1')
    return
  }

  if (currentCategoryQuestions.value[qIdx].type === QuestionType.MCQ) {
    mcqAnswerStates[qid] = {
      optionIndex: (answerResponse.answer as MCQAnswer).optionIndex,
    }
    savedMcqAnswerStates.value[qid] = { ...mcqAnswerStates[qid] }
  } else {
    subjectiveAnswerStates[qid] = {
      text: (answerResponse.answer as SubjectiveAnswer).text,
    }
    savedSubjectiveAnswerStates.value[qid] = { ...subjectiveAnswerStates[qid] }
  }
}

function saveUpdatedAnswers() {
  const promises = []

  for (const [qid, mcqAnswer] of Object.entries(mcqAnswerStates)) {
    const savedState = savedMcqAnswerStates.value[qid]
    if (savedState.optionIndex !== mcqAnswer.optionIndex) {
      promises.push(
        saveMcqAnswer(parseInt(qid), savedState, mcqAnswer).then(res => {
          if (isNullOrUndefined(res)) return
          savedState.optionIndex = (res.answer as MCQAnswer).optionIndex
        })
      )
    }
  }
  for (const [qid, subjectiveAnswer] of Object.entries(
    subjectiveAnswerStates
  )) {
    const savedState = savedSubjectiveAnswerStates.value[qid]
    if (savedState.text !== subjectiveAnswer.text) {
      promises.push(
        saveSubjectiveAnswer(parseInt(qid), savedState, subjectiveAnswer).then(
          res => {
            if (isNullOrUndefined(res)) return
            savedState.text = (res.answer as SubjectiveAnswer).text
          }
        )
      )
    }
  }

  return promises
}
useIntervalFn(saveUpdatedAnswers, AUTO_SAVE_EXAM_ANSWER_INTERVAL_SECONDS * 1000)

/** Save answer for a MCQ question */
function saveMcqAnswer(
  questionId: number,
  savedAnswer: MCQAnswer,
  newAnswer: MCQAnswer
) {
  const upsertAnswerBody = {
    question_id: questionId,
    answer: null as MCQAnswer | null,
  }

  const currentOptionIndex = savedAnswer ? savedAnswer.optionIndex : undefined

  if (savedAnswer && isNullOrUndefined(newAnswer.optionIndex)) {
    return upsertAnswer(examId, upsertAnswerBody)
  }

  if (
    isNullOrUndefined(newAnswer.optionIndex) &&
    isNullOrUndefined(currentOptionIndex)
  ) {
    return Promise.resolve(null)
  }

  if (newAnswer.optionIndex === currentOptionIndex) {
    return Promise.resolve(null)
  }

  upsertAnswerBody.answer = { optionIndex: newAnswer.optionIndex }
  return upsertAnswer(examId, upsertAnswerBody)
}

/** Save answer for a subjective question */
function saveSubjectiveAnswer(
  questionId: number,
  savedAnswer: SubjectiveAnswer,
  newAnswer: SubjectiveAnswer
) {
  const upsertAnswerBody = {
    question_id: questionId,
    answer: null as SubjectiveAnswer | null,
  }

  const currentText = savedAnswer ? savedAnswer.text : ''
  const answerText = newAnswer.text ?? ''

  if (savedAnswer && !answerText) {
    return upsertAnswer(examId, upsertAnswerBody)
  }

  if (!currentText && !answerText) {
    return Promise.resolve(null)
  }

  if (currentText === answerText) {
    return Promise.resolve(null)
  }

  upsertAnswerBody.answer = { text: answerText }
  return upsertAnswer(examId, upsertAnswerBody)
}

// ___________________AUTO-END EXAM ON TIMEOUT____________________
const isExamEnded = ref(false)
const { remaining: redirectCountdown, start: startRedirectCountdown } =
  useCountdown(5, {
    onComplete() {
      navigateTo(`/exams/${examId}/results`)
    },
  })

async function handleExamSubmit() {
  await Promise.all(saveUpdatedAnswers()) // Save any remaining unsaved answers
  await endExam(examId)
  isExamEnded.value = true
  startRedirectCountdown()
}
</script>
