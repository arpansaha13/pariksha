<template>
  <div v-if="exam" class="col-span-2 flex items-center gap-2">
    <Icon name="i-heroicons-document-text" size="2rem" />
    <h1 class="text-xl font-semibold">{{ exam.title }}</h1>
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

  <div v-if="question" class="col-span-2">
    <!-- prettier-ignore -->
    <EvaluationQuestionMcq
      v-if="question.type === QuestionType.MCQ"
      :answer="(answer?.answer as MCQAnswer)"
      :question="question.question"
    />
    <!-- prettier-ignore -->
    <EvaluationQuestionNonMcq
      v-else
      :answer="(answer?.answer as GeneralAnswer)"
      :question="question.question"
    />
  </div>

  <UCard
    v-if="currentCategoryQuestions.length > 0"
    :ui="{ root: 'col-span-2', body: 'flex' }"
  >
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
      label="Next"
      class="ml-auto"
      :to="{ query: { ...route.query, question: nextQuestionId } }"
    />
  </UCard>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'
import {
  QuestionType,
  type ExamPermission,
  type GeneralAnswer,
  type MCQAnswer,
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
  // { data: participant },
  { data: groupedQuestions },
  { data: sortedCategories },
] = await Promise.all([
  useExam(examId),
  // useExamParticipant(examId),
  useExamQuestions(examId),
  useExamCategories(examId),
])

// const overlay = useOverlay()
// const confirmModal = overlay.create(ConfirmModal as Component)
// provide(InjectionKeys.ConfirmModal, confirmModal)

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

const [{ data: question }, { data: answer }] = await Promise.all([
  useExamQuestion(currentQuestionId),
  useEvaluationAnswer(participantId, currentQuestionId),
])
</script>
