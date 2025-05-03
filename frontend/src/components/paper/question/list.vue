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
          :show="!isNullOrUndefined(editQuestionFormStates[q.id])"
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
import { isNullOrUndefined } from '@arpansaha13/utils'
import { QuestionId, type QuestionMinimal } from '~/types'

const props = defineProps({
  currentQuestionId: {
    type: Number,
    required: true,
  },
  currentCategoryId: {
    type: Number,
    required: true,
  },
  currentCategoryQuestions: {
    type: Array as PropType<QuestionMinimal[]>,
    required: true,
  },
  editQuestionFormStates: {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    type: Object as PropType<Record<number, any>>,
    required: true,
  },
  questionNavigation: {
    type: Object as PropType<
      Record<'prev' | 'next', number | null | QuestionId>
    >,
    required: true,
  },
})

const route = useRoute()
const paperId = parseInt(route.params.paperId as string)

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
async function handleDeleteQuestion(questionId: number) {
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
          QuestionId.ADD,
      },
    })
  }

  await deleteQuestion(questionId, paperId, props.currentCategoryId)
}
</script>
