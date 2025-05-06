<template>
  <div>
    <p class="mb-4">{{ question.statement }}</p>

    <URadioGroup
      v-model="selectedOptionIndex"
      :items="options"
      disabled
      :ui="{
        wrapper: 'ml-3',
        fieldset: 'space-y-1',
      }"
    />
  </div>
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
