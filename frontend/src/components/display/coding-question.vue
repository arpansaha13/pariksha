<template>
  <article>
    <div class="mb-4 flex items-start justify-between gap-x-4">
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
      v-if="
        !isNullOrUndefined(content.test_cases) && content.test_cases.length > 0
      "
    >
      <template
        v-for="(testCase, testCaseIdx) in content.test_cases"
        :key="testCaseIdx"
      >
        <h3 class="heading mt-4 mb-3">Example {{ testCaseIdx + 1 }}</h3>
        <DisplayCodeBlock>
          <p>
            <span class="font-semibold">Input:</span>
            {{
              testCase.inputs
                .map(
                  (val, i) =>
                    `${content.input_definitions[i].variable_name} = ${val}`
                )
                .join(', ')
            }}
          </p>

          <p>
            <span class="font-semibold">Output:</span>
            {{ testCase.output }}
          </p>

          <p v-if="testCase.explanation">
            <span class="font-semibold">Explanation:</span>
            {{ testCase.explanation }}
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
