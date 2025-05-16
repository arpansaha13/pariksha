<template>
  <div v-if="paper" class="col-span-2 flex items-center gap-2">
    <Icon name="i-heroicons-document-text" size="2rem" />

    <PaperTitle :paper="paper" />

    <div class="ml-auto">
      <PaperDurationModal :paper="paper" />
    </div>
  </div>

  <div class="flex items-center justify-end">
    <UButton
      to="/papers"
      label="Back"
      icon="i-heroicons-arrow-uturn-left"
      size="sm"
      color="neutral"
      variant="ghost"
    />
  </div>

  <div
    v-if="!isNullOrUndefined(sortedCategories)"
    class="col-span-2 flex items-center justify-between gap-x-2 border-b border-gray-200 dark:border-gray-800"
  >
    <PaperCategoryNavigation
      :sorted-categories="sortedCategories"
      :unsaved-count="unsavedCount"
      :get-question-id-for-category-id="getQuestionIdForCategoryId"
    />

    <PaperCategoryManageModal
      v-if="!isNullOrUndefined(groupedQuestions) && currentCategoryId"
      :sorted-categories="sortedCategories"
      :grouped-questions="groupedQuestions"
      :current-category-id="currentCategoryId"
      :get-question-id-for-category-id="getQuestionIdForCategoryId"
    />
  </div>

  <UCard
    :ui="{
      root: 'col-start-3 row-span-2 row-start-2 overflow-hidden flex flex-col',
      body: 'overflow-auto grow p-0 sm:p-0',
    }"
  >
    <template #header>
      <h2 class="text-lg font-semibold">Question Pallet</h2>
    </template>

    <div>
      <PaperQuestionList
        v-if="currentCategoryId && !isNullOrUndefined(currentQuestionId)"
        :current-category-id="currentCategoryId"
        :current-question-id="currentQuestionId"
        :current-category-questions="currentCategoryQuestions"
        :edit-question-form-states="editQuestionFormStates"
        :question-navigation="questionNavigation"
      />
    </div>

    <template #footer>
      <UButton
        :to="{ query: { ...route.query, question: QuestionId.ADD } }"
        icon="i-heroicons-plus"
        label="New question"
        :color="currentQuestionId === QuestionId.ADD ? 'primary' : 'neutral'"
        :variant="currentQuestionId === QuestionId.ADD ? 'subtle' : 'outline'"
        replace
      />
    </template>
  </UCard>

  <UCard
    :ui="{ root: 'col-span-2 overflow-hidden', body: 'h-full overflow-auto' }"
  >
    <PaperQuestionForm
      v-if="
        currentQuestionId === QuestionId.ADD &&
        currentCategoryId &&
        createQuestionFormStates[currentCategoryId]
      "
      ref="createQuestionForm"
      v-model:form-data="createQuestionFormStates[currentCategoryId]"
      @submit="onCreateQuestionSubmit"
    />
    <PaperQuestionForm
      v-if="currentQuestionId && editQuestionFormStates[currentQuestionId]"
      ref="editQuestionForm"
      v-model:form-data="editQuestionFormStates[currentQuestionId]!"
      @submit="onEditQuestionSubmit"
    />
    <PaperQuestionMcq
      v-else-if="question && question.type === QuestionType.MCQ"
      :question="question.question"
    />
    <PaperQuestionNonMcq v-else-if="question" :question="question.question" />
  </UCard>

  <UCard :ui="{ root: 'col-span-2', body: 'flex justify-between' }">
    <div>
      <UButton
        v-if="questionNavigation.prev"
        replace
        label="Previous"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: questionNavigation.prev } }"
      />
    </div>
    <div class="space-x-2">
      <UButton
        v-if="currentQuestionId === QuestionId.ADD"
        label="Add question"
        color="primary"
        variant="solid"
        @click="createQuestionFormRef?.submit()"
      />
      <UButton
        v-else-if="
          currentQuestionId &&
          isNullOrUndefined(editQuestionFormStates[currentQuestionId])
        "
        label="Edit question"
        color="primary"
        variant="subtle"
        @click="startQuestionEdit"
      />
      <template v-else-if="currentQuestionId">
        <UButton
          label="Save"
          color="primary"
          variant="solid"
          @click="editQuestionFormRef?.submit()"
        />
        <UButton
          label="Cancel"
          color="neutral"
          variant="outline"
          @click="cancelQuestionEdit"
        />
      </template>
      <UButton
        v-if="!isNullOrUndefined(questionNavigation.next)"
        replace
        label="Next"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: questionNavigation.next } }"
      />
    </div>
  </UCard>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { ComponentExposed } from 'vue-component-type-helpers'
import { ConfirmModal, PaperQuestionForm } from '#components'
import { QuestionId, QuestionType } from '~/types'

definePageMeta({
  layout: 'paper',
  middleware: ['check-paper-permission'],
})

const route = useRoute()
const paperId = parseInt(route.params.paperId as string)

const overlay = useOverlay()
const confirmModal = overlay.create(ConfirmModal as Component)
provide(InjectionKeys.ConfirmModal, confirmModal)

const [
  { data: paper },
  { data: groupedQuestions },
  { data: sortedCategories },
] = await Promise.all([
  usePaper(paperId),
  usePaperQuestions(paperId),
  usePaperCategories(paperId),
])

const getQuestionIdForCategoryId = usePaperQuestionIdForCategoryId({
  groupedQuestions,
})

// Add initial `category` and `question` queries, if missing
if (
  (!route.query.category || !route.query.question) &&
  sortedCategories.value?.length
) {
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

const { questionNavigation } = usePaperQuestionNavigation({
  currentQuestionId,
  currentCategoryQuestions,
})

const { data: question } = await usePaperQuestion(currentQuestionId)

// ________________CREATE/EDIT QUESTION PREREQUISITES_______________
const defaultCreateQuestionFormState = {
  type: '' as QuestionType,
  question: {
    statement: '',
    options: ['', ''],
  },
  max_score: 0,
  tags: [] as string[],
  correct_answer: '' as string | null | undefined,
}

type QuestionFormState = typeof defaultCreateQuestionFormState

function createQuestionRequestBody(formState: QuestionFormState) {
  const payload = {
    type: formState.type,
    category_id: currentCategoryId.value,
    question: {
      statement: formState.question.statement,
    },
    max_score: formState.max_score,
    tags: [],
    correct_answer: formState.correct_answer,
  } as Parameters<typeof createQuestion>[1]

  if (payload.type === QuestionType.MCQ) {
    payload.question.options = formState.question.options
  }

  return payload
}

// ___________UNSAVED COUNT FOR CHIPS ON CATEGORY LINKS___________
const unsavedCount = ref<Record<number, number>>({})

function incUnsavedCount(categoryId: number) {
  if (!unsavedCount.value[categoryId]) {
    unsavedCount.value[categoryId] = 1
  } else {
    unsavedCount.value[categoryId]++
  }
}

function decUnsavedCount(categoryId: number) {
  if (!unsavedCount.value[categoryId]) {
    unsavedCount.value[categoryId] = 0
  } else {
    unsavedCount.value[categoryId]--
  }
}

// ________________________CREATE QUESTION________________________
const createQuestionFormRef =
  useTemplateRef<ComponentExposed<typeof PaperQuestionForm>>(
    'createQuestionForm'
  )
const createQuestionFormStates = reactive<Record<number, QuestionFormState>>({})

watchImmediate(currentCategoryId, categoryId => {
  if (categoryId && isNullOrUndefined(createQuestionFormStates[categoryId])) {
    createQuestionFormStates[categoryId] = structuredClone(
      defaultCreateQuestionFormState
    )
  }
})

async function onCreateQuestionSubmit() {
  if (isNullOrUndefined(currentCategoryId.value)) return

  const formState = createQuestionFormStates[currentCategoryId.value]
  if (isNullOrUndefined(formState)) return

  const payload = createQuestionRequestBody(formState)

  try {
    await createQuestion(paperId, payload)

    // Navigate to the newly created question
    // currentCategoryQuestions will be refreshed by createQuestion
    const latestQuestion = currentCategoryQuestions.value.at(-1)
    if (latestQuestion) {
      navigateTo({
        query: { ...route.query, question: latestQuestion.id },
      })
    }

    // Reset formState after submission
    createQuestionFormStates[currentCategoryId.value] = structuredClone(
      defaultCreateQuestionFormState
    )
  } catch (error) {
    console.error('Failed to create question:', error)
  }
}

// ________________________EDIT QUESTION__________________________
const editQuestionFormRef =
  useTemplateRef<ComponentExposed<typeof PaperQuestionForm>>('editQuestionForm')
const editQuestionFormStates = reactive<
  Record<number, QuestionFormState | null>
>({})

function startQuestionEdit() {
  if (!question.value || !currentQuestionId.value) return
  if (currentQuestionId.value === QuestionId.ADD) return

  // Create form state for editing if it doesn't exist
  if (isNullOrUndefined(editQuestionFormStates[currentQuestionId.value])) {
    editQuestionFormStates[currentQuestionId.value] = {
      type: question.value.type,
      question: {
        statement: '',
        options: ['', ''],
      },
      max_score: question.value.max_score,
      tags: [...question.value.tags],
      correct_answer: question.value.correct_answer ?? undefined,
    }

    // Parse and populate question data based on type
    if (question.value.type === QuestionType.MCQ) {
      const mcqQuestion = question.value.question
      editQuestionFormStates[currentQuestionId.value]!.question = {
        statement: mcqQuestion.statement,
        options: [...mcqQuestion.options], // Store new array reference
      }
    } else {
      const subjectiveQuestion = question.value.question
      editQuestionFormStates[currentQuestionId.value]!.question.statement =
        subjectiveQuestion.statement
    }

    incUnsavedCount(question.value.category_id)
  }
}

function cancelQuestionEdit() {
  if (!currentQuestionId.value) return
  editQuestionFormStates[currentQuestionId.value] = null
  decUnsavedCount(currentCategoryId.value!)
}

async function onEditQuestionSubmit() {
  if (!currentQuestionId.value) return
  if (currentQuestionId.value === QuestionId.ADD) return

  const formState = editQuestionFormStates[currentQuestionId.value]!
  const payload = createQuestionRequestBody(formState)

  try {
    await updateQuestion(currentQuestionId.value, paperId, payload)
    // Clear edit form state after successful update
    editQuestionFormStates[currentQuestionId.value] = null
    decUnsavedCount(currentCategoryId.value!)
  } catch (error) {
    console.error('Failed to update question:', error)
  }
}

usePaperAutoCancelQuestionEdit({
  question,
  editQuestionFormStates,
  decUnsavedCount,
})
</script>

<style scoped>
@reference "../../../assets/css/main.css";

.draggable-ghost {
  @apply bg-gray-200;
}
.draggable-hold {
  @apply cursor-grabbing opacity-0;
}
</style>
