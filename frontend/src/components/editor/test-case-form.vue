<template>
  <UForm :state="testCase" class="flex flex-col gap-4">
    <UFormField
      v-for="(_, inputIdx) in testCase.inputs"
      :key="inputIdx"
      :label="`Input ${inputIdx + 1}`"
      :name="`test-case-${testCaseIdx + 1}-input-${inputIdx + 1}`"
    >
      <UInput
        v-model="testCase.inputs[inputIdx]"
        size="lg"
        variant="subtle"
        class="w-full font-mono"
      />
    </UFormField>
    <UFormField
      label="Output"
      :name="`test-case-${testCaseIdx + 1}-expected-output`"
    >
      <UInput
        v-model="testCase.expectedOutput"
        size="lg"
        variant="subtle"
        class="w-full font-mono"
      />
    </UFormField>
  </UForm>
</template>

<script lang="ts" setup>
defineProps<{
  testCaseIdx: number
}>()

const testCase = defineModel<TestCase>('test-case', { required: true })
</script>
