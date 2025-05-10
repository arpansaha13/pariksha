<template>
  <UCard>
    <p class="font-medium">{{ question.statement }}</p>
  </UCard>

  <UCard>
    <div v-if="answer">
      <p>
        {{ answer.text }}
      </p>
    </div>
  </UCard>

  <UCard :ui="{ root: 'grow' }">
    <UFormField
      label="Score"
      description="Score to be awarded for this answer"
      name="score_awarded"
      required
    >
      <UInputNumber :min="0" :max="MAX_SCORE_PER_QUESTION" required />
    </UFormField>
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
  answer: {
    type: [Object, null] as PropType<GeneralAnswer | null>,
    default: null, // the question may be unanswered
  },
})
</script>
