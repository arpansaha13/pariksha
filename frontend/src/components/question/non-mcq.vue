<template>
  <div>
    <p>{{ question.statement }}</p>
  </div>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { QuestionLong, QuestionShort } from '~/types'

type QuestionNonMcq = QuestionShort | QuestionLong

defineProps({
  question: {
    type: Object as PropType<QuestionNonMcq['question']>,
    required: true,
  },
})
</script>
