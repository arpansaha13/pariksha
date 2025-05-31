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
      label="Parameters"
      description="Inputs and output of the program"
      :name="FieldLabels.PARAMETERS"
      required
    >
      <p class="mb-1 font-semibold">Inputs:</p>
      <ul class="mb-1 ml-2 space-y-1">
        <li
          v-for="(inputDefinition, inputIdx) in formState.question
            .input_definitions"
          :key="inputIdx"
          class="space-x-1"
        >
          <Dot />
          <span class="inline-block font-semibold">
            {{ inputDefinition.variable_name }}:
          </span>
          <span :class="['inline-block', !inputDefinition.type && 'italic']">
            {{
              inputDefinition.type
                ? getCodingQuestionParameterLabel(
                    inputDefinition.type,
                    inputDefinition.items
                  )
                : '(empty)'
            }}
          </span>
        </li>
      </ul>

      <p class="mb-2">
        <span class="font-semibold">Output: </span>

        <span
          :class="[
            'inline-block',
            !formState.question.output_definition.type && 'italic',
          ]"
        >
          {{
            formState.question.output_definition.type
              ? getCodingQuestionParameterLabel(
                  formState.question.output_definition.type,
                  formState.question.output_definition.items
                )
              : '(empty)'
          }}
        </span>
      </p>

      <PaperQuestionFormDefineParametersModal
        v-model:coding-question-content="formState.question"
        @after:leave="triggerValidate(FieldLabels.PARAMETERS)"
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
      hint="Optional"
      label="Test cases"
      :name="FieldLabels.TEST_CASES"
      description="Predefined sets of inputs and expected outputs"
    >
      <ol v-if="formState.question.test_cases" class="mb-2 space-y-1">
        <li
          v-for="(testCase, testCaseIdx) in formState.question.test_cases"
          :key="testCaseIdx"
          class="flex gap-2"
        >
          <p class="inline-block">{{ testCaseIdx + 1 }}.</p>
          <div>
            <div>
              <p class="mb-1 font-semibold">Inputs:</p>

              <ul class="mb-1 ml-2 space-y-1">
                <li
                  v-for="(testCaseInput, testCaseInputIdx) in testCase.inputs"
                  :key="testCaseInputIdx"
                  class="space-x-1"
                >
                  <Dot />
                  <span class="inline-block font-semibold">
                    {{
                      formState.question.input_definitions[testCaseInputIdx]
                        .variable_name
                    }}
                    =
                  </span>
                  <span
                    :class="[
                      'inline-block',
                      !testCaseInput ? 'italic' : 'font-mono',
                    ]"
                  >
                    {{ testCaseInput ? testCaseInput : '(empty)' }}
                  </span>
                </li>
              </ul>
            </div>

            <p>
              <span class="font-semibold">Output: </span>
              <span :class="[!testCase.output ? 'italic' : 'font-mono']">
                {{ testCase.output ? testCase.output : '(empty)' }}
              </span>
            </p>

            <p v-if="testCase.explanation" class="font-mono">
              <span class="font-semibold">Explanation: </span>
              {{ testCase.explanation }}
            </p>
          </div>
        </li>
      </ol>

      <PaperQuestionFormDefineTestCasesModal
        v-model:coding-question-content="formState.question"
        @after:leave="triggerValidate(FieldLabels.TEST_CASES)"
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

enum FieldLabels {
  PARAMETERS = 'parameters',
  TEST_CASES = 'test_cases',
}

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
    const outputDefinition = codingQuestionContent.output_definition

    if (inputDefinitions.length === 0) {
      errors.push({
        name: FieldLabels.PARAMETERS,
        message: 'Please specify the inputs for this question',
      })
    } else if (!isParameterValid(outputDefinition)) {
      errors.push({
        name: FieldLabels.PARAMETERS,
        message: 'Incomplete output definition',
      })
    } else if (inputDefinitions.some(x => !isParameterValid(x))) {
      errors.push({
        name: FieldLabels.PARAMETERS,
        message:
          'Incomplete input definition. Please complete or remove the field.',
      })
    }
  }

  function isParameterValid(parameter: QuestionCodingContentParameter) {
    const hasNoArrayItemType = (x: QuestionCodingContentParameter) =>
      x.type === QuestionCodingContentCompositeInputTypes.ARRAY &&
      !x.items?.[0].type

    if (!parameter.type || hasNoArrayItemType(parameter)) {
      return false
    }

    return true
  }

  // Trigger validate when PaperQuestionFormDefineInputsModal closes
  const formRef = useTemplateRef<ComponentExposed<typeof UForm>>('form')
  const triggerValidate = (fieldName: FieldLabels) =>
    formRef.value?.validate({ name: fieldName })

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
</script>
