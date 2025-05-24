<template>
  <main class="h-screen w-screen px-4 py-2">
    <ClientOnly>
      <SplitterGroup
        id="splitter-group-1"
        direction="horizontal"
        auto-save-id="editor-splitter-group-1-save"
      >
        <SplitterPanel
          id="splitter-group-1-panel-1"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard
            :ui="{
              root: 'overflow-hidden h-full',
              body: 'overflow-y-auto h-full',
            }"
          >
            <DisplayCodingQuestion
              v-if="
                !isNullOrUndefined(questionData) &&
                questionData.type === QuestionType.CODING
              "
              :content="questionData.question"
            />
          </UCard>
        </SplitterPanel>

        <SplitterResizeHandle
          id="splitter-group-1-resize-handle-1"
          class="group w-2 outline-none"
        >
          <div
            class="group-hover:bg-primary-500 group-focus:bg-primary-500 mx-auto h-full w-0.5 transition-colors delay-150"
          />
        </SplitterResizeHandle>

        <SplitterPanel
          id="splitter-group-1-panel-2"
          :min-size="20"
          class="splitter-panel"
        >
          <UCard :ui="{ root: 'h-full', body: 'h-full p-0!' }">
            <MonacoEditor
              v-if="editorStore.isEditorPrepared"
              v-model="value"
              :lang="editorLang"
              class="h-full"
            />
          </UCard>
        </SplitterPanel>
      </SplitterGroup>
    </ClientOnly>
  </main>
</template>

<script lang="ts" setup>
import { isNullOrUndefined } from '@arpansaha13/utils'

definePageMeta({
  layout: 'blank',
})

const route = useRoute()
const questionId = parseInt(route.params.questionId as string) as QuestionId

const editorStore = useEditorStore()
const {
  data: questionData,
  error: questionError,
  status: questionStatus,
} = await usePaperQuestion(questionId)

await callOnce(
  () => {
    if (questionStatus.value === 'error' && questionError.value) {
      throw createError({
        statusCode: questionError.value.statusCode,
        message: 'You do not have access to this paper.',
      })
    }
    if (
      isNullOrUndefined(questionData.value) ||
      questionData.value.type !== QuestionType.CODING
    ) {
      throw createError({
        statusCode: HttpStatus.NOT_FOUND,
        message: 'We could not find the question you are looking for.',
      })
    }
  },
  { mode: 'navigation' }
)

onMounted(async () => {
  await editorStore.prepareEditor()
})

const editorLang = ref<EditorLanguage>('javascript')
const value = ref(`function getColor() {
  const a = ""
}`)
</script>

<style scoped>
@reference "~/assets/css/main.css";

.splitter-panel {
  @apply rounded-md border border-neutral-400;
}
</style>
