<template>
  <UCard>
    <p class="font-medium">{{ question.statement }}</p>
  </UCard>

  <UCard>
    <URadioGroup
      v-model="selectedOptionIndex"
      :items="options"
      variant="card"
      disabled
      :ui="{
        wrapper: 'ml-3',
        fieldset: 'space-y-1',
      }"
    />
  </UCard>

  <UCard :ui="{ root: 'flex-grow' }">
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
import type { MCQAnswer, QuestionMcq } from '~/types'

const props = defineProps({
  question: {
    type: Object as PropType<QuestionMcq['question']>,
    required: true,
  },
  answer: {
    type: [Object, null] as PropType<MCQAnswer | null>,
    default: null, // the question may be unanswered
  },
})

const selectedOptionIndex = ref(props.answer?.optionIndex)

watch(
  () => props.answer,
  val => (selectedOptionIndex.value = val?.optionIndex)
)

const options = computed(() =>
  props.question.options.map((option, i) => ({
    value: i,
    label: option,
  }))
)
</script>
