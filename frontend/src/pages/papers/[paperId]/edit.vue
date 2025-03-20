<template>
  <div v-if="paper" class="col-span-2 flex items-center gap-2">
    <Icon name="i-heroicons-document-text" size="2rem" />

    <EditableRoot
      v-model="editablePaperTitle"
      name="paper-title"
      activation-mode="focus"
      submit-mode="both"
      placeholder=""
      @submit="updatePaperTitle"
    >
      <EditableArea
        class="rounded-sm px-1 text-xl focus-within:outline hover:outline"
      >
        <EditablePreview as="h1" class="font-semibold" />
        <EditableInput class="font-semibold outline-none" />
      </EditableArea>
    </EditableRoot>

    <UButton
      to="/papers"
      label="Back"
      icon="i-heroicons-arrow-uturn-left"
      size="sm"
      color="neutral"
      variant="ghost"
      class="ml-auto"
    />
  </div>

  <div
    v-if="!isNullOrUndefined(categoryLinks)"
    class="col-span-2 flex items-center justify-between gap-x-2 border-b border-gray-200 dark:border-gray-800"
  >
    <!-- subtract button-width and gap -->
    <ScrollAreaRoot class="max-w-[calc(100%-44px)]">
      <ScrollAreaViewport>
        <UNavigationMenu
          :items="categoryLinks"
          color="primary"
          orientation="horizontal"
          variant="link"
          highlight
        >
          <template #item="{ item }">
            <UChip
              :show="!!unsavedCount[item.to.query.category]"
              :ui="{
                root: 'overflow-hidden shrink',
                base: '-top-1 -right-1.5',
              }"
            >
              <span class="truncate">{{ item.label }}</span>
            </UChip>
          </template>
        </UNavigationMenu>
      </ScrollAreaViewport>
      <ScrollAreaScrollbar
        class="flex touch-none bg-white p-0.5 transition-colors ease-out select-none data-[orientation=horizontal]:h-2 data-[orientation=horizontal]:flex-col"
        orientation="horizontal"
      >
        <ScrollAreaThumb
          class="relative flex-1 rounded-sm bg-gray-200 transition-colors before:absolute before:top-1/2 before:left-1/2 before:h-full before:w-full before:-translate-x-1/2 before:-translate-y-1/2 before:content-[''] hover:bg-gray-300"
        />
      </ScrollAreaScrollbar>
    </ScrollAreaRoot>

    <UModal
      title="Manage categories"
      description="Add, edit, remove, or reorder your categories"
      :ui="{
        body: 'space-y-2',
      }"
      @after:leave="handleReorder"
    >
      <UTooltip text="Manage categories">
        <UButton
          icon="i-heroicons-adjustments-vertical"
          size="sm"
          color="neutral"
          square
          variant="outline"
        />
      </UTooltip>

      <template #body>
        <ul>
          <Draggable
            v-model="editableCategoriesCopy"
            item-key="id"
            group="category-editable"
            handle=".draggable-handle"
            ghost-class="draggable-ghost"
            drag-class="draggable-hold"
            :animation="250"
            @start="dragging = true"
            @end="dragging = false"
          >
            <template #item="{ element: category }">
              <li class="group flex items-center gap-2 rounded-sm">
                <div class="draggable-handle flex size-5 cursor-grab">
                  <Icon
                    name="i-heroicons-bars-2"
                    class="m-auto text-gray-400 transition-colors"
                  />
                </div>

                <EditableRoot
                  v-slot="{ isEditing }"
                  v-model="categoryNames[category.id]"
                  activation-mode="focus"
                  submit-mode="both"
                  class="grow"
                  placeholder=""
                  @submit="() => handleUpdateCategory(category)"
                >
                  <EditableArea
                    :class="[
                      'px-2 py-2 text-sm transition-colors dark:border-gray-800',
                      isEditing
                        ? 'border-primary-500 border-b-2'
                        : 'border-b border-gray-200',
                    ]"
                  >
                    <EditablePreview class="" />
                    <EditableInput class="outline-none" />
                  </EditableArea>
                </EditableRoot>

                <div v-show="!dragging">
                  <UButton
                    icon="i-heroicons-trash"
                    size="sm"
                    color="error"
                    square
                    variant="soft"
                    class="invisible group-hover:visible"
                    loading-auto
                    :disabled="sortedCategories?.length === 1"
                    @click="handleDeleteCategory(category)"
                  />
                </div>
              </li>
            </template>
          </Draggable>
        </ul>
      </template>

      <template #footer>
        <UButton
          label="Add category"
          color="primary"
          variant="soft"
          loading-auto
          @click="createCategory(paperId)"
        />
      </template>
    </UModal>
  </div>

  <div class="col-start-3 row-span-2 row-start-2">
    <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

    <UCard v-if="currentCategoryQuestions">
      <ul class="flex flex-wrap gap-4">
        <li v-for="(q, i) of currentCategoryQuestions" :key="q.id">
          <UChip :show="!isNullOrUndefined(editQuestionFormStates[q.id])" inset>
            <UButton
              :to="{ query: { ...route.query, question: q.id } }"
              :color="currentQuestionId === q.id ? 'primary' : 'neutral'"
              :variant="currentQuestionId === q.id ? 'subtle' : 'outline'"
              size="lg"
              class="flex size-10 items-center justify-center rounded-full"
            >
              {{ i + 1 }}
            </UButton>
          </UChip>
        </li>
        <li>
          <UTooltip text="Add question">
            <UButton
              :to="{ query: { ...route.query, question: QuestionId.ADD } }"
              icon="i-heroicons-plus"
              :color="
                currentQuestionId === QuestionId.ADD ? 'primary' : 'neutral'
              "
              :variant="
                currentQuestionId === QuestionId.ADD ? 'subtle' : 'outline'
              "
              size="lg"
              class="flex size-10 items-center justify-center rounded-full"
            />
          </UTooltip>
        </li>
      </ul>
    </UCard>
  </div>

  <UCard
    :ui="{ root: 'col-span-2 overflow-hidden', body: 'h-full overflow-auto' }"
  >
    <QuestionCreationForm
      v-if="
        currentQuestionId === QuestionId.ADD &&
        currentCategoryId &&
        createQuestionFormStates[currentCategoryId]
      "
      ref="createQuestionForm"
      v-model:form-data="createQuestionFormStates[currentCategoryId]"
      @submit="onCreateQuestionSubmit"
    />
    <QuestionCreationForm
      v-if="currentQuestionId && editQuestionFormStates[currentQuestionId]"
      ref="editQuestionForm"
      v-model:form-data="editQuestionFormStates[currentQuestionId]!"
      @submit="onEditQuestionSubmit"
    />
    <QuestionMcq
      v-else-if="question && question.type === QuestionType.MCQ"
      :question="question.question"
    />
    <QuestionNonMcq v-else-if="question" :question="question.question" />
  </UCard>

  <UCard :ui="{ root: 'col-span-2', body: 'flex justify-between' }">
    <div>
      <UButton
        v-if="prevQuestionId"
        label="Previous"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: prevQuestionId } }"
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
        v-if="!isNullOrUndefined(nextQuestionId)"
        label="Next"
        color="neutral"
        variant="outline"
        :to="{ query: { ...route.query, question: nextQuestionId } }"
      />
    </div>
  </UCard>
</template>

<script setup lang="ts">
import Draggable from 'vuedraggable'
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { ComponentExposed } from 'vue-component-type-helpers'
import { ConfirmModal, QuestionCreationForm } from '#components'
import {
  type Question,
  QuestionId,
  QuestionType,
  type QuestionCategory,
} from '~/types'

definePageMeta({
  layout: 'paper',
})

enum QuestionIndex {
  NON_EXISTENT = -1,

  /** Special case for add question */
  ADD = -2,
}

const route = useRoute()
const overlay = useOverlay()
const paperId = parseInt(route.params.paperId as string)
const { data: paper } = await usePaper(paperId)
const { data: groupedQuestions } = await usePaperQuestions(paperId)
const { data: sortedCategories } = await usePaperCategories(paperId)

const dragging = ref(false)
const editableCategoriesCopy = ref<QuestionCategory[]>([])
const lastVisitedQuestionForCategory = ref<Record<number, string>>({})

const confirmModal = overlay.create(ConfirmModal)

watch(
  sortedCategories,
  val => {
    if (val) editableCategoriesCopy.value = [...val]
  },
  { deep: true, immediate: true }
)

watchImmediate(route, newRoute => {
  const query = newRoute.query
  if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
  const categoryId = parseInt(query.category as string)
  lastVisitedQuestionForCategory.value[categoryId] = query.question as string
})

function getQuestionIdForCategoryId(categoryId: number) {
  const categoryQuestions = groupedQuestions.value?.[categoryId]
  if (isNullOrUndefined(categoryQuestions)) return QuestionId.ADD
  const questionId =
    lastVisitedQuestionForCategory.value[categoryId] ?? categoryQuestions[0].id
  return questionId
}

// Add initial `category` and `question` queries, if missing
if (
  (!route.query.category || !route.query.question) &&
  sortedCategories.value?.length
) {
  const categoryId = sortedCategories.value[0].id
  const questionId = getQuestionIdForCategoryId(categoryId)
  const query = { category: categoryId, question: questionId }
  navigateTo({ query }, { replace: true })
}

const categoryLinks = computed(() => {
  if (!sortedCategories.value) return null

  return sortedCategories.value.map(category => ({
    label: category.name,
    to: {
      query: {
        category: category.id,
        question: getQuestionIdForCategoryId(category.id),
      },
    },
    exactQuery: true,
  }))
})

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

const currentQuestionIdx = computed(() => {
  if (!currentQuestionId.value) return QuestionIndex.NON_EXISTENT
  if (currentQuestionId.value === QuestionId.ADD) return QuestionIndex.ADD
  return (
    currentCategoryQuestions.value?.findIndex(
      q => q.id === currentQuestionId.value
    ) ?? QuestionIndex.NON_EXISTENT
  )
})

const { data: question } = await useQuestion(currentQuestionId)

const prevQuestionId = computed(() => {
  if (!currentCategoryQuestions.value) return null

  // If on add question page, show last question as prev
  if (currentQuestionId.value === QuestionId.ADD) {
    return currentCategoryQuestions.value.at(-1)?.id
  }

  if (currentQuestionIdx.value <= 0) return null
  return currentCategoryQuestions.value[currentQuestionIdx.value - 1].id
})

const nextQuestionId = computed(() => {
  if (!currentCategoryQuestions.value) return null

  // If on add question page, there is no next
  if (currentQuestionId.value === QuestionId.ADD) return null

  // If on last question, show add question as next
  if (currentQuestionIdx.value === currentCategoryQuestions.value.length - 1) {
    return QuestionId.ADD
  }

  if (currentQuestionIdx.value === QuestionIndex.NON_EXISTENT) return null
  return currentCategoryQuestions.value[currentQuestionIdx.value + 1].id
})

// ________________________EDIT PAPER TITLE_________________________
const editablePaperTitle = ref(paper.value!.title)
function updatePaperTitle() {
  editablePaperTitle.value.trim()
  if (!editablePaperTitle.value) {
    editablePaperTitle.value = 'Untitled Paper'
  }
  if (editablePaperTitle.value !== paper.value!.title) {
    return updatePaper(paperId, { title: editablePaperTitle.value })
  }
}
watch(paper, newPaper => {
  editablePaperTitle.value = newPaper!.title
})

// ________________________DELETE CATEGORY_________________________
async function handleDeleteCategory(category: QuestionCategory) {
  let shouldDelete = true

  // If category has questions, then show a confirmation modal
  const categoryQuestions = groupedQuestions.value![category.id]
  if (categoryQuestions && categoryQuestions.length > 0) {
    shouldDelete = await confirmModal.open({
      title: 'Confirm category deletion',
      description: `This category "${category.name}" has ${categoryQuestions.length} questions which will be deleted along with it.`,
    })
  }

  if (shouldDelete) {
    return doDeleteCategory(category.id)
  }
}
async function doDeleteCategory(categoryId: number) {
  // If deleting current category, switch to another category first
  if (currentCategoryId.value === categoryId) {
    // Find another category to switch to
    const nextCategory = sortedCategories.value?.find(c => c.id !== categoryId)
    if (nextCategory) {
      await navigateTo({
        query: {
          category: nextCategory.id,
          question: getQuestionIdForCategoryId(nextCategory.id),
        },
      })
    }
  }

  await deleteCategory(categoryId, paperId)
}

// ________________________UPDATE CATEGORY_________________________
const categoryNames = ref<Record<number, string>>({})
async function handleUpdateCategory(category: QuestionCategory) {
  categoryNames.value[category.id].trim()

  // If empty name, use default "Category {order}"
  const name = categoryNames.value[category.id] || `Category ${category.order}`

  if (name !== category.name) {
    await updateCategory(category.id, paperId, { name })
  }
}
watchImmediate(sortedCategories, newCategories => {
  if (!newCategories) return

  newCategories.forEach(category => {
    categoryNames.value[category.id] = category.name
  })
})

// _______________________REORDER CATEGORY________________________
/**
 * `reorder-categories` api is fired after the modal closes, so that all reorders can be batched.
 * Because individual reorders cause inconsistencies during rollback (in case of error).
 */
function handleReorder() {
  let isReordered = false

  // Check if the order was changed
  for (let i = 0; i < sortedCategories.value!.length; i++) {
    if (sortedCategories.value![i].id !== editableCategoriesCopy.value[i].id) {
      isReordered = true
      break
    }
  }

  if (isReordered) {
    reorderCategories(
      paperId,
      editableCategoriesCopy.value.map(cat => cat.id)
    )
  }
}

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
  useTemplateRef<ComponentExposed<typeof QuestionCreationForm>>(
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
  useTemplateRef<ComponentExposed<typeof QuestionCreationForm>>(
    'editQuestionForm'
  )
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
      const generalQuestion = question.value.question
      editQuestionFormStates[currentQuestionId.value]!.question.statement =
        generalQuestion.statement
    }

    incUnsavedCount(question.value.category.id)
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

// ___________________AUTO-CANCEL QUESTION EDIT MODE_______________
function isEditFormStateDirty(
  oldQuestion: Question,
  formState: QuestionFormState
): boolean {
  if (!oldQuestion || !formState) return false

  if (
    formState.type !== oldQuestion.type ||
    formState.max_score !== oldQuestion.max_score ||
    formState.correct_answer !== (oldQuestion.correct_answer ?? undefined) ||
    !arrayEquals(formState.tags, oldQuestion.tags ?? [])
  ) {
    return true
  }

  // Check question data based on type
  if (oldQuestion.type === QuestionType.MCQ) {
    const mcqQuestion = oldQuestion.question
    return (
      formState.question.statement !== mcqQuestion.statement ||
      !arrayEquals(formState.question.options, mcqQuestion.options)
    )
  }

  const generalQuestion = oldQuestion.question
  return formState.question.statement !== generalQuestion.statement
}

watch(question, (_, oldQuestion) => {
  if (!oldQuestion) return

  const formState = editQuestionFormStates[oldQuestion.id]

  // If previous question was in edit mode but not dirty, cancel its edit
  if (formState && !isEditFormStateDirty(oldQuestion, formState)) {
    editQuestionFormStates[oldQuestion.id] = null
    decUnsavedCount(oldQuestion.category.id)
  }
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
