<template>
  <article>
    <div class="mb-4 flex items-center justify-between gap-x-4">
      <h2 class="heading">
        {{ content.title }}
      </h2>

      <UButton
        v-if="editorLink"
        :to="editorLink"
        size="sm"
        color="neutral"
        variant="outline"
        icon="lucide:square-terminal"
        label="Open in editor"
      />
    </div>

    <p>
      {{ content.statement }}
    </p>

    <template
      v-if="!isNullOrUndefined(content.examples) && content.examples.length > 0"
    >
      <template
        v-for="(testCase, testCaseIdx) in content.examples"
        :key="testCaseIdx"
      >
        <h3 class="heading mt-4 mb-3">Example {{ testCaseIdx + 1 }}</h3>
        <DisplayCodeBlock>
          <p>Input: {{ testCase.input }}</p>
          <p>Output: {{ testCase.output }}</p>
          <p v-if="testCase.explanation">
            Explanation: {{ testCase.explanation }}
          </p>
        </DisplayCodeBlock>
      </template>
    </template>
  </article>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

defineProps<{
  content: QuestionCodingContent
  editorLink?: string
}>()
</script>
