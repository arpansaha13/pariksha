<template>
  <UModal
    title="Manage categories"
    description="Add, edit, remove, or reorder your categories"
    :ui="{
      body: 'space-y-2',
    }"
    @after:leave="handleReorderCategories"
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
      <Draggable
        v-model="categoriesCopyForReorder"
        tag="ol"
        item-key="id"
        group="category-editable"
        handle=".draggable-handle"
        ghost-class="draggable-ghost"
        drag-class="draggable-hold"
        :animation="250"
        @start="isDragging = true"
        @end="isDragging = false"
      >
        <template #item="{ element: category }">
          <li
            :key="category.id"
            class="group flex items-center gap-2 rounded-sm"
          >
            <!-- //NOSONAR gives error because this is a comment node and counts as multiple nodes inside template -->
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
                <EditablePreview />
                <EditableInput class="outline-none" />
              </EditableArea>
            </EditableRoot>

            <div v-show="!isDragging">
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
</template>

<script setup lang="ts">
import Draggable from 'vuedraggable'

const props = defineProps<{
  sortedCategories: QuestionCategory[]
  groupedQuestions: Record<number, QuestionMinimal[]>
  currentCategoryId: CategoryId
}>()

const route = useRoute()
const paperId = route.params.paperId as PaperId

const paperStore = usePaperStore()

const isDragging = ref(false)
const sortedCategories = toRef(props, 'sortedCategories')
const confirmModal = inject(InjectionKeys.ConfirmModal)!

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

// _______________________DELETE CATEGORY________________________
async function handleDeleteCategory(category: QuestionCategory) {
  let shouldDelete = true

  // If category has questions, then show a confirmation modal
  const categoryQuestions = props.groupedQuestions[category.id]
  if (categoryQuestions && categoryQuestions.length > 0) {
    const instance = confirmModal.open({
      title: 'Confirm category deletion',
      description: `This category "${category.name}" has ${categoryQuestions.length} questions which will be deleted along with it.`,
      confirmLabel: 'Delete category',
    })
    shouldDelete = await instance.result
  }

  if (shouldDelete) {
    // If deleting current category, switch to another category first
    if (props.currentCategoryId === category.id) {
      // Find another category to switch to
      const nextCategory = props.sortedCategories?.find(
        c => c.id !== category.id
      )
      if (nextCategory) {
        await navigateTo({
          query: {
            category: nextCategory.id,
            question: paperStore.getQuestionIdForCategoryId(nextCategory.id),
          },
        })
      }
    }

    return deleteCategory(category.id, paperId)
  }
}

// _______________________REORDER CATEGORY________________________
/**
 * `reorder-categories` api is fired after the modal closes, so that all reorders can be batched.
 * Because individual reorders cause inconsistencies during rollback (in case of error).
 */

const categoriesCopyForReorder = shallowRef<QuestionCategory[]>([])

watchImmediate(
  sortedCategories,
  val => {
    if (val) categoriesCopyForReorder.value = [...val]
  },
  { deep: true }
)

function handleReorderCategories() {
  let isReordered = false

  // Check if the order was changed
  for (let i = 0; i < props.sortedCategories.length; i++) {
    if (props.sortedCategories[i].id !== categoriesCopyForReorder.value[i].id) {
      isReordered = true
      break
    }
  }

  if (isReordered) {
    reorderCategories(
      paperId,
      categoriesCopyForReorder.value.map(cat => cat.id)
    )
  }
}
</script>
