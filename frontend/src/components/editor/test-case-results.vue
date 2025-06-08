<template>
  <!-- Can't display here without v-model on UTabs (which is not working properly currently) -->
  <!-- <div class="mb-1 flex items-center justify-between gap-2">
    <p :class="['text-xl font-semibold', selectedTestCaseStatusColor]">
      {{ selectedTestCaseStatusText }}
    </p>
    <p class="text-sm">Execution time: {{ selectedTestCaseExecutionTime }}</p>
  </div> -->

  <UTabs :items="engineRunResult.results" size="sm" variant="link">
    <template #default="{ index: resultIdx }">
      Test case {{ resultIdx + 1 }}
    </template>

    <template #content="{ item: testCaseResult, index: resultIdx }">
      <div class="mt-1 mb-3 flex items-center justify-between gap-2">
        <p
          :class="['text-xl font-semibold', getTestCaseStatusColor(resultIdx)]"
        >
          {{ getTestCaseStatusText(resultIdx) }}
        </p>
        <p class="text-sm">
          Execution time: {{ getTestCaseExecutionTime(resultIdx) }}
        </p>
      </div>

      <EditorTestCaseResultItem
        :result="testCaseResult"
        :result-idx="resultIdx"
        :input-definitions="questionData.question.input_definitions"
      />
    </template>
  </UTabs>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

const props = defineProps<{
  engineRunResult: EngineRunResponse
  questionData: QuestionCoding
}>()

// ______________________SELECTED TAB GETTERS_______________________

// v-model on UTabs is not working correctly
// falling back to getter functions for selected tab data

const testCaseStatusTextMap = Object.freeze({
  [EngineRunResultStatus.UNKNOWN]: '',
  [EngineRunResultStatus.SUCCESS]: 'Success',
  [EngineRunResultStatus.WRONG_ANSWER]: 'Wrong answer',
  [EngineRunResultStatus.RUNTIME_ERROR]: 'Runtime error',
} as const)

const getTestCaseStatusText = (selectedTabIdx: number) => {
  if (isNullOrUndefined(selectedTabIdx)) return ''

  const status = props.engineRunResult.results[selectedTabIdx].status

  if (!isNullOrUndefined(testCaseStatusTextMap[status])) {
    return testCaseStatusTextMap[status]
  }

  logWarning(`Unknown value for EngineRunResult status: ${status}`)
  return ''
}

const getTestCaseStatusColor = (selectedTabIdx: number) => {
  if (isNullOrUndefined(selectedTabIdx)) {
    return 'text-primary-500'
  }

  const isSuccess =
    props.engineRunResult.results[selectedTabIdx].status ===
    EngineRunResultStatus.SUCCESS

  if (isSuccess) return 'text-primary-500'
  return 'text-error-500'
}

const getTestCaseExecutionTime = (selectedTabIdx: number) => {
  if (isNullOrUndefined(selectedTabIdx)) return null

  return `${props.engineRunResult.results[selectedTabIdx].execution_time} ms`
}
</script>
