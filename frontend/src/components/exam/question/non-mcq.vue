<template>
  <div>
    <div>
      <p>{{ question.statement }}</p>
    </div>

    <USeparator :ui="{ root: 'my-4' }" />

    <UTextarea
      v-model="answer"
      autoresize
      placeholder="Write your answer here..."
      :ui="{ root: 'flex' }"
    />
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

const answer = ref('')
</script>
