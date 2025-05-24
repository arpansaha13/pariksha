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
        v-for="(example, exampleIdx) in content.examples"
        :key="exampleIdx"
      >
        <h3 class="heading mt-4 mb-3">Example {{ exampleIdx + 1 }}</h3>
        <div class="rounded-md bg-neutral-100 p-4 font-mono shadow-sm">
          <p>Input: {{ example.input }}</p>
          <p>Output: {{ example.output }}</p>
          <p v-if="example.explanation">
            Explanation: {{ example.explanation }}
          </p>
        </div>
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
