<template>
  <div class="space-y-4">
    <DisplayCodeBlock
      v-if="runResult.stderr"
      color="error"
      preserve-white-space
    >
      <template #header> Stderr </template>
      <p class="text-error-500">
        {{ runResult.stderr }}
      </p>
    </DisplayCodeBlock>

    <DisplayCodeBlock v-if="runResult.stdout" preserve-white-space>
      <template #header> Stdout </template>
      <p>{{ runResult.stdout }}</p>
    </DisplayCodeBlock>
  </div>
</template>

<script setup lang="ts">
defineProps<{
  runResult: EngineRunResult
}>()
</script>
