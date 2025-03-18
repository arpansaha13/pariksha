<template>
  <UForm
    :state="formState"
    class="flex flex-col gap-y-5"
    @submit.prevent="onSubmit"
  >
    <UFormField label="Type" description="Choose the type of question" required>
      <USelect
        v-model="formState.type"
        :items="questionTypes"
        required
        class="w-48"
      />
    </UFormField>

    <UFormField
      label="Question"
      description="The question statement"
      name="question"
      required
    >
      <UTextarea
        v-model="formState.question.statement"
        required
        autoresize
        placeholder="Write your question here..."
        :ui="{ root: 'flex' }"
      />
    </UFormField>

    <UForm
      v-if="formState.type === QuestionType.MCQ"
      :state="formState"
      class="space-y-2"
    >
      <UFormField
        v-for="(option, i) in formState.question.options"
        :key="i"
        :label="i === 0 ? 'Options' : undefined"
        :name="`option-${i + 1}`"
      >
        <div class="flex items-center gap-2">
          <div>{{ getOptionLabel(i) }}.</div>
          <UInput
            v-model="formState.question.options[i]"
            required
            class="w-52"
          />
          <UButton
            icon="i-heroicons-trash"
            color="error"
            variant="subtle"
            size="sm"
            :disabled="formState.question.options.length <= 2"
            @click="removeOption(i)"
          />
        </div>
      </UFormField>

      <UButton
        color="neutral"
        variant="subtle"
        size="sm"
        :disabled="formState.question.options.length >= 5"
        @click="addOption"
      >
        Add option
      </UButton>
    </UForm>

    <UFormField
      label="Max score"
      description="Maximum score that can be awared for this question"
      name="max_score"
    >
      <UInputNumber v-model="formState.max_score" :min="0" required />
    </UFormField>

    <UFormField
      label="Correct answer"
      description="You can optionally write the correct answer to this question. This will only be for your reference, and will not be shown during the exam."
      name="correct_answer"
    >
      <URadioGroup
        v-if="formState.type === QuestionType.MCQ"
        v-model="formState.correct_answer"
        :items="mcqAnswerOptions"
        :ui="{
          wrapper: 'ml-3',
          fieldset: 'space-y-1',
        }"
      />
      <UTextarea
        v-else
        v-model="formState.correct_answer"
        autoresize
        placeholder="Write the answer here..."
        :ui="{ root: 'flex' }"
      />
    </UFormField>

    <button ref="submitButton" type="submit" class="hidden" />
  </UForm>
</template>

<script setup lang="ts">
import { QuestionType } from '~/types'

interface CreateQuestionFormState {
  type: QuestionType | undefined
  question: {
    statement: string
    options: string[]
  }
  max_score: number
  tags: string[]
  correct_answer: string | undefined
}

const formState = defineModel<CreateQuestionFormState>('form-data', {
  required: true,
})

// Submitting through form.submit() bypasses the native validation
// Hence using a hidden submit button
const submitButtonRef = useTemplateRef('submitButton')
defineExpose({
  submit: () => submitButtonRef.value?.click(),
})

const emit = defineEmits<{
  submit: [form: CreateQuestionFormState]
}>()

async function onSubmit() {
  emit('submit', formState.value)
}

const questionTypes = [
  {
    label: 'Multiple choice (MCQ)',
    value: QuestionType.MCQ,
  },
  {
    label: 'Short answer type',
    value: QuestionType.SHORT,
  },
  {
    label: 'Long answer type',
    value: QuestionType.LONG,
  },
]

function getOptionLabel(index: number): string {
  const ASCII_CODE_A = 65
  return String.fromCharCode(ASCII_CODE_A + index)
}

function addOption() {
  if (!formState.value.question.options) {
    formState.value.question.options = []
  }
  // Prevent adding more than 5 options
  if (formState.value.question.options.length < 5) {
    formState.value.question.options.push('')
  }
}

function removeOption(idx: number) {
  // Prevent deleting options when there are only 2 left
  if (formState.value.question.options?.length > 2) {
    formState.value.question.options.splice(idx, 1)
  }
}

const mcqAnswerOptions = computed(() =>
  formState.value.question.options.filter(Boolean).map((option, idx) => ({
    label: `${getOptionLabel(idx)}. ${option}`,
    value: option,
  }))
)
</script>
