<template>
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
})

const route = useRoute()
</script>
