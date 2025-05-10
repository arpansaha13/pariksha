<template>
  <UCard>
    <p class="font-medium">{{ question.statement }}</p>
  </UCard>

  <UCard :ui="{ root: 'grow' }">
    <URadioGroup
      v-model="mcqAnswer.optionIndex"
      :items="options"
      variant="card"
      :ui="{
        wrapper: 'ml-3',
        fieldset: 'space-y-1',
      }"
    />

    <UButton
      variant="ghost"
      :disabled="isNullOrUndefined(mcqAnswer.optionIndex)"
      :ui="{
        base: 'mt-5',
      }"
      @click="clearSelection"
    >
      Clear selection
    </UButton>
  </UCard>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { PropType } from 'vue'
import type { MCQAnswer, QuestionMcq } from '~/types'

const props = defineProps({
  question: {
    type: Object as PropType<QuestionMcq['question']>,
    required: true,
  },
})

const mcqAnswer = defineModel<MCQAnswer>('answer', {
  required: true,
})

const options = computed(() =>
  props.question.options.map((option, i) => ({
    value: i,
    label: option,
  }))
)

function clearSelection() {
  mcqAnswer.value.optionIndex = undefined
}
</script>
