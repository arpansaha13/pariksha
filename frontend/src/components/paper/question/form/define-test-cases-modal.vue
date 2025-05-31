<template>
  <UModal
    title="Define test cases"
    description="Predefined sets of inputs and expected outputs"
    :ui="{ body: 'divide-y divide-default' }"
    @update:open="handleAutoAddTestCase"
    @after:leave="emit('after:leave')"
  >
    <UButton
      icon="lucide:test-tube-diagonal"
      size="sm"
      color="neutral"
      variant="subtle"
      label="Define test cases"
    />

    <template #body>
      <EmptyState
        v-if="codingQuestionContent.test_cases?.length === 0"
        icon="lucide:test-tube-diagonal"
        description="No test cases to show"
      />
      <UForm
        v-for="(testCase, testCaseIdx) in codingQuestionContent.test_cases"
        :key="testCaseIdx"
        :state="testCase"
        attach
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
    </template>

    <template #footer>
      <UButton
        color="neutral"
        variant="subtle"
        label="Add test case"
        :disabled="
          codingQuestionContent.test_cases!.length >= MAX_CODING_EXAMPLES_COUNT
        "
        @click="addTestCase"
      />
    </template>
  </UModal>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

const codingQuestionContent = defineModel<
  Pick<
    QuestionCodingContent,
    'test_cases' | 'input_definitions' | 'output_definition'
  >
>('coding-question-content', {
  required: true,
})

const emit = defineEmits(['after:leave'])

// __________________CODING QUESTION TEST CASES___________________
function addTestCase() {
  if (isNullOrUndefined(codingQuestionContent.value.test_cases)) {
    codingQuestionContent.value.test_cases = []
  }
  // Prevent adding more than 4 test cases
  if (
    codingQuestionContent.value.test_cases.length < MAX_CODING_EXAMPLES_COUNT
  ) {
    codingQuestionContent.value.test_cases.push({
      inputs: Array.from({
        length: codingQuestionContent.value.input_definitions.length,
      }),
      output: '',
      explanation: '',
    })
  }
}

function removeTestCase(idx: number) {
  if (!isNullOrUndefined(codingQuestionContent.value.test_cases)) {
    codingQuestionContent.value.test_cases.splice(idx, 1)
  }
}

// _________________AUTO ADD TEST CASE WHEN EMPTY_________________
const autoAddedOnOpen = ref(true)

function handleAutoAddTestCase(open: boolean) {
  if (isNullOrUndefined(codingQuestionContent.value.test_cases)) {
    codingQuestionContent.value.test_cases = []
  }

  const testCases = codingQuestionContent.value.test_cases

  if (open) {
    if (testCases.length === 0) {
      testCases.push({
        inputs: Array.from({
          length: codingQuestionContent.value.input_definitions.length,
        }),
        output: '',
        explanation: '',
      })
      autoAddedOnOpen.value = true
    }
  } else {
    if (
      autoAddedOnOpen.value &&
      testCases.length === 1 &&
      !testCases[0].output &&
      !testCases[0].explanation &&
      Object.values(testCases[0].inputs).every(val => !val) // if all inputs are empty
    ) {
      console.log(testCases[0])
      testCases.length = 0
    }
    autoAddedOnOpen.value = false
  }
}
</script>
