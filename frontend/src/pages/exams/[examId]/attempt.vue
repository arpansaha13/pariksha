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

    <UCard v-if="question" :ui="{ root: 'col-span-2' }">
      <ExamQuestionMcq
        v-if="question.type === QuestionType.MCQ"
        :question="question.question"
      />
      <ExamQuestionNonMcq v-else :question="question.question" />
    </UCard>

    <UCard :ui="{ root: 'col-span-2', body: 'flex' }">
      <UButton
        v-if="prevQuestionId"
        replace
        label="Previous"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: prevQuestionId } }"
      />
      <UButton
        v-if="nextQuestionId"
        replace
        label="Next"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: nextQuestionId } }"
        class="ml-auto"
      />
    </UCard>
  </template>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import { ConfirmModal } from '#components'
import { QuestionType } from '~/types'

definePageMeta({
  layout: 'paper',
  middleware: ['check-exam-participant'],
})

const route = useRoute()
const examId = parseInt(route.params.examId as string)
const { data: exam } = await useExam(examId)
const { data: participant } = await useExamParticipant(examId)
const { data: groupedQuestions } = await useExamQuestions(examId)
const { data: sortedCategories } = await useExamCategories(examId)

const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal)
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
  return route.query.category ? parseInt(route.query.category as string) : null
})

const currentCategoryQuestions = computed(() => {
  if (!groupedQuestions.value || !currentCategoryId.value) return []
  return groupedQuestions.value[currentCategoryId.value] ?? []
})

const currentQuestionId = computed(() => {
  return route.query.question ? parseInt(route.query.question as string) : null
})

const { prevQuestionId, nextQuestionId } = useExamQuestionNavigation({
  currentQuestionId,
  currentCategoryQuestions,
})

const { data: question } = await useExamQuestion(currentQuestionId)

// ___________________AUTO-END EXAM ON TIMEOUT____________________
const isExamEnded = ref(false)
const { remaining: redirectCountdown, start: startRedirectCountdown } =
  useCountdown(5, {
    onComplete() {
      navigateTo(`/exams/${examId}/results`)
    },
  })

async function handleExamSubmit() {
  isExamEnded.value = true
  startRedirectCountdown()
  endExam(examId)
}
</script>
