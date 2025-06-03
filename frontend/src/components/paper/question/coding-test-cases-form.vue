<template>
  <article class="border-default border-b pb-4">
    <h2 class="heading mb-4">
      {{ codingQuestionContent.title }}
    </h2>

    <p>
      {{ codingQuestionContent.statement }}
    </p>
  </article>

  <UForm
    :state="testCases"
    class="divide-default divide-y py-2"
    @submit.prevent="onSubmit"
  >
    <UForm
      v-for="(testCase, testCaseIdx) in testCases"
      :key="testCaseIdx"
      attach
      :state="testCase"
      class="space-y-1.5 py-2 first:pt-0 last:pb-0"
    >
      <div class="flex items-center justify-between">
        <h3 class="text-sm font-bold">Test case {{ testCaseIdx + 1 }}</h3>

        <UButton
          icon="heroicons:trash"
          size="sm"
          color="error"
          variant="subtle"
          class="flex"
          :disabled="testCases.length === 1"
          @click="removeTestCase(testCaseIdx)"
        />
      </div>

      <UFormField
        v-for="(_, testCaseInputIdx) in testCase.inputs"
        :key="testCaseInputIdx"
        :label="
          codingQuestionContent.input_definitions[testCaseInputIdx]
            .variable_name
        "
        :name="`test-case-${testCaseIdx + 1}-input-${testCaseInputIdx + 1}`"
        required
      >
        <UInput
          v-model="testCase.inputs[testCaseInputIdx]"
          required
          class="w-64 font-mono"
        />
      </UFormField>

      <UFormField
        label="Output"
        :name="`test-case-${testCaseIdx + 1}-output`"
        required
      >
        <UInput v-model="testCase.output" required class="w-64 font-mono" />
      </UFormField>
      <UFormField
        hint="Optional"
        label="Explanation"
        :name="`test-case-${testCaseIdx + 1}-explanation`"
      >
        <UTextarea
          v-model="testCase.explanation"
          :rows="2"
          placeholder="Write the explanation here..."
          autoresize
          :ui="{ root: 'flex font-mono' }"
        />
      </UFormField>
    </UForm>

    <UButton
      color="neutral"
      variant="subtle"
      label="Add test case"
      :disabled="testCases!.length >= MAX_CODING_TEST_CASES_COUNT"
      @click="addTestCase"
    />

    <button ref="submitButton" type="submit" class="hidden" />
  </UForm>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

const props = defineProps<{
  codingQuestionContent: Pick<
    QuestionCodingContent,
    'title' | 'statement' | 'input_definitions' | 'output_definition'
  >
}>()

const testCases = defineModel<PartialQuestionCodingTestCase[]>('test-cases', {
  required: true,
})

const submitButtonRef = useTemplateRef('submitButton')
defineExpose({
  submit: () => submitButtonRef.value?.click(),
})

const emit = defineEmits<{
  submit: [form: PartialQuestionCodingTestCase[]]
}>()

async function onSubmit() {
  emit('submit', testCases.value)
}

// __________________CODING QUESTION TEST CASES___________________
function addTestCase() {
  // Prevent adding more than 4 test cases
  if (testCases.value.length < MAX_CODING_TEST_CASES_COUNT) {
    testCases.value.push({
      inputs: Array.from({
        length: props.codingQuestionContent.input_definitions.length,
      }),
      output: '',
      explanation: '',
      hidden: false,
    })
  }
}

function removeTestCase(idx: number) {
  if (!isNullOrUndefined(testCases.value)) {
    testCases.value.splice(idx, 1)
  }
}
</script>
