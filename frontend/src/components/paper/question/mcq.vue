<template>
  <div>
    <p class="mb-4">{{ question.statement }}</p>

    <URadioGroup
      disabled
      :items="options"
      variant="card"
      :ui="{
        wrapper: 'ml-3',
        fieldset: 'space-y-1',
        label: 'opacity-100',
        item: 'opacity-100',
      }"
    />
  </div>
</template>

<script setup lang="ts">
import type { PropType } from 'vue'
import type { QuestionMcq } from '~/types'

const props = defineProps({
  question: {
    type: Object as PropType<QuestionMcq['question']>,
    required: true,
  },
})

const options = computed(() =>
  props.question.options.map((option, i) => ({
    value: i,
    label: option,
  }))
)
</script>
