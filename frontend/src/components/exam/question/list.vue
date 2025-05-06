<template>
  <ul class="flex flex-wrap gap-4">
    <li v-for="(q, i) of currentCategoryQuestions" :key="q.id">
      <UButton
        :color="currentQuestionId === q.id ? 'primary' : 'neutral'"
        :variant="currentQuestionId === q.id ? 'subtle' : 'outline'"
        size="lg"
        class="flex size-10 items-center justify-center rounded-full"
        @click="
          saveAndNavigateTo(
            { query: { ...route.query, question: q.id } },
            { replace: true }
          )
        "
      >
        {{ i + 1 }}
      </UButton>
    </li>
  </ul>
</template>

<script setup lang="ts">
import type { ExamQuestionMinimal } from '~/types'

defineProps({
  currentQuestionId: {
    type: Number,
    required: true,
  },
  currentCategoryQuestions: {
    type: Array as PropType<ExamQuestionMinimal[]>,
    required: true,
  },
  saveAndNavigateTo: {
    type: Function as PropType<typeof navigateTo>,
    required: true,
  },
})

const route = useRoute()
</script>
