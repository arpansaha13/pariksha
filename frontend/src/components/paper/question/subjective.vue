<template>
  <div>
    <p>{{ question.statement }}</p>
  </div>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { QuestionSubjectiveContent } from '~/types'

defineProps({
  question: {
    type: Object as PropType<QuestionSubjectiveContent>,
    required: true,
  },
})
</script>
