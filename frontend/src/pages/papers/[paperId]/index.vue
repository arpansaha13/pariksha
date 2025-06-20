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
    <PaperCategoryNavigation :sorted-categories="sortedCategories" />

    <PaperCategoryManageModal
      v-if="!isNullOrUndefined(groupedQuestions) && currentCategoryId"
      :sorted-categories="sortedCategories"
      :grouped-questions="groupedQuestions"
      :current-category-id="currentCategoryId"
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
        :show-question-edit-chip="showQuestionEditChip"
        :question-navigation="questionNavigation"
      />
    </div>

    <template #footer>
      <UButton
        :to="{ query: { ...route.query, question: QUESTION_ID_ADD } }"
        icon="i-heroicons-plus"
        label="New question"
        :color="currentQuestionId === QUESTION_ID_ADD ? 'primary' : 'neutral'"
        :variant="currentQuestionId === QUESTION_ID_ADD ? 'subtle' : 'outline'"
        replace
      />
    </template>
  </UCard>

  <UCard
    :ui="{ root: 'col-span-2 overflow-hidden', body: 'h-full overflow-auto' }"
  >
    <PaperQuestionForm
      v-if="
        currentQuestionId === QUESTION_ID_ADD &&
        currentCategoryId &&
        createQuestionFormStates[currentCategoryId]
      "
      ref="createQuestionForm"
      v-model:form-data="createQuestionFormStates[currentCategoryId]"
      :has-test-cases="false"
      @submit="onCreateQuestionSubmit"
    />
    <PaperQuestionForm
      v-else-if="currentQuestionId && editQuestionFormStates[currentQuestionId]"
      ref="editQuestionForm"
      v-model:form-data="editQuestionFormStates[currentQuestionId]!"
      :has-test-cases="
        !isNullOrUndefined(question) &&
        question.type === QuestionType.CODING &&
        !isNullOrUndefined(question.test_cases) &&
        question.test_cases.length > 0
      "
      @submit="onEditQuestionSubmit"
    />
    <PaperQuestionCodingTestCasesForm
      v-else-if="
        question &&
        currentQuestionId &&
        question.type === QuestionType.CODING &&
        codingQuestionTestCaseFormStates[currentQuestionId]
      "
      ref="codingQuestionTestCaseForm"
      v-model:test-cases="codingQuestionTestCaseFormStates[currentQuestionId]!"
      :coding-question-content="question.question"
      @submit="onDefineTestCasesSubmit"
    />
    <template v-else-if="question">
      <PaperQuestionMcq
        v-if="question.type === QuestionType.MCQ"
        :question="question.question"
      />
      <div v-else-if="question.type === QuestionType.SUBJECTIVE">
        <p>{{ question.question.statement }}</p>
      </div>
      <DisplayCodingQuestion
        v-else-if="question.type === QuestionType.CODING"
        :content="question.question"
        :test-cases="question.test_cases ?? []"
        :editor-link="`/editor/questions/${question.id}`"
      />
    </template>
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
        v-if="currentQuestionId === QUESTION_ID_ADD"
        label="Add question"
        color="primary"
        variant="solid"
        @click="createQuestionFormRef?.submit()"
      />
      <template
        v-else-if="
          currentQuestionId &&
          isNullOrUndefined(
            codingQuestionTestCaseFormStates[currentQuestionId]
          ) &&
          isNullOrUndefined(editQuestionFormStates[currentQuestionId])
        "
      >
        <UButton
          v-if="question?.type === QuestionType.CODING"
          label="Define test cases"
          color="primary"
          variant="subtle"
          @click="startDefineTestCases"
        />
        <UButton
          label="Edit question"
          color="primary"
          variant="subtle"
          @click="startQuestionEdit"
        />
      </template>
      <template v-else>
        <UButton
          label="Save"
          color="primary"
          variant="solid"
          @click="saveQuestionEditMode"
        />
        <UButton
          label="Cancel"
          color="neutral"
          variant="outline"
          @click="cancelQuestionEditMode"
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
import { defu } from 'defu'
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { ComponentExposed } from 'vue-component-type-helpers'
import {
  ConfirmModal,
  PaperQuestionCodingTestCasesForm,
  PaperQuestionForm,
} from '#components'

definePageMeta({
  layout: 'paper',
  middleware: ['check-paper-permission'],
})

const route = useRoute()
const paperId = route.params.paperId as PaperId

const paperStore = usePaperStore()
onUnmounted(() => {
  paperStore.$reset()
})

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

// Add initial `category` and `question` queries, if missing
if (
  (!route.query.category || !route.query.question) &&
  sortedCategories.value?.length
) {
  const categoryId = sortedCategories.value[0].id
  const questionId = paperStore.getQuestionIdForCategoryId(categoryId)
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

const { questionNavigation } = usePaperQuestionNavigation({
  currentQuestionId,
  currentCategoryQuestions,
})

const { data: question } = await usePaperQuestion(currentQuestionId)

// ________________PAPER QUESTION-ID FOR CATEGORY-ID________________
watchImmediate(route, newRoute => {
  const query = newRoute.query
  if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
  const categoryId = parseInt(query.category as string) as CategoryId
  paperStore.lastVisitedQuestionForCategory[categoryId] =
    query.question as QuestionId
})

// ________________CREATE/EDIT QUESTION PREREQUISITES_______________
const defaultCreateQuestionFormState: MergedQuestion = {
  type: 0 as QuestionType,
  question: {
    title: '',
    statement: '',
    options: ['', ''],
    input_definitions: [
      {
        variable_name: getDefaultCodingQuestionInputVariableName(1),
        type: 0 as QuestionCodingContentInputTypes,
      },
    ],
    output_definition: {
      type: 0 as QuestionCodingContentInputTypes,
    },
  },
  max_score: 0,
  tags: [],
  correct_answer: undefined,
}

// ________________________CREATE QUESTION________________________
const createQuestionFormRef =
  useTemplateRef<ComponentExposed<typeof PaperQuestionForm>>(
    'createQuestionForm'
  )
const createQuestionFormStates = reactive<Record<number, MergedQuestion>>({})

watchImmediate(currentCategoryId, categoryId => {
  if (categoryId && isNullOrUndefined(createQuestionFormStates[categoryId])) {
    createQuestionFormStates[categoryId] = structuredClone(
      defaultCreateQuestionFormState
    )
  }
})

async function onCreateQuestionSubmit() {
  const catId = currentCategoryId.value
  if (isNullOrUndefined(catId)) return

  const formState = createQuestionFormStates[catId]
  if (isNullOrUndefined(formState)) return

  try {
    const newQuestionId = await createQuestion(paperId, catId, formState)

    // Navigate to the newly created question
    if (!isNullOrUndefined(newQuestionId)) {
      navigateTo({
        query: { ...route.query, question: newQuestionId },
        replace: true,
      })
    }

    // Reset formState after submission
    createQuestionFormStates[catId] = structuredClone(
      defaultCreateQuestionFormState
    )
  } catch (error) {
    console.error('Failed to create question:', error)
  }
}

// _________________QUESTION EDIT CHIP-INDICATOR__________________
const showQuestionEditChip = ref<Record<QuestionId, boolean>>({})

// ________________________EDIT QUESTION__________________________
const editQuestionFormRef =
  useTemplateRef<ComponentExposed<typeof PaperQuestionForm>>('editQuestionForm')
const editQuestionFormStates = reactive<
  Record<QuestionId, MergedQuestion | null>
>({})

function startQuestionEdit() {
  const qid = currentQuestionId.value
  if (!question.value || !qid) return
  if (qid === QUESTION_ID_ADD) return

  // Create form state for editing if it doesn't exist
  if (isNullOrUndefined(editQuestionFormStates[qid])) {
    editQuestionFormStates[qid] = defu(
      {
        type: question.value.type,
        max_score: question.value.max_score,
        tags: [...question.value.tags],
        question: {
          statement: question.value.question.statement,
        },
      },
      defaultCreateQuestionFormState
    )

    // defu adds a null type to correct_answer causing ts-error
    // add correct_answer separately
    editQuestionFormStates[qid].correct_answer = question.value.correct_answer

    // Parse and populate question data based on type
    const formState = editQuestionFormStates[qid]!.question
    const qType = question.value.type

    if (qType === QuestionType.MCQ) {
      const mcqQuestion = question.value.question
      formState.options = [...mcqQuestion.options] // Store new array reference
    } else if (qType === QuestionType.CODING) {
      const codingQuestion = question.value.question
      formState.title = codingQuestion.title
      formState.input_definitions = structuredClone(
        toRaw(codingQuestion.input_definitions)
      )
      formState.output_definition = structuredClone(
        toRaw(codingQuestion.output_definition)
      )
    }

    paperStore.incUnsavedCount(question.value.category_id)
  }

  showQuestionEditChip.value[qid] = true
}

async function onEditQuestionSubmit() {
  const qid = currentQuestionId.value
  if (!qid || qid === QUESTION_ID_ADD) return

  const formState = editQuestionFormStates[qid]!

  try {
    await updateQuestion(qid, paperId, formState)
    // Clear edit form state after successful update
    editQuestionFormStates[qid] = null
    showQuestionEditChip.value[qid] = false
    paperStore.decUnsavedCount(currentCategoryId.value!)
  } catch (error) {
    console.error('Failed to update question:', error)
  }
}

// __________________CODING QUESTION TEST CASES___________________
const codingQuestionTestCaseFormRef = useTemplateRef<
  ComponentExposed<typeof PaperQuestionCodingTestCasesForm>
>('codingQuestionTestCaseForm')
const codingQuestionTestCaseFormStates = reactive<
  Record<QuestionId, QuestionCodingTestCase[] | null>
>({})

function startDefineTestCases() {
  const qid = currentQuestionId.value
  if (!question.value || !qid) return
  if (qid === QUESTION_ID_ADD) return

  const qType = question.value.type

  if (qType !== QuestionType.CODING) {
    logWarning('startDefineTestCases called without QuestionType.CODING')
    return
  }

  // Create form state for test-cases if it doesn't exist
  if (isNullOrUndefined(codingQuestionTestCaseFormStates[qid])) {
    // Parse and populate question data based on type

    const testCases = question.value.test_cases
    if (testCases && testCases.length > 0) {
      codingQuestionTestCaseFormStates[qid] = structuredClone(toRaw(testCases))
    } else {
      codingQuestionTestCaseFormStates[qid] = [
        {
          inputs: Array.from({
            length: question.value.question.input_definitions.length,
          }),
          output: '',
          explanation: '',
          hidden: false,
        },
      ]
    }

    paperStore.incUnsavedCount(question.value.category_id)
  }

  showQuestionEditChip.value[qid] = true
}

async function onDefineTestCasesSubmit() {
  const qid = currentQuestionId.value
  if (
    !qid ||
    qid === QUESTION_ID_ADD ||
    !question.value ||
    question.value.type !== QuestionType.CODING
  ) {
    return
  }

  const testCases = codingQuestionTestCaseFormStates[qid]!

  try {
    await upsertPaperTestCases(qid, testCases)
    // Clear form state after successful update
    codingQuestionTestCaseFormStates[qid] = null
    showQuestionEditChip.value[qid] = false
    paperStore.decUnsavedCount(currentCategoryId.value!)
  } catch (error) {
    console.error('Failed to update question:', error)
  }
}

// ___________________COMBINED SAVE AND CANCEL____________________
/** Combined "Save" for question-edit and define-test-cases */
function saveQuestionEditMode() {
  if (!currentQuestionId.value) return

  if (!isNullOrUndefined(editQuestionFormStates[currentQuestionId.value])) {
    editQuestionFormRef.value?.submit()
  } else {
    codingQuestionTestCaseFormRef.value?.submit()
  }
}

/** Combined "Cancel" for question-edit and define-test-cases */
function cancelQuestionEditMode() {
  const qid = currentQuestionId.value
  if (!qid) return
  editQuestionFormStates[qid] = null
  codingQuestionTestCaseFormStates[qid] = null
  showQuestionEditChip.value[qid] = false
  paperStore.decUnsavedCount(currentCategoryId.value!)
}
</script>

<style scoped>
@reference "~/assets/css/main.css";

.draggable-ghost {
  @apply bg-gray-200;
}
.draggable-hold {
  @apply cursor-grabbing opacity-0;
}
</style>
