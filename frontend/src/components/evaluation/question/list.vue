<template>
  <ul class="flex flex-wrap gap-4">
    <li v-for="(item, i) of currentCategoryQuestions" :key="item.question.id">
      <UButton
        replace
        :to="{ query: { ...route.query, question: item.question.id } }"
        :color="currentQuestionId === item.question.id ? 'primary' : 'neutral'"
        :variant="currentQuestionId === item.question.id ? 'subtle' : 'outline'"
        size="lg"
        class="flex size-10 items-center justify-center rounded-full"
      >
        {{ i + 1 }}
      </UButton>
    </li>
  </ul>
</template>

<script setup lang="ts">
import type { QuestionAnswer } from '~/types'

defineProps({
  currentQuestionId: {
    type: Number,
    required: true,
  },
  currentCategoryQuestions: {
    type: Array as PropType<QuestionAnswer[]>,
    required: true,
  },
})

const route = useRoute()
</script>
