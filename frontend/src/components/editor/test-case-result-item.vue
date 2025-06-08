<template>
  <div class="space-y-4">
    <DisplayCodeBlock
      v-if="result.error"
      color="error"
      padding="small"
      preserve-white-space
    >
      <p class="text-error-500 text-sm">
        {{ result.error }}
      </p>
    </DisplayCodeBlock>

    <div>
      <div class="mb-1 text-sm font-semibold">Inputs</div>

      <ul class="space-y-1.5">
        <DisplayCodeBlock
          v-for="(inputVal, inputIdx) in result.inputs"
          :key="inputIdx"
          as="li"
          padding="small"
        >
          <p class="text-xs font-medium">
            {{ inputDefinitions[inputIdx].variable_name + ' =' }}
          </p>
          <p class="mt-1">{{ inputVal }}</p>
        </DisplayCodeBlock>
      </ul>
    </div>

    <DisplayCodeBlock v-if="result.stdout" preserve-white-space padding="small">
      <template #header> Stdout </template>
      <p>{{ result.stdout }}</p>
    </DisplayCodeBlock>

    <DisplayCodeBlock v-if="result.output" padding="small">
      <template #header> Output </template>
      <p>{{ result.output }}</p>
    </DisplayCodeBlock>

    <DisplayCodeBlock padding="small">
      <template #header> Expected output </template>
      <p>{{ result.expected_output }}</p>
    </DisplayCodeBlock>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  result: EngineRunTestCaseResult
  resultIdx: number
  inputDefinitions: QuestionCodingContentInputDefinition[]
}>()
</script>
