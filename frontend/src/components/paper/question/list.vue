<template>
  <Draggable
    v-model="categoryQuestionsCopyForReorder"
    tag="ol"
    item-key="id"
    group="question-editable"
    handle=".draggable-handle"
    class="divide-y divide-neutral-200 py-2 transition-colors"
    ghost-class="draggable-ghost"
    drag-class="draggable-hold"
    :animation="250"
    @start="isDragging = true"
    @end="onQuestionsReorderEnd"
  >
    <template #item="{ element: q }">
      <li
        :key="q.id"
        :class="[
          currentQuestionId === q.id
            ? 'bg-(--ui-primary)/10 hover:bg-(--ui-primary)/15 disabled:bg-(--ui-primary)/10 aria-disabled:bg-(--ui-primary)/10'
            : 'bg-(--ui-bg) hover:bg-(--ui-bg-elevated) disabled:bg-(--ui-bg) aria-disabled:bg-(--ui-bg)',
        ]"
      >
        <!-- //NOSONAR gives error because this is a comment node and counts as multiple nodes inside template -->
        <UChip
          :show="!!showQuestionEditChip[q.id]"
          inset
          position="top-left"
          :ui="{ root: 'flex gap-2 px-2', base: 'top-1 left-1' }"
        >
          <div class="draggable-handle flex size-6 shrink-0 cursor-grab">
            <Icon
              name="i-heroicons-bars-2"
              class="m-auto text-gray-400 transition-colors"
            />
          </div>

          <ULink
            :to="{ query: { ...route.query, question: q.id } }"
            raw
            replace
            exact-query
            class="block grow py-2.5 text-sm"
            active-class="text-(--ui-primary)"
            inactive-class="text-(--ui-text)"
          >
            <span class="line-clamp-2">
              {{ q.question.statement }}
            </span>
          </ULink>

          <UButton
            icon="i-heroicons-trash"
            size="sm"
            color="error"
            square
            variant="ghost"
            loading-auto
            class="shrink-0"
            @click="handleDeleteQuestion(q.id)"
          />
        </UChip>
      </li>
    </template>
  </Draggable>
</template>

<script setup lang="ts">
import Draggable from 'vuedraggable'

const props = defineProps<{
  currentQuestionId: QuestionId
  currentCategoryId: CategoryId
  currentCategoryQuestions: QuestionMinimal[]
  showQuestionEditChip: Record<QuestionId, boolean>
  questionNavigation: Record<'prev' | 'next', number | null | QuestionId>
}>()

const route = useRoute()
const paperId = route.params.paperId as PaperId

const confirmModal = inject(InjectionKeys.ConfirmModal)!

// ________________________REORDER QUESTIONS________________________
const isDragging = ref(false)
const currentCategoryQuestions = toRef(props, 'currentCategoryQuestions')
const categoryQuestionsCopyForReorder = shallowRef<QuestionMinimal[]>([])

watchImmediate(
  currentCategoryQuestions,
  val => {
    if (val) categoryQuestionsCopyForReorder.value = [...val]
  },
  { deep: true }
)

function onQuestionsReorderEnd() {
  isDragging.value = false

  if (!props.currentCategoryId) return
  let isReordered = false

  // Check if the order was changed
  for (let i = 0; i < props.currentCategoryQuestions.length; i++) {
    if (
      props.currentCategoryQuestions[i].id !==
      categoryQuestionsCopyForReorder.value[i].id
    ) {
      isReordered = true
      break
    }
  }

  if (isReordered) {
    reorderQuestions(
      paperId,
      props.currentCategoryId,
      categoryQuestionsCopyForReorder.value.map(q => q.id)
    )
  }
}

// ________________________DELETE QUESTION________________________
async function handleDeleteQuestion(questionId: QuestionId) {
  if (!props.currentCategoryId) return

  const instance = confirmModal.open({
    title: 'Confirm question deletion',
    description:
      'Are you sure you want to delete this question? This action cannot be undone.',
    confirmLabel: 'Delete question',
  })

  const shouldDelete = await instance.result
  if (!shouldDelete) return

  // If deleting current question, switch to another one first
  if (props.currentQuestionId === questionId) {
    await navigateTo({
      query: {
        ...route.query,
        question:
          props.questionNavigation.prev ??
          props.questionNavigation.next ??
          QUESTION_ID_ADD,
      },
    })
  }

  await deleteQuestion(questionId, paperId, props.currentCategoryId)
}
</script>
