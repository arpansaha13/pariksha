<template>
  <UForm
    ref="form"
    :state="formState"
    :validate="validate"
    :validate-on="['blur']"
    class="flex flex-col gap-y-5"
    @submit.prevent="onSubmit"
  >
    <UFormField
      name="type"
      label="Type"
      description="Choose the type of question"
      required
    >
      <USelect
        v-model="formState.type"
        :items="questionTypes"
        required
        class="w-48"
      />
    </UFormField>

    <UFormField
      v-if="formState.type === QuestionType.CODING"
      name="title"
      label="Title"
      description="The question title"
      required
    >
      <UInput v-model="formState.question.title" required class="w-full" />
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

    <UFormField
      v-if="formState.type === QuestionType.CODING"
      label="Inputs"
      description="Inputs given to the program"
      name="input_definitions"
      required
    >
      <ul class="mb-2 ml-2 space-y-1">
        <li
          v-for="(inputDefinition, inputIdx) in formState.question
            .input_definitions"
          :key="inputIdx"
          class="space-x-1"
        >
          <span class="inline-block">{{ inputIdx + 1 }}.</span>
          <span class="inline-block font-semibold">
            {{
              inputDefinition.variableName ??
              getDefaultCodingQuestionInputVariableName(inputIdx + 1)
            }}:
          </span>
          <span :class="['inline-block', !inputDefinition.type && 'italic']">
            {{
              inputDefinition.type
                ? getInputDefinitionLabel(
                    inputDefinition.type,
                    inputDefinition.items
                  )
                : '(empty)'
            }}
          </span>
        </li>
      </ul>

      <PaperQuestionFormDefineInputsModal
        v-model:coding-question-content="formState.question"
        @after:leave="triggerValidate"
      />
    </UFormField>

    <div v-if="formState.type === QuestionType.MCQ" class="space-y-2">
      <UForm
        v-for="(_, optionIdx) in formState.question.options"
        :key="optionIdx"
        :state="formState.question"
        attach
      >
        <UFormField
          :label="optionIdx === 0 ? 'Options' : undefined"
          :name="`option-${optionIdx + 1}`"
        >
          <div class="flex items-center gap-2">
            <div>{{ getOptionLabel(optionIdx) }}.</div>
            <UInput
              v-model="formState.question.options[optionIdx]"
              required
              class="w-52"
            />
            <UButton
              icon="i-heroicons-trash"
              size="sm"
              color="error"
              variant="subtle"
              aria-label="Remove option"
              :disabled="
                formState.question.options.length <= MIN_MCQ_OPTIONS_COUNT
              "
              @click="removeOption(optionIdx)"
            />
          </div>
        </UFormField>
      </UForm>

      <UButton
        label="Add option"
        color="neutral"
        variant="subtle"
        size="sm"
        :disabled="formState.question.options.length >= MAX_MCQ_OPTIONS_COUNT"
        @click="addOption"
      />
    </div>

    <UFormField
      v-else-if="formState.type === QuestionType.CODING"
      label="Test cases"
      description="Predefined sets of inputs and expected outputs"
      hint="Optional"
      :ui="{ container: 'flex flex-col gap-y-2' }"
    >
      <UForm
        v-for="(testCase, testCaseIdx) in formState.question.examples"
        :key="testCaseIdx"
        :state="testCase"
        attach
        class="space-y-1"
      >
        <UFormField
          label="Input"
          :name="`test-case-${testCaseIdx + 1}-input`"
          :ui="{ labelWrapper: 'ml-5' }"
          required
        >
          <div class="flex items-center gap-2">
            <div>{{ testCaseIdx + 1 }}.</div>
            <UInput v-model="testCase.input" required class="w-64" />
            <UButton
              icon="i-heroicons-trash"
              size="sm"
              color="error"
              variant="subtle"
              aria-label="Remove example"
              class="ml-auto"
              @click="removeExample(testCaseIdx)"
            />
          </div>
        </UFormField>
        <UFormField
          label="Output"
          :name="`test-case-${testCaseIdx + 1}-output`"
          :ui="{ root: 'ml-5' }"
          required
        >
          <UInput v-model="testCase.output" required class="w-64" />
        </UFormField>
        <UFormField
          label="Explanation"
          name="correct_answer"
          hint="Optional"
          :ui="{ root: 'ml-5' }"
        >
          <UTextarea
            v-model="testCase.explanation"
            :rows="2"
            :name="`test-case-${testCaseIdx + 1}-explanation`"
            placeholder="Write the explanation here..."
            autoresize
            :ui="{ root: 'flex' }"
          />
        </UFormField>
      </UForm>

      <UButton
        color="neutral"
        variant="subtle"
        size="sm"
        label="Add example"
        class="w-max"
        :disabled="
          formState.question.examples!.length >= MAX_CODING_EXAMPLES_COUNT
        "
        @click="addExample"
      />
    </UFormField>

    <UFormField
      label="Max score"
      description="Maximum score that can be awarded for this question"
      name="max_score"
      required
    >
      <UInputNumber
        v-model="formState.max_score"
        :min="0"
        :max="MAX_SCORE_PER_QUESTION"
        required
      />
    </UFormField>

    <UFormField
      label="Correct answer"
      description="You can optionally write the correct answer to this question. This will only be for your reference, and will not be shown during the exam."
      name="correct_answer"
      hint="Optional"
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
import { isNullOrUndefined } from '@arpansaha13/utils'
import type { FormError } from '@nuxt/ui'
import type { ComponentExposed } from 'vue-component-type-helpers'
import type { UForm } from '#components'

const formState = defineModel<MergedQuestion>('form-data', {
  required: true,
})

// Submitting through form.submit() bypasses the native validation
// Hence using a hidden submit button
const submitButtonRef = useTemplateRef('submitButton')
defineExpose({
  submit: () => submitButtonRef.value?.click(),
})

const emit = defineEmits<{
  submit: [form: MergedQuestion]
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
    label: 'Subjective',
    value: QuestionType.SUBJECTIVE,
  },
  {
    label: 'Coding',
    value: QuestionType.CODING,
  },
]

// _________________________FORM VALIDATION_________________________
function usePaperQuestionFormValidate() {
  function validate(formState: MergedQuestion): FormError[] {
    const errors: FormError[] = []

    validateCommon(formState, errors)

    if (formState.type === QuestionType.CODING) {
      validateCodingQuestion(formState.question, errors)
    }

    return errors
  }

  /** Common validation for all question types */
  function validateCommon(
    formState: Omit<BaseQuestion, MergedQuestionOmit>,
    errors: FormError[]
  ) {
    if (formState.max_score === 0) {
      errors.push({
        name: 'max_score',
        message: 'Please specify the maximum score for this question',
      })
    } else if (formState.max_score > MAX_SCORE_PER_QUESTION) {
      errors.push({
        name: 'max_score',
        message: 'Maximum score cannot be greater than 1000',
      })
    }
  }

  function validateCodingQuestion(
    codingQuestionContent: QuestionCodingContent,
    errors: FormError[]
  ) {
    const inputDefinitions = codingQuestionContent.input_definitions

    if (inputDefinitions.length === 0) {
      errors.push({
        name: 'input_definitions',
        message: 'Please specify the inputs for this question',
      })
    } else {
      const hasNoArrayItemType = (x: QuestionCodingContentInputDefinition) =>
        x.type === QuestionCodingContentCompositeInputTypes.ARRAY &&
        !x.items?.[0].type

      if (inputDefinitions.some(x => !x.type || hasNoArrayItemType(x)))
        errors.push({
          name: 'input_definitions',
          message:
            'Incomplete input definition. Please complete or remove the field.',
        })
    }
  }

  // Trigger validate when PaperQuestionFormDefineInputsModal closes
  const formRef = useTemplateRef<ComponentExposed<typeof UForm>>('form')
  const triggerValidate = () =>
    formRef.value?.validate({ name: 'input_definitions' })

  return { validate, triggerValidate }
}

const { validate, triggerValidate } = usePaperQuestionFormValidate()

// _________________________MCQ OPTIONS___________________________
function getOptionLabel(index: number): string {
  const ASCII_CODE_A = 65
  return String.fromCharCode(ASCII_CODE_A + index)
}

function addOption() {
  if (formState.value.type !== QuestionType.MCQ) {
    logWarning('addOption called without MCQ type question')
    return
  }

  if (isNullOrUndefined(formState.value.question.options)) {
    formState.value.question.options = []
  }
  // Prevent adding more than 5 options
  if (formState.value.question.options.length < MAX_MCQ_OPTIONS_COUNT) {
    formState.value.question.options.push('')
  }
}

function removeOption(idx: number) {
  if (formState.value.type !== QuestionType.MCQ) {
    logWarning('removeOption called without MCQ type question')
    return
  }

  // Prevent deleting options when there are only 2 left
  if (formState.value.question.options?.length > MIN_MCQ_OPTIONS_COUNT) {
    formState.value.question.options.splice(idx, 1)
  }
}

const mcqAnswerOptions = computed(() => {
  if (formState.value.type !== QuestionType.MCQ) {
    logWarning('mcqAnswerOptions accessed without MCQ type question')
    return []
  }

  return formState.value.question.options
    .filter(Boolean)
    .map((option, idx) => ({
      label: `${getOptionLabel(idx)}. ${option}`,
      value: option,
    }))
})

// ___________________CODING QUESTION EXAMPLES____________________
function addExample() {
  if (formState.value.type !== QuestionType.CODING) {
    logWarning('addExample called without CODING type question')
    return
  }

  if (isNullOrUndefined(formState.value.question.examples)) {
    formState.value.question.examples = []
  }
  // Prevent adding more than 4 examples
  if (formState.value.question.examples.length < MAX_CODING_EXAMPLES_COUNT) {
    formState.value.question.examples.push({
      input: '',
      output: '',
      explanation: '',
    })
  }
}

function removeExample(idx: number) {
  if (formState.value.type !== QuestionType.CODING) {
    logWarning('removeExample called without CODING type question')
    return
  }

  if (!isNullOrUndefined(formState.value.question.examples)) {
    formState.value.question.examples.splice(idx, 1)
  }
}
</script>
