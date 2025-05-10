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
        :save-and-navigate-to="saveAndNavigateTo"
      />
    </UCard>
  </div>

  <div
    v-if="question && currentQuestionId"
    class="col-span-2 flex flex-col gap-y-2.5"
  >
    <UCard>
      <p class="font-medium">{{ question.question.statement }}</p>
    </UCard>

    <UCard v-if="answerPending" :ui="{ body: 'space-y-1.5' }">
      <template v-if="question.type === QuestionType.MCQ">
        <div class="flex gap-x-2.5 rounded-md border border-neutral-300 p-3.5">
          <USkeleton class="size-4" />
          <USkeleton class="h-4.5 w-[250px]" />
        </div>
        <div class="flex gap-x-2.5 rounded-md border border-neutral-300 p-3.5">
          <USkeleton class="size-4" />
          <USkeleton class="h-4.5 w-[250px]" />
        </div>
      </template>

      <template v-else>
        <USkeleton class="h-4 w-full" />
        <USkeleton class="h-4 w-full" />
        <USkeleton class="h-4 w-[250px]" />
      </template>
    </UCard>

    <UCard
      v-else
      :ui="{ root: isNullOrUndefined(answerData!.answer) ? 'grow' : '' }"
    >
      <template
        v-if="
          question.type === QuestionType.MCQ &&
          !isNullOrUndefined(currentQuestionMcqOptions)
        "
      >
        <USkeleton
          v-if="isNullOrUndefined(selectedOptionIndex)"
          class="h-4 w-[250px]"
        />
        <URadioGroup
          v-else
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

      <template v-else-if="question.type !== QuestionType.MCQ">
        <div
          v-if="isNullOrUndefined(answerData?.answer)"
          class="flex flex-col items-center"
        >
          <div class="mt-1 text-gray-600">
            <Icon name="lucide:file-question" size="2.5rem" />
          </div>
          <p class="text-gray-500">This question is unanswered</p>
        </div>

        <p v-else>
          {{ (answerData!.answer as GeneralAnswer).text }}
        </p>
      </template>
    </UCard>

    <UCard v-if="answerPending" :ui="{ root: 'grow', body: 'space-y-1' }">
      <USkeleton class="h-4 w-[44px]" />
      <USkeleton class="h-4 w-[250px]" />
      <USkeleton class="h-8 w-[230px]" />
    </UCard>

    <UCard
      v-else-if="!isNullOrUndefined(answerData!.answer)"
      :ui="{ root: 'grow' }"
    >
      <UFormField
        label="Score"
        description="Score to be awarded for this answer"
        name="score_awarded"
        required
      >
        <UInputNumber
          v-model="evaluationStates[question.id].score_awarded"
          :min="0"
          :max="currentQuestionMaxScore"
          required
        />
      </UFormField>
    </UCard>
  </div>

  <UCard
    v-if="currentCategoryQuestions.length > 0"
    :ui="{ root: 'col-span-2', body: 'flex' }"
  >
    <UButton
      v-if="prevQuestionId"
      label="Previous"
      color="neutral"
      variant="outline"
      :to="{ query: { ...route.query, question: prevQuestionId } }"
      @click="
        saveAndNavigateTo(
          { query: { ...route.query, question: prevQuestionId } },
          { replace: true }
        )
      "
    />
    <UButton
      v-if="nextQuestionId"
      label="Save and next"
      class="ml-auto"
      @click="
        saveAndNavigateTo(
          { query: { ...route.query, question: nextQuestionId } },
          { replace: true }
        )
      "
    />
    <UButton
      v-else-if="!isNullOrUndefined(question)"
      label="Save"
      class="ml-auto"
      @click="saveEvaluation(question.id!)"
    />
  </UCard>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'
import {
  QuestionType,
  type EvaluationAnswer,
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

const [{ data: exam }, { data: groupedQuestions }, { data: sortedCategories }] =
  await Promise.all([
    useExam(examId),
    useExamQuestions(examId),
    useExamCategories(examId),
  ])

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

const { prevQuestionId, nextQuestionId, currentQuestionIdx } =
  useExamQuestionNavigation({
    currentQuestionId,
    currentCategoryQuestions,
  })

// ____________________PREPARE EVALUATION STATES____________________
const evaluationStates = reactive<Record<number, Partial<EvaluationAnswer>>>({})

for (const questionMinimals of Object.values(groupedQuestions.value!)) {
  for (const questionMinimal of questionMinimals) {
    evaluationStates[questionMinimal.id] = {
      score_awarded: undefined,
    }
  }
}

const [
  { data: question },
  { data: answerData, pending: answerPending },
  { data: evaluationAnswer },
] = await Promise.all([
  useExamQuestion(currentQuestionId),
  useExamAnswer(examId, currentQuestionId),
  useEvaluationAnswer(participantId, currentQuestionId),
])

const currentQuestionMaxScore = computed(() => {
  return currentCategoryQuestions.value[currentQuestionIdx.value].max_score
})

// Update evaluationStates when evaluationAnswer is fetched
watchImmediate(evaluationAnswer, newEvaluationAnswer => {
  if (isNullOrUndefined(newEvaluationAnswer)) return

  evaluationStates[newEvaluationAnswer.question_id].score_awarded =
    newEvaluationAnswer.score_awarded
})

/** Save answer for a specific question */
function saveEvaluation(questionId: number) {
  const evaluationState = evaluationStates[questionId]
  const { data: currentEvaluation } = useNuxtData<EvaluationAnswer | null>(
    AsyncDataKeys.EVALUATION_ANSWER(participantId, questionId)
  )

  if (
    currentEvaluation.value &&
    evaluationState.score_awarded === currentEvaluation.value.score_awarded
  ) {
    return
  }

  const updateEvaluationBody = {
    new_score: evaluationState.score_awarded,
    evaluated: true,
  }

  updateAnswerEvaluation(answerData.value!.id, updateEvaluationBody)
}

const saveAndNavigateTo = (async (to, options) => {
  if (question.value) saveEvaluation(question.value.id)
  return navigateTo(to, options)
}) as typeof navigateTo

const selectedOptionIndex = ref<number>()

watchImmediate(answerData, val => {
  if (
    isNullOrUndefined(question.value) ||
    isNullOrUndefined(val?.answer) ||
    question.value.type !== QuestionType.MCQ
  ) {
    selectedOptionIndex.value = undefined
  } else {
    selectedOptionIndex.value = (val.answer as MCQAnswer).optionIndex
  }
})

const currentQuestionMcqOptions = computed(() => {
  if (
    isNullOrUndefined(question.value) ||
    question.value.type !== QuestionType.MCQ
  )
    return null

  return question.value.question.options.map((option, i) => ({
    value: i,
    label: option,
  }))
})
</script>
