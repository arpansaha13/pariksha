<template>
  <UCard>
    <p class="font-medium">{{ question.statement }}</p>
  </UCard>

  <UCard :ui="{ root: 'grow' }">
    <UTextarea
      v-model="generalAnswer.text"
      autoresize
      placeholder="Write your answer here..."
      :ui="{ root: 'flex' }"
    />
  </UCard>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { GeneralAnswer, QuestionLong, QuestionShort } from '~/types'

type QuestionNonMcq = QuestionShort | QuestionLong

defineProps({
  question: {
    type: Object as PropType<QuestionNonMcq['question']>,
    required: true,
  },
})

const generalAnswer = defineModel<GeneralAnswer>('answer', {
  required: true,
})
</script>
