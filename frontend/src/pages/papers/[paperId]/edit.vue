<template>
  <UContainer
    as="main"
    class="grid h-full grow grid-cols-3 grid-rows-[auto_1fr] gap-6 py-4"
  >
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

    <div />

    <div class="col-span-2 flex h-full flex-col gap-y-4">
      <div
        v-if="categoryLinks !== null"
        class="flex items-center justify-between gap-x-2 border-b border-gray-200 dark:border-gray-800"
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
            />
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

      <UCard v-if="question" class="grow">
        <QuestionMcq
          v-if="question.type === QuestionType.MCQ"
          :question="question.question"
        />
        <QuestionNonMcq v-else :question="question.question" />
      </UCard>

      <UCard :ui="{ body: 'flex' }">
        <UButton
          v-if="prevQuestionId"
          label="Previous"
          color="neutral"
          variant="outline"
          :to="{ query: { ...route.query, question: prevQuestionId } }"
        />
        <UButton
          v-if="nextQuestionId"
          label="Next"
          color="neutral"
          variant="outline"
          :to="{ query: { ...route.query, question: nextQuestionId } }"
          class="ml-auto"
        />
      </UCard>
    </div>

    <div>
      <h2 class="mb-4 text-lg font-semibold">Question Pallet</h2>

      <UCard v-if="currentCategoryQuestions">
        <ul class="flex flex-wrap gap-4">
          <li v-for="(q, i) of currentCategoryQuestions" :key="q.id">
            <UButton
              :to="{ query: { ...route.query, question: q.id } }"
              :color="currentQuestionId === q.id ? 'primary' : 'neutral'"
              :variant="currentQuestionId === q.id ? 'subtle' : 'outline'"
              size="lg"
              class="flex size-10 items-center justify-center rounded-full"
            >
              {{ i + 1 }}
            </UButton>
          </li>
        </ul>
      </UCard>
    </div>
  </UContainer>
</template>

<script setup lang="ts">
import Draggable from 'vuedraggable'
import { isNullOrUndefined } from '@arpansaha13/utils'
import { ConfirmModal } from '#components'
import { QuestionType, type QuestionCategory } from '~/types'

definePageMeta({
  layout: 'cover',
})

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

watch(
  route,
  newRoute => {
    const query = newRoute.query
    if (isNullOrUndefined(query) || isNullOrUndefined(query.category)) return
    const categoryId = parseInt(query.category as string)
    lastVisitedQuestionForCategory.value[categoryId] = query.question as string
  },
  { immediate: true }
)

function getQuestionIdForCategoryId(categoryId: number) {
  const categoryQuestions = groupedQuestions.value?.[categoryId]
  if (isNullOrUndefined(categoryQuestions)) return
  const questionId =
    lastVisitedQuestionForCategory.value[categoryId] ?? categoryQuestions[0].id
  return questionId
}

// Add initial `category` and `question` queries, if missing
if (!route.query.category && sortedCategories.value?.length) {
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
  if (!currentCategoryQuestions.value || !currentQuestionId.value) return -1
  return currentCategoryQuestions.value.findIndex(
    q => q.id === currentQuestionId.value
  )
})

const question = computed(() => {
  if (currentQuestionIdx.value === -1) return null
  return currentCategoryQuestions.value?.[currentQuestionIdx.value] ?? null
})

const prevQuestionId = computed(() => {
  if (!currentCategoryQuestions.value || currentQuestionIdx.value <= 0)
    return null
  return currentCategoryQuestions.value[currentQuestionIdx.value - 1].id
})

const nextQuestionId = computed(() => {
  if (
    !currentCategoryQuestions.value ||
    currentQuestionIdx.value === -1 ||
    currentQuestionIdx.value >= currentCategoryQuestions.value.length - 1
  ) {
    return null
  }

  return currentCategoryQuestions.value[currentQuestionIdx.value + 1].id
})

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

const categoryNames = ref<Record<number, string>>({})
async function handleUpdateCategory(category: QuestionCategory) {
  categoryNames.value[category.id].trim()

  // If empty name, use default "Category {order}"
  const name = categoryNames.value[category.id] || `Category ${category.order}`

  if (name !== category.name) {
    await updateCategory(category.id, paperId, { name })
  }
}
watch(
  sortedCategories,
  newCategories => {
    if (!newCategories) return

    newCategories.forEach(category => {
      categoryNames.value[category.id] = category.name
    })
  },
  { immediate: true }
)

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
