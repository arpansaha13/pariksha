<template>
  <div>
    <p class="mb-4">{{ question.statement }}</p>

    <URadioGroup
      v-model="selected"
      :items="options"
      :ui="{
        wrapper: 'ml-3',
        fieldset: 'space-y-1',
      }"
    />

    <UButton
      variant="ghost"
      :disabled="isNullOrUndefined(selected)"
      :ui="{
        base: 'mt-5',
      }"
      @click="clearSelection"
    >
      Clear selection
    </UButton>
  </div>
</template>

<script setup lang="ts">
import { isNullOrUndefined } from '@arpansaha13/utils'
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

const selected = ref<number>()

function clearSelection() {
  selected.value = undefined
}
</script>
