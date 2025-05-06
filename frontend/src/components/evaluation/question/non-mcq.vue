<template>
  <div>
    <p>{{ question.statement }}</p>

    <USeparator :ui="{ root: 'my-4' }" />

    <div v-if="answer">
      <p>
        {{ answer.text }}
      </p>
    </div>
  </div>
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
  answer: {
    type: [Object, null] as PropType<GeneralAnswer | null>,
    default: null, // the question may be unanswered
  },
})
</script>
